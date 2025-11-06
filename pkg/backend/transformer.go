// Package backend contains the core logic for the Catalyst datasource.
// This file provides data transformation functions to convert API responses
// into Grafana data frames.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	log "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// IssueResponseToDataFrame transforms a list of assurance issues from the API
// into a Grafana data.Frame. It enriches the data by resolving site IDs to names
// using the provided SiteTranslator.
func IssueResponseToDataFrame(
	ctx context.Context,
	allIssues []map[string]any,
	refID string,
	timeRange backend.TimeRange,
	hardLimit int64,
	siteTranslator *SiteTranslator,
) *data.Frame {
	// Internal row structure for organizing issue data
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
		site, err := siteTranslator.GetSiteByID(ctx, siteIDValue)
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
			r.TimeMs = timeRange.From.UnixMilli()
		}
		issueRows = append(issueRows, r)
	}

	// Build the Grafana data.Frame.
	frame := data.NewFrame(refID)
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

	return frame
}

// SiteHealthResponseToDataFrame transforms site health data from the API
// into a Grafana data.Frame. It dynamically creates fields based on the
// data structure and user's metric selection.
func SiteHealthResponseToDataFrame(
	allSiteHealthData []map[string]any,
	refID string,
	selectedMetrics []string,
) *data.Frame {
	// Handle empty data case
	if len(allSiteHealthData) == 0 {
		frame := data.NewFrame(refID)
		frame.SetMeta(&data.FrameMeta{
			Notices: []data.Notice{{Severity: data.NoticeSeverityInfo, Text: "No site health data found for the selected filters."}},
		})
		return frame
	}

	// Build metric selection map
	metricSet := make(map[string]struct{})
	if len(selectedMetrics) > 0 {
		for _, m := range selectedMetrics {
			metricSet[m] = struct{}{}
		}
	}

	// Dynamically create fields based on the first item and the user's metric selection
	firstItem := allSiteHealthData[0]

	frame := data.NewFrame(refID)
	// Always include site name and ID
	frame.Fields = append(frame.Fields, data.NewField("siteName", nil, make([]string, 0, len(allSiteHealthData))))
	frame.Fields = append(frame.Fields, data.NewField("siteId", nil, make([]string, 0, len(allSiteHealthData))))

	// Add fields for selected metrics
	for key := range firstItem {
		if _, isSelected := metricSet[key]; isSelected || len(metricSet) == 0 {
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

	return frame
}
