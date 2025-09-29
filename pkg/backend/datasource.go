// Package backend contains the core logic for the Catalyst datasource.
package backend

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	log "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Datasource is the main backend implementation for the Catalyst datasource.
// It handles all backend operations: querying data, checking health, and
// processing resource calls.
type Datasource struct {
	tm *tokenManager
}

// dsInstance represents a single configured instance of the datasource.
// It holds the parsed settings and the unique instance UID.
type dsInstance struct {
	Settings *InstanceSettings
	UID      string
}

// NewDatasource creates a new datasource instance with its own token manager.
func NewDatasource() *Datasource {
	return &Datasource{
		tm: newTokenManager(),
	}
}

// httpClientFor creates an HTTP client that respects the InsecureSkipVerify setting
// for the given datasource instance. This is crucial for environments with
// self-signed certificates.
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
//  2. Paginates through the Catalyst Center API to fetch all relevant issues,
//     respecting the user-defined limit.
//  3. Handles token acquisition and automatic refresh on 401/403 errors.
//  4. Optionally enriches the data by resolving site IDs to names if the `enrich` flag is set.
//  5. Transforms the API response into a Grafana data.Frame.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return nil, err
	}
	settings := inst.Settings
	httpClient := d.httpClientFor(settings)

	for _, q := range req.Queries {
		dr := backend.DataResponse{}

		// 1. Unmarshal the query model sent from the frontend.
		var qm QueryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			dr.Error = fmt.Errorf("invalid query model: %w", err)
			resp.Responses[q.RefID] = dr
			continue
		}
		if strings.TrimSpace(qm.QueryType) == "siteHealth" {
			siteHealthURL, err := SiteHealthURL(settings.BaseURL)
			if err != nil {
				dr.Error = err
				resp.Responses[q.RefID] = dr
				continue
			}
			pageSize := 25
			offset := 0
			params := buildSiteHealthParamsFromQuery(qm, pageSize, offset+1)
			token, err := d.tm.getToken(ctx, inst.UID, settings, httpClient)
			if err != nil {
				dr.Error = fmt.Errorf("token: %w", err)
				resp.Responses[q.RefID] = dr
				continue
			}
			reqURL := siteHealthURL + "?" + params.Encode()
			httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			httpReq.Header.Set("X-Auth-Token", token)
			httpResp, err := httpClient.Do(httpReq)
			if err != nil {
				dr.Error = fmt.Errorf("site-health request failed: %w", err)
				resp.Responses[q.RefID] = dr
				continue
			}
			body, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
			if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
				dr.Error = fmt.Errorf("site-health endpoint returned %s: %s", httpResp.Status, string(body))
				resp.Responses[q.RefID] = dr
				continue
			}
			var env struct {
				Response []map[string]any `json:"response"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				dr.Error = fmt.Errorf("site-health response: %w", err)
				resp.Responses[q.RefID] = dr
				continue
			}
			// Filter by parentSiteName and siteName if set
			filtered := env.Response
			if qm.ParentSiteName != "" || qm.SiteName != "" {
				filtered = make([]map[string]any, 0, len(env.Response))
				for _, site := range env.Response {
					if qm.ParentSiteName != "" && site["parentSiteName"] != qm.ParentSiteName {
						continue
					}
					if qm.SiteName != "" && site["siteName"] != qm.SiteName {
						continue
					}
					filtered = append(filtered, site)
				}
			}
			// Build Grafana frames for selected metrics
			frame := data.NewFrame(q.RefID)
			for _, metric := range qm.Metrics {
				f := data.NewField(metric, nil, make([]int64, 0, len(filtered)))
				for _, site := range filtered {
					if v, ok := site[metric]; ok {
						switch x := v.(type) {
						case float64:
							f.Append(int64(x))
						case int64:
							f.Append(x)
						case json.Number:
							n, _ := x.Int64()
							f.Append(n)
						}
					} else {
						f.Append(0)
					}
				}
				frame.Fields = append(frame.Fields, f)
			}
			dr.Frames = append(dr.Frames, frame)
			resp.Responses[q.RefID] = dr
			continue
		}

		issuesURL, err := IssuesURL(settings.BaseURL)
		if err != nil {
			dr.Error = err
			resp.Responses[q.RefID] = dr
			continue
		}

		// 2. Set up pagination. We'll loop until we either hit the hard limit
		//    or the API returns fewer results than the page size.
		pageSize := 25
		offset := 0
		var hardLimit int64 = 25
		if qm.Limit != nil && *qm.Limit > 0 {
			hardLimit = *qm.Limit
		}

		type row struct {
			TimeMs   int64
			ID       string
			Title    string
			Severity string
			Status   string
			Category string
			Device   string
			MAC      string
			Site     string
			Rule     string
			Details  string
		}
		issueRows := make([]row, 0, 256)
		allIssues := make([]map[string]any, 0, 256)

		for int64(len(allIssues)) < hardLimit {
			limitForThisPage := pageSize
			remaining := int(hardLimit - int64(len(allIssues)))
			if remaining < limitForThisPage {
				limitForThisPage = remaining
			}

			params := buildAssuranceParamsFromQuery(
				qm,
				q.TimeRange.From.UnixMilli(),
				q.TimeRange.To.UnixMilli(),
				limitForThisPage,
				offset+1,
			)

			// 3. Get a valid token, either from cache or by fetching a new one.
			token, err := d.tm.getToken(ctx, inst.UID, settings, httpClient)
			if err != nil {
				dr.Error = fmt.Errorf("token: %w", err)
				break
			}

			reqURL := issuesURL + "?" + params.Encode()
			httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			httpReq.Header.Set("X-Auth-Token", token)

			httpResp, err := httpClient.Do(httpReq)
			if err != nil {
				dr.Error = fmt.Errorf("issues request failed: %w", err)
				break
			}
			body, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()

			// If the token has expired, the API will return 401 or 403.
			// In this case, we force a token refresh and retry the request once.
			if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden {
				log.DefaultLogger.Warn("Unauthorized; refreshing token and retrying")
				d.tm.set(inst.UID, "") // Force refresh by clearing the cached token.
				token, err = d.tm.getToken(ctx, inst.UID, settings, httpClient)
				if err != nil {
					dr.Error = fmt.Errorf("token refresh: %w", err)
					break
				}
				httpReq, _ = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
				httpReq.Header.Set("X-Auth-Token", token)
				httpResp, err = httpClient.Do(httpReq)
				if err != nil {
					dr.Error = fmt.Errorf("issues request retry failed: %w", err)
					break
				}
				body, _ = io.ReadAll(httpResp.Body)
				httpResp.Body.Close()
			}

			if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
				dr.Error = fmt.Errorf("issues endpoint returned %s: %s", httpResp.Status, string(body))
				break
			}

			var env IssuesEnvelope
			var arr []map[string]any
			if err := json.Unmarshal(body, &env); err == nil && len(env.Response) > 0 {
				arr = env.Response
			} else {
				// Some API versions might return a raw array instead of an envelope.
				_ = json.Unmarshal(body, &arr)
			}
			if len(arr) == 0 {
				// No more results, exit the pagination loop.
				break
			}

			allIssues = append(allIssues, arr...)
			if len(arr) < pageSize {
				// The API returned fewer items than we asked for, so this is the last page.
				break
			}
			offset += pageSize
		}

		// 4. Site Name Enrichment: If the 'enrich' flag is set, resolve site IDs to names.
		// This is done after collecting all issues to batch the site ID lookups into one API call.
		siteIDToNameMap := make(map[string]string)
		if qm.Enrich && len(allIssues) > 0 {
			uniqueSiteIDs := make(map[string]struct{})
			for _, issue := range allIssues {
				if siteID, ok := issue["siteId"].(string); ok && siteID != "" {
					uniqueSiteIDs[siteID] = struct{}{}
				}
			}

			var siteIDs []string
			for id := range uniqueSiteIDs {
				siteIDs = append(siteIDs, id)
			}

			if len(siteIDs) > 0 {
				var err error
				siteIDToNameMap, err = d.getSiteNamesByID(ctx, httpClient, inst, siteIDs)
				if err != nil {
					log.DefaultLogger.Warn("failed to resolve site names", "err", err)
				}
			}
		}

		// 5. Data Transformation: Convert the raw API response into a structured format
		//    that can be used to build the Grafana data.Frame.
		for _, it := range allIssues {
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

			siteID := getStr("siteId")
			siteName := siteID // Fallback to ID if enrichment is disabled or fails.
			if name, ok := siteIDToNameMap[siteID]; ok {
				siteName = name // Use resolved name if available.
			}

			r := row{
				TimeMs:   firstNonZero(getNum("timestamp"), getNum("firstOccurredTime"), getNum("startTime")),
				ID:       firstNonEmpty(getStr("issueId"), getStr("id"), getStr("instanceId")),
				Title:    firstNonEmpty(getStr("name"), getStr("title"), getStr("issueTitle")),
				Severity: firstNonEmpty(getStr("priority"), getStr("severity")),
				Status:   firstNonEmpty(getStr("issueStatus"), getStr("status")),
				Category: firstNonEmpty(getStr("category"), getStr("type")),
				Device:   firstNonEmpty(getStr("deviceId"), getStr("deviceIp"), getStr("device")),
				MAC:      firstNonEmpty(getStr("macAddress"), getStr("clientMac")),
				Site:     siteName,
				Rule:     getStr("ruleId"),
				Details:  firstNonEmpty(getStr("description"), getStr("details"), getStr("issueDescription")),
			}
			if r.TimeMs == 0 {
				r.TimeMs = q.TimeRange.From.UnixMilli()
			}
			issueRows = append(issueRows, r)
		}

		// 6. Build the Grafana data.Frame, which is the final structure that gets
		//    sent back to the frontend for rendering.
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
			fRule.Append(r.Rule)
			fDetails.Append(r.Details)
		}

		frame.Fields = append(frame.Fields,
			fTime, fID, fTitle, fSeverity, fStatus, fCategory, fDevice, fMAC, fSite, fRule, fDetails,
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

// getSiteNamesByID performs a batch lookup to resolve a list of site IDs to their
// corresponding site names. This is more efficient than making one request per site.
func (d *Datasource) getSiteNamesByID(ctx context.Context, httpClient *http.Client, inst *dsInstance, siteIDs []string) (map[string]string, error) {
	siteURL, err := SiteURL(inst.Settings.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("bad site baseUrl: %w", err)
	}

	// The Catalyst Center API supports fetching multiple sites by providing a comma-separated list of IDs.
	params := url.Values{}
	params.Set("siteId", strings.Join(siteIDs, ","))
	reqURL := siteURL + "?" + params.Encode()

	token, err := d.tm.getToken(ctx, inst.UID, inst.Settings, httpClient)
	if err != nil {
		return nil, fmt.Errorf("token for site lookup: %w", err)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	httpReq.Header.Set("X-Auth-Token", token)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("site request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("site endpoint returned %s: %s", httpResp.Status, string(body))
	}

	var envelope SiteEnvelope
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode site response: %w", err)
	}

	nameMap := make(map[string]string)
	for _, site := range envelope.Response {
		if site.ID != "" && site.Name != "" {
			nameMap[site.ID] = site.Name
		}
	}
	return nameMap, nil
}

// ---- CheckHealth ----

// CheckHealth is called by Grafana to verify that the datasource is configured
// correctly and can connect to the Catalyst Center API. It performs a lightweight
// check by attempting to fetch a token and then making a simple API call.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "instance error: " + err.Error(),
		}, nil
	}
	settings := inst.Settings
	httpClient := d.httpClientFor(settings)

	// 1. Verify that we can obtain an authentication token.
	if _, err := d.tm.getToken(ctx, inst.UID, settings, httpClient); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "token: " + err.Error(),
		}, nil
	}

	issuesURL, err := IssuesURL(settings.BaseURL)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "invalid base URL",
		}, nil
	}

	// 2. Make a lightweight test query to the issues endpoint.
	u := issuesURL + "?limit=1"
	reqHTTP, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	tok, _ := d.tm.getToken(ctx, inst.UID, settings, httpClient)
	reqHTTP.Header.Set("X-Auth-Token", tok)

	httpResp, err := httpClient.Do(reqHTTP)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "issues probe failed: " + err.Error(),
		}, nil
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: "Successfully connected to Catalyst Center (issues)",
		}, nil
	}
	b, _ := io.ReadAll(httpResp.Body)
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: fmt.Sprintf("issues probe %s: %s", httpResp.Status, string(b)),
	}, nil
}

// ---- CallResource passthrough (honors TLS flag as well) ----

// CallResource handles custom API requests from the frontend, typically used for
// things like fetching values for template variables. This acts as a secure
// proxy to the Catalyst Center API.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	inst, err := getInstanceFromPluginContext(req.PluginContext)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("instance error: " + err.Error()),
		})
	}
	settings := inst.Settings
	httpClient := d.httpClientFor(settings)

	switch req.Path {
	case "issues":
		// The 'issues' resource path is used by the frontend to populate template variables.
		return d.resourceIssues(ctx, inst, req, sender, httpClient)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   []byte("not found"),
		})
	}
}

// resourceIssues handles requests to the /issues resource path. It forwards the
// query parameters from the frontend to the Catalyst Center issues API.
func (d *Datasource) resourceIssues(ctx context.Context, inst *dsInstance, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender, httpClient *http.Client) error {
	issuesURL, err := IssuesURL(inst.Settings.BaseURL)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusBadRequest, Body: []byte("bad baseUrl")})
	}

	var rawQuery string
	if req.URL != "" {
		if u, err := url.Parse(req.URL); err == nil {
			rawQuery = u.RawQuery
		}
	}

	q := ""
	if rawQuery != "" {
		q = "?" + rawQuery
	}

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, issuesURL+q, nil)
	tok, err := d.tm.getToken(ctx, inst.UID, inst.Settings, httpClient)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusUnauthorized, Body: []byte("token: " + err.Error())})
	}
	httpReq.Header.Set("X-Auth-Token", tok)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusBadGateway, Body: []byte("request failed: " + err.Error())})
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)

	return sender.Send(&backend.CallResourceResponse{
		Status:  httpResp.StatusCode,
		Body:    body,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
	})
}

// ---- helpers ----

// firstNonEmpty returns the first non-empty string from a list of arguments.
// This is useful for coalescing values from multiple possible API fields.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// firstNonZero returns the first non-zero int64 from a list of arguments.
// Useful for finding a valid timestamp from multiple potential fields.
func firstNonZero(vals ...int64) int64 {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
