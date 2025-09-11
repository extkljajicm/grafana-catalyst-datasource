package backend

import (
	"net/url"
	"strconv"
	"strings"
)

// Allowed sets
var (
	allowedPriority    = map[string]struct{}{"P1": {}, "P2": {}, "P3": {}, "P4": {}}
	allowedIssueStatus = map[string]struct{}{"ACTIVE": {}, "RESOLVED": {}, "IGNORED": {}}
)

// normalizePriority returns a valid P1..P4 if present, falling back from severity if needed.
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

// normalizeIssueStatus returns a valid status if present, falling back from the legacy status field if needed.
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

// normalizeBoolish turns various inputs into "true"/"false".
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

// clampLimit enforces sane bounds while preserving explicit choices.
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

// buildAssuranceParamsFromQuery converts a QueryModel + paging + optional time range into ?k=v params.
// - Skips empty/invalid filters
// - Maps severity→priority if priority empty
// - Maps aiDriven string-ish to boolean string
// - Enforces one-based offset (caller must pass the correct offset)
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

	if st, ok := normalizeIssueStatus(q.IssueStatus, q.Status); ok {
		v.Set("status", strings.ToLower(st))
	}

	// AIDriven is a custom StringOrBool type (backward-compatible)
	if b, ok := normalizeBoolish(q.AIDriven.String()); ok {
		v.Set("aiDriven", b)
	}

	return v
}