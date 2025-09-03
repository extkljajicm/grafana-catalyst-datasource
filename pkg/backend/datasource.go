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

type Datasource struct {
	tm *tokenManager
}

type dsInstance struct {
	Settings *InstanceSettings
	UID      string
}

func NewDatasource() *Datasource {
	return &Datasource{
		tm: newTokenManager(),
	}
}

// Build an HTTP client honoring InsecureSkipVerify for this instance.
func (d *Datasource) httpClientFor(s *InstanceSettings) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: s.InsecureSkipVerify}, //nolint:gosec
	}
	return &http.Client{Timeout: 30 * time.Second, Transport: tr}
}

// ---- helpers to read instance settings directly from PluginContext ----

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

		var qm QueryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			dr.Error = fmt.Errorf("invalid query model: %w", err)
			resp.Responses[q.RefID] = dr
			continue
		}
		if strings.TrimSpace(qm.QueryType) != "alerts" {
			dr.Frames = append(dr.Frames, data.NewFrame(q.RefID))
			resp.Responses[q.RefID] = dr
			continue
		}

		issuesURL, err := IssuesURL(settings.BaseURL)
		if err != nil {
			dr.Error = err
			resp.Responses[q.RefID] = dr
			continue
		}

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
		rows := make([]row, 0, 256)

		for int64(len(rows)) < hardLimit {
			// Calculate the page size for this specific request to not exceed hardLimit
			limitForThisPage := pageSize
			remaining := int(hardLimit - int64(len(rows)))
			if remaining < limitForThisPage {
				limitForThisPage = remaining
			}

			// Use the helper function to build params, ensuring a 1-based offset
			params := buildAssuranceParamsFromQuery(
				qm,
				q.TimeRange.From.UnixMilli(),
				q.TimeRange.To.UnixMilli(),
				limitForThisPage,
				offset+1, // The assurance API uses a 1-based offset
			)

			// Ensure token (using the same TLS behavior)
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

			if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden {
				log.DefaultLogger.Warn("Unauthorized; refreshing token and retrying")
				d.tm.set(inst.UID, "") // force refresh
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
				_ = json.Unmarshal(body, &arr)
			}
			if len(arr) == 0 {
				break
			}

			for _, it := range arr {
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

				r := row{
					TimeMs:   firstNonZero(getNum("timestamp"), getNum("firstOccurredTime"), getNum("startTime")),
					ID:       firstNonEmpty(getStr("issueId"), getStr("id"), getStr("instanceId")),
					Title:    firstNonEmpty(getStr("name"), getStr("title"), getStr("issueTitle")),
					Severity: firstNonEmpty(getStr("priority"), getStr("severity")),
					Status:   firstNonEmpty(getStr("issueStatus"), getStr("status")),
					Category: firstNonEmpty(getStr("category"), getStr("type")),
					Device:   firstNonEmpty(getStr("deviceId"), getStr("deviceIp"), getStr("device")),
					MAC:      firstNonEmpty(getStr("macAddress"), getStr("clientMac")),
					Site:     firstNonEmpty(getStr("siteId"), getStr("site")),
					Rule:     getStr("ruleId"),
					Details:  firstNonEmpty(getStr("description"), getStr("details"), getStr("issueDescription")),
				}
				if r.TimeMs == 0 {
					r.TimeMs = q.TimeRange.From.UnixMilli()
				}
				rows = append(rows, r)
				if int64(len(rows)) >= hardLimit {
					break
				}
			}

			if len(arr) < pageSize {
				break
			}
			offset += pageSize
		}

		// Build frame
		frame := data.NewFrame(q.RefID)
		fTime := data.NewField("Time", nil, make([]time.Time, 0, len(rows)))
		fID := data.NewField("Issue ID", nil, make([]string, 0, len(rows)))
		fTitle := data.NewField("Title", nil, make([]string, 0, len(rows)))
		fSeverity := data.NewField("Priority", nil, make([]string, 0, len(rows)))
		fStatus := data.NewField("Status", nil, make([]string, 0, len(rows)))
		fCategory := data.NewField("Category", nil, make([]string, 0, len(rows)))
		fDevice := data.NewField("Device ID", nil, make([]string, 0, len(rows)))
		fMAC := data.NewField("MAC", nil, make([]string, 0, len(rows)))
		fSite := data.NewField("Site ID", nil, make([]string, 0, len(rows)))
		fRule := data.NewField("Rule", nil, make([]string, 0, len(rows)))
		fDetails := data.NewField("Details", nil, make([]string, 0, len(rows)))

		for _, r := range rows {
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

		// If no rows, add an informational notice for better UX
		if len(rows) == 0 {
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

// ---- CheckHealth ----

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
		return d.resourceIssues(ctx, inst, req, sender, httpClient)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   []byte("not found"),
		})
	}
}

func (d *Datasource) resourceIssues(ctx context.Context, inst *dsInstance, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender, httpClient *http.Client) error {
	issuesURL, err := IssuesURL(inst.Settings.BaseURL)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusBadRequest, Body: []byte("bad baseUrl")})
	}

	// req.URL is a string; parse to extract query string
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

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func firstNonZero(vals ...int64) int64 {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}