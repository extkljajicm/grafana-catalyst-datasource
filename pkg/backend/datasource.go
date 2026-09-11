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

		// Log warning if many sites are selected (may impact performance)
		if len(qm.SiteID) > 5 {
			log.DefaultLogger.Warn("Large number of sites selected may increase query time", "count", len(qm.SiteID))
		}

		// 3. Use the Client to fetch all issues (Client handles pagination and retries).
		// Note: qm.Enrich field is reserved for future use (e.g., full device details).
		// Site name enrichment always runs as it is efficient (see lines 228-241 in transformer.go).
		client := NewClient(httpClient, d.tm, inst)
		allIssues, err := client.FetchAllIssues(ctx, issuesURL, qm, hardLimit)
		if err != nil {
			dr.Error = err
			resp.Responses[q.RefID] = dr
			continue
		}

		// 4. Transform API response into Grafana data.Frame
		frame := IssueResponseToDataFrame(ctx, allIssues, q.RefID, q.TimeRange, hardLimit, siteTranslator)
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

	// Transform API response into Grafana data.Frame
	frame := SiteHealthResponseToDataFrame(allSiteHealthData, refID, qm.Metrics)
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
	switch req.Path {
	case "sites":
		return d.handleSitesRequest(ctx, req, sender)
	case "issues":
		inst, err := getInstanceFromPluginContext(req.PluginContext)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte("failed to get instance settings: " + err.Error()),
			})
		}
		httpClient := d.httpClientFor(inst.Settings)
		return d.resourceIssues(ctx, inst, req, sender, httpClient)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   []byte("not found"),
		})
	}
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
		Status:  http.StatusOK,
		Body:    body,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
	})
}

// resourceIssues handles requests to the /issues resource path. It forwards the
// query parameters from the frontend to the Catalyst Center issues API.
func (d *Datasource) resourceIssues(ctx context.Context, inst *dsInstance, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender, httpClient *http.Client) error {
	issuesURL, err := IssuesURL(inst.Settings.BaseURL)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   []byte("bad baseUrl: " + err.Error()),
		})
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, issuesURL+q, nil)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to create request: " + err.Error()),
		})
	}

	tok, err := d.tm.getToken(ctx, inst.UID, inst.Settings, httpClient)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusUnauthorized,
			Body:   []byte("token: " + err.Error()),
		})
	}
	httpReq.Header.Set("X-Auth-Token", tok)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte("request failed: " + err.Error()),
		})
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to read response: " + err.Error()),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status:  httpResp.StatusCode,
		Body:    body,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
	})
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
