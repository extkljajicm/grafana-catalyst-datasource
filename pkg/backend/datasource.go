// Package backend contains the core logic for the Catalyst datasource.
package backend

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	log "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Datasource is the main backend implementation for the Catalyst datasource.
// It holds shared resources like the token manager and site translators.
type Datasource struct {
	tm *tokenManager
	// A map of site translators, one for each datasource instance UID.
	translators   map[string]*SiteTranslator
	translatorMux sync.Mutex
}

// dsInstance represents a single configured instance of the datasource.
// It holds the parsed settings and the unique instance UID.
type dsInstance struct {
	Settings *InstanceSettings
	UID      string
}

// NewDatasource creates a new datasource instance.
func NewDatasource() *Datasource {
	return &Datasource{
		tm:          newTokenManager(),
		translators: make(map[string]*SiteTranslator),
	}
}

// getSiteTranslator returns a site translator for the given instance.
// If one doesn't exist, it creates and caches it.
func (d *Datasource) getSiteTranslator(inst *dsInstance, httpClient *http.Client) *SiteTranslator {
	d.translatorMux.Lock()
	defer d.translatorMux.Unlock()

	if t, ok := d.translators[inst.UID]; ok {
		return t
	}
	t := NewSiteTranslator(httpClient, inst, d.tm)
	d.translators[inst.UID] = t
	log.DefaultLogger.Info("Created new site translator for instance", "uid", inst.UID)
	return t
}

// httpClientFor creates an HTTP client that respects the InsecureSkipVerify setting.
func (d *Datasource) httpClientFor(s *InstanceSettings) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: s.InsecureSkipVerify}, //nolint:gosec
	}
	return &http.Client{Timeout: 30 * time.Second, Transport: tr}
}

// ---- helpers to read instance settings directly from PluginContext ----

// getInstanceFromPluginContext retrieves and parses the settings for the current
// datasource instance from the plugin context provided by Grafana.
func getInstanceFromPluginContext(pc backend.PluginContext) (*dsInstance, error) {
	ds := pc.DataSourceInstanceSettings
	if ds == nil {
		return nil, fmt.Errorf("no datasource instance settings in plugin context")
	}
	cfg, err := ParseInstanceSettings(ds.JSONData, ds.DecryptedSecureJSONData)
	if err != nil {
		return nil, err
	}
	return &dsInstance{
		Settings: cfg,
		UID:      ds.UID,
	}, nil
}

// ---- QueryData ----

// QueryData is the primary method for handling data queries from Grafana panels.
// It executes the following steps:
//  1. Parses the query from the frontend.
//  2. Uses the Client to fetch all relevant issues from the API (Client handles pagination and retries).
//  3. Enriches the data by resolving site IDs to names using the SiteTranslator.
//  4. Transforms the API response into a Grafana data.Frame.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return nil, err
	}
	httpClient := d.httpClientFor(inst.Settings)
	siteTranslator := d.getSiteTranslator(inst, httpClient)

	// Ensure the site cache is warm before processing queries. This prevents
	// repeated cache checks inside the loop.
	if err := siteTranslator.EnsureCache(ctx); err != nil {
		// Log the error but don't fail the whole query; it might still succeed
		// if the cache is partially available or if no site enrichment is needed.
		log.DefaultLogger.Error("Failed to ensure site cache", "err", err)
	}

	for _, q := range req.Queries {
		dr := backend.DataResponse{}

		// 1. Unmarshal the query model sent from the frontend.
		var qm QueryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			dr.Error = fmt.Errorf("invalid query model: %w", err)
			resp.Responses[q.RefID] = dr
			log.DefaultLogger.Error("Failed to unmarshal query model", "err", err, "json", string(q.JSON))
			continue
		}
		qm.TimeRange = q.TimeRange
		log.DefaultLogger.Info("Executing query", "refId", q.RefID, "queryType", qm.QueryType, "limit", qm.Limit)
		log.DefaultLogger.Debug("Full query model", "query", fmt.Sprintf("%+v", qm))

		// If site IDs are not provided, but site names are, resolve them.
		// This handles cases where the user might manually enter a site name.
		if len(qm.SiteID) == 0 && len(qm.SiteName) > 0 {
			log.DefaultLogger.Info("Resolving site names to IDs", "names", qm.SiteName)
			var resolvedIDs []string
			for _, name := range qm.SiteName {
				siteID, err := siteTranslator.GetSiteID(ctx, name)
				if err != nil {
					// Log the error but continue, so one bad name doesn't fail the whole query.
					log.DefaultLogger.Warn("failed to resolve site name", "name", name, "err", err)
					continue
				}
				resolvedIDs = append(resolvedIDs, siteID)
			}
			qm.SiteID = resolvedIDs
			log.DefaultLogger.Info("Resolved site IDs", "ids", resolvedIDs)
		}

		if strings.TrimSpace(qm.QueryType) == "siteHealth" {
			log.DefaultLogger.Info("Routing to siteHealth handler")
			frame, err := d.querySiteHealth(ctx, inst, qm, q.RefID, httpClient)
			if err != nil {
				dr.Error = err
			} else {
				dr.Frames = append(dr.Frames, frame)
			}
			resp.Responses[q.RefID] = dr
			continue
		}

		log.DefaultLogger.Info("Routing to assuranceIssues handler")
		issuesURL, err := IssuesURL(inst.Settings.BaseURL)
		if err != nil {
			dr.Error = err
			resp.Responses[q.RefID] = dr
			continue
		}

		// 2. Determine the hard limit for this query.
		var hardLimit int64 = 1000 // A higher default hard limit
		if qm.Limit != nil && *qm.Limit > 0 {
			hardLimit = *qm.Limit
		}
		log.DefaultLogger.Info("Pagination configured", "hardLimit", hardLimit)

		// 3. Use the Client to fetch all issues (Client handles pagination and retries).
		client := NewClient(httpClient, d.tm, inst)
		allIssues, err := client.FetchAllIssues(ctx, issuesURL, qm, hardLimit)
		if err != nil {
			dr.Error = err
			resp.Responses[q.RefID] = dr
			continue
		}

		// 4. Data Transformation and Enrichment.
		type row struct {
			TimeMs     int64
			ID         string
			Title      string
			Severity   string
			Status     string
			Category   string
			Device     string
			MAC        string
			Site       string
			ParentSite string
			Rule       string
			Details    string
		}

		issueRows := make([]row, 0, min(len(allIssues), int(hardLimit)))
		for _, it := range allIssues[:min(len(allIssues), int(hardLimit))] {
			getStr := func(k string) string {
				if v, ok := it[k]; ok && v != nil {
					if s, ok2 := v.(string); ok2 {
						return s
					}
				}
				return ""
			}
			getNum := func(k string) int64 {
				if v, ok := it[k]; ok && v != nil {
					switch x := v.(type) {
					case float64:
						return int64(x)
					case int64:
						return x
					case json.Number:
						n, _ := x.Int64()
						return n
					}
				}
				return 0
			}

			siteIDValue := getStr("siteId")
			siteNameValue := siteIDValue // Fallback to ID.
			parentSiteNameValue := ""      // Default to empty string
			var site Site
			var parentSite Site
			// Use the translator to get the site and its parent.
			site, err = siteTranslator.GetSiteByID(ctx, siteIDValue)
			if err == nil {
				siteNameValue = site.Name
				if site.ParentID != "" {
					parentSite, err = siteTranslator.GetSiteByID(ctx, site.ParentID)
					if err == nil {
						parentSiteNameValue = parentSite.Name
					} else {
						log.DefaultLogger.Warn("failed to resolve parent site ID to name", "parentSiteId", site.ParentID, "err", err)
					}
				}
			} else {
				log.DefaultLogger.Warn("failed to resolve site ID to name", "siteId", siteIDValue, "err", err)
			}

			r := row{
				TimeMs:     firstNonZero(getNum("timestamp"), getNum("firstOccurredTime"), getNum("startTime")),
				ID:         firstNonEmpty(getStr("issueId"), getStr("id"), getStr("instanceId")),
				Title:      firstNonEmpty(getStr("name"), getStr("title"), getStr("issueTitle")),
				Severity:   firstNonEmpty(getStr("priority"), getStr("severity")),
				Status:     firstNonEmpty(getStr("issueStatus"), getStr("status")),
				Category:   firstNonEmpty(getStr("category"), getStr("type")),
				Device:     firstNonEmpty(getStr("networkDeviceId"), getStr("deviceId"), getStr("deviceIp"), getStr("device")),
				MAC:        firstNonEmpty(getStr("macAddress"), getStr("clientMac")),
				Site:       siteNameValue,
				ParentSite: parentSiteNameValue,
				Rule:       getStr("ruleId"),
				Details:    firstNonEmpty(getStr("description"), getStr("details"), getStr("issueDescription")),
			}
			if r.TimeMs == 0 {
				r.TimeMs = q.TimeRange.From.UnixMilli()
			}
			issueRows = append(issueRows, r)
		}

		// 5. Build the Grafana data.Frame.
		frame := data.NewFrame(q.RefID)
		fTime := data.NewField("Time", nil, make([]time.Time, 0, len(issueRows)))
		fID := data.NewField("Issue ID", nil, make([]string, 0, len(issueRows)))
		fTitle := data.NewField("Title", nil, make([]string, 0, len(issueRows)))
		fSeverity := data.NewField("Priority", nil, make([]string, 0, len(issueRows)))
		fStatus := data.NewField("Status", nil, make([]string, 0, len(issueRows)))
		fCategory := data.NewField("Category", nil, make([]string, 0, len(issueRows)))
		fDevice := data.NewField("Device ID", nil, make([]string, 0, len(issueRows)))
		fMAC := data.NewField("MAC", nil, make([]string, 0, len(issueRows)))
		fSite := data.NewField("Site Name", nil, make([]string, 0, len(issueRows)))
		fParentSite := data.NewField("Parent Site", nil, make([]string, 0, len(issueRows)))
		fRule := data.NewField("Rule", nil, make([]string, 0, len(issueRows)))
		fDetails := data.NewField("Details", nil, make([]string, 0, len(issueRows)))

		for _, r := range issueRows {
			fTime.Append(time.UnixMilli(r.TimeMs))
			fID.Append(r.ID)
			fTitle.Append(r.Title)
			fSeverity.Append(r.Severity)
			fStatus.Append(r.Status)
			fCategory.Append(r.Category)
			fDevice.Append(r.Device)
			fMAC.Append(r.MAC)
			fSite.Append(r.Site)
			fParentSite.Append(r.ParentSite)
			fRule.Append(r.Rule)
			fDetails.Append(r.Details)
		}

		frame.Fields = append(frame.Fields,
			fTime, fID, fTitle, fSeverity, fStatus, fCategory, fDevice, fMAC, fSite, fParentSite, fRule, fDetails,
		)

		if len(issueRows) == 0 {
			frame.SetMeta(&data.FrameMeta{
				Notices: []data.Notice{
					{
						Severity: data.NoticeSeverityInfo,
						Text:     "No issues found for the selected time range/filters",
					},
				},
			})
		}
		dr.Frames = append(dr.Frames, frame)
		resp.Responses[q.RefID] = dr
	}
	return resp, nil
}

// querySiteHealth handles the specific logic for the 'siteHealth' query type.
func (d *Datasource) querySiteHealth(ctx context.Context, inst *dsInstance, qm QueryModel, refID string, httpClient *http.Client) (*data.Frame, error) {
	siteHealthURL, err := url.Parse(inst.Settings.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	prefix := dnacPrefix(siteHealthURL.Path)
	siteHealthURL.Path = prefix + "/dna/intent/api/v1/site-health"

	// Use the Client to fetch all site health data
	client := NewClient(httpClient, d.tm, inst)
	allSiteHealthData, err := client.FetchAllSiteHealth(ctx, siteHealthURL, qm)
	if err != nil {
		return nil, err
	}

	// If user selected specific sites, filter the results now.
	if len(qm.SiteID) > 0 {
		siteIDSet := make(map[string]struct{})
		for _, id := range qm.SiteID {
			siteIDSet[id] = struct{}{}
		}

		var filteredData []map[string]any
		for _, siteData := range allSiteHealthData {
			if id, ok := siteData["siteId"].(string); ok {
				if _, exists := siteIDSet[id]; exists {
					filteredData = append(filteredData, siteData)
				}
			}
		}
		allSiteHealthData = filteredData
	}

	// Create the data frame from the (potentially filtered) data.
	if len(allSiteHealthData) == 0 {
		frame := data.NewFrame(refID)
		frame.SetMeta(&data.FrameMeta{
			Notices: []data.Notice{{Severity: data.NoticeSeverityInfo, Text: "No site health data found for the selected filters."}},
		})
		return frame, nil
	}

	// Dynamically create fields based on the first item and the user's metric selection
	firstItem := allSiteHealthData[0]
	selectedMetrics := make(map[string]struct{})
	if len(qm.Metrics) > 0 {
		for _, m := range qm.Metrics {
			selectedMetrics[m] = struct{}{}
		}
	}

	frame := data.NewFrame(refID)
	// Always include site name and ID
	frame.Fields = append(frame.Fields, data.NewField("siteName", nil, make([]string, 0, len(allSiteHealthData))))
	frame.Fields = append(frame.Fields, data.NewField("siteId", nil, make([]string, 0, len(allSiteHealthData))))

	// Add fields for selected metrics
	for key := range firstItem {
		if _, isSelected := selectedMetrics[key]; isSelected || len(selectedMetrics) == 0 {
			// Determine field type
			switch firstItem[key].(type) {
			case string:
				frame.Fields = append(frame.Fields, data.NewField(key, nil, make([]*string, 0, len(allSiteHealthData))))
			case float64:
				frame.Fields = append(frame.Fields, data.NewField(key, nil, make([]*float64, 0, len(allSiteHealthData))))
			case bool:
				frame.Fields = append(frame.Fields, data.NewField(key, nil, make([]*bool, 0, len(allSiteHealthData))))
			default:
				// Fallback to string for other types
				frame.Fields = append(frame.Fields, data.NewField(key, nil, make([]*string, 0, len(allSiteHealthData))))
			}
		}
	}

	// Populate the frame
	for _, item := range allSiteHealthData {
		for i, field := range frame.Fields {
			value, exists := item[field.Name]
			if !exists {
				// Append nil if the key doesn't exist for this item
				field.Append(nil)
				continue
			}

			// Use type assertion to append the correct type
			switch v := value.(type) {
			case string:
				if i < len(frame.Fields) && frame.Fields[i].Type() == data.FieldTypeNullableString {
					frame.Fields[i].Append(&v)
				}
			case float64:
				if i < len(frame.Fields) && frame.Fields[i].Type() == data.FieldTypeNullableFloat64 {
					frame.Fields[i].Append(&v)
				}
			case bool:
				if i < len(frame.Fields) && frame.Fields[i].Type() == data.FieldTypeNullableBool {
					frame.Fields[i].Append(&v)
				}
			default:
				// Fallback for other types
				if i < len(frame.Fields) && frame.Fields[i].Type() == data.FieldTypeNullableString {
					strVal := fmt.Sprintf("%v", v)
					frame.Fields[i].Append(&strVal)
				}
			}
		}
	}

	return frame, nil
}

// firstNonEmpty returns the first non-empty string from a list of arguments.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstNonZero returns the first non-zero int64 from a list of arguments.
func firstNonZero(vals ...int64) int64 {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}

// ---- CheckHealth ----

// CheckHealth is called by Grafana to verify that the datasource is configured correctly.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "instance error: " + err.Error(),
		}, nil
	}
	httpClient := d.httpClientFor(inst.Settings)

	// 1. Verify that we can obtain an authentication token.
	if _, err := d.tm.getToken(ctx, inst.UID, inst.Settings, httpClient); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "token error: " + err.Error(),
		}, nil
	}

	// 2. Perform a lightweight API call to the site translator to ensure connectivity.
	siteTranslator := d.getSiteTranslator(inst, httpClient)
	if _, err := siteTranslator.GetAllSites(ctx); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Health check failed: could not fetch sites. " + err.Error(),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Successfully connected to Catalyst Center and fetched sites.",
	}, nil
}

// ---- CallResource ----

// CallResource handles custom API requests from the frontend.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if req.Path == "sites" {
		return d.handleSitesRequest(ctx, req, sender)
	}
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusNotFound,
	})
}

// handleSitesRequest fetches all sites using the translator and returns them as JSON.
func (d *Datasource) handleSitesRequest(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to get instance settings: " + err.Error()),
		})
	}
	httpClient := d.httpClientFor(inst.Settings)
	siteTranslator := d.getSiteTranslator(inst, httpClient)

	sites, err := siteTranslator.GetAllSites(ctx)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to fetch sites: " + err.Error()),
		})
	}

	body, err := json.Marshal(sites)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to marshal sites to JSON: " + err.Error()),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   body,
	})
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
