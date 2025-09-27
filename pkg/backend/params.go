// Package backend contains the core logic for the Catalyst datasource.
// This file, params.go, is responsible for converting the frontend query model
// into the URL query parameters expected by the Catalyst Center API. It handles
// normalization, validation, and formatting of filter values.
package backend

import (
	"net/url"
	"strconv"
	"strings"
)

// Allowed value sets for validation and normalization.
var (
	// allowedPriority defines the valid priority values for the API.
	allowedPriority = map[string]struct{}{"P1": {}, "P2": {}, "P3": {}, "P4": {}}
	// allowedIssueStatus defines the valid status values for the API.
	allowedIssueStatus = map[string]struct{}{"ACTIVE": {}, "RESOLVED": {}, "IGNORED": {}}
)

// normalizePriority returns a valid priority string (P1-P4) if the input
// matches a known value. It checks both 'priority' and the legacy 'severity' fields.
func normalizePriority(priority, severity string) (string, bool) {
	p := strings.ToUpper(strings.TrimSpace(priority))
	if _, ok := allowedPriority[p]; ok {
		return p, true
	}
	s := strings.ToUpper(strings.TrimSpace(severity))
	if _, ok := allowedPriority[s]; ok {
		return s, true
	}
	return "", false
}

// normalizeIssueStatus returns a valid status string if the input matches a known
// value. It checks both 'issueStatus' and the legacy 'status' fields.
func normalizeIssueStatus(issueStatus, status string) (string, bool) {
	is := strings.ToUpper(strings.TrimSpace(issueStatus))
	if _, ok := allowedIssueStatus[is]; ok {
		return is, true
	}
	s := strings.ToUpper(strings.TrimSpace(status))
	if _, ok := allowedIssueStatus[s]; ok {
		return s, true
	}
	return "", false
}

// normalizeBoolish converts various string representations of a boolean
// (e.g., "true", "yes", "1") into a canonical "true" or "false" string.
func normalizeBoolish(s string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "true", "yes", "1":
		return "true", true
	case "false", "no", "0":
		return "false", true
	default:
		return "", false
	}
}

// clampLimit enforces sane bounds on the limit parameter, preventing excessively
// large or invalid values from being sent to the API.
func clampLimit(n, def, min, max int) int {
	if n <= 0 {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// buildAssuranceParamsFromQuery converts a QueryModel from the frontend into a
// url.Values map suitable for encoding as URL query parameters.
// It performs the following key operations:
// - Sets pagination parameters ('limit' and 'offset').
// - Adds time range filters ('startTime', 'endTime') if provided.
// - Adds normalized and validated filters for site, device, status, etc.
// - Skips any empty or invalid filter values to create a clean API request.
func buildAssuranceParamsFromQuery(q QueryModel, startTime, endTime int64, pageSize, offset int) url.Values {
	v := url.Values{}

	// Paging (one-based offset agreed)
	v.Set("limit", strconv.Itoa(clampLimit(pageSize, 100, 1, 1000)))
	if offset < 1 {
		offset = 1
	}
	v.Set("offset", strconv.Itoa(offset))

	// Optional time range (ignored if zero)
	if startTime > 0 {
		v.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		v.Set("endTime", strconv.FormatInt(endTime, 10))
	}

	// Filters (skip empties)
	if s := strings.TrimSpace(q.SiteID); s != "" {
		v.Set("siteId", s)
	}
	if s := strings.TrimSpace(q.DeviceID); s != "" {
		v.Set("deviceId", s)
	}
	if s := strings.TrimSpace(q.MacAddress); s != "" {
		v.Set("macAddress", s)
	}

	// Handle Priority: The API expects a comma-separated string.
	if len(q.Priority) > 0 {
		var validPriorities []string
		for _, p := range q.Priority {
			if norm, ok := normalizePriority(p, ""); ok {
				validPriorities = append(validPriorities, norm)
			}
		}
		if len(validPriorities) > 0 {
			v.Set("priority", strings.Join(validPriorities, ","))
		}
	}

	if st, ok := normalizeIssueStatus(q.IssueStatus, q.Status); ok {
		v.Set("status", strings.ToLower(st))
	}

	// AIDriven is a custom StringOrBool type (backward-compatible)
	if b, ok := normalizeBoolish(q.AIDriven.String()); ok {
		v.Set("aiDriven", b)
	}

	return v
}
