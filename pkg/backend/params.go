// Package backend contains the core logic for the Catalyst datasource.
// This file, params.go, is responsible for converting the frontend query model
// into the URL query parameters expected by the Catalyst Center API. It handles
// normalization, validation, and formatting of filter values.

package backend

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// buildSiteHealthParamsFromQuery converts a QueryModel into url.Values for the site-health endpoint.
func buildSiteHealthParamsFromQuery(q QueryModel, timestamp int64, limit int, offset int) url.Values {
	p := url.Values{}
	p.Set("timestamp", fmt.Sprintf("%d", timestamp))

	if q.SiteType != "" {
		p.Set("siteType", q.SiteType)
	}

	// The site-health endpoint does not support filtering by siteId or siteName directly in the query params.
	// This filtering should be done on the client-side after fetching the data,
	// or the query needs to be adapted if a different endpoint is more suitable.
	// For now, we will not add siteId/siteName to the query.

	// Clamp limit to sane values: default 50, min 1, max 50 (API max)
	clampedLimit := clampLimit(limit, 50, 1, 50)
	p.Set("limit", strconv.Itoa(clampedLimit))
	
	// Ensure offset is at least 1 (API uses 1-based indexing)
	if offset < 1 {
		offset = 1
	}
	p.Set("offset", strconv.Itoa(offset))

	return p
}

// Allowed value sets for validation and normalization.
var (
	// allowedPriority defines the valid priority values for the API.
	// Priority values are case-insensitive (p1, P1) but normalized to lowercase for API.
	allowedPriority = map[string]struct{}{"p1": {}, "p2": {}, "p3": {}, "p4": {}}
	// allowedIssueStatus defines the valid status values for the API.
	// Status values are case-insensitive (ACTIVE, active) but normalized to lowercase for API.
	allowedIssueStatus = map[string]struct{}{"active": {}, "resolved": {}, "ignored": {}}
)

// normalizePriority returns a valid priority string (p1-p4) in lowercase if the input
// matches a known value. Accepts case-insensitive input (e.g., P1, p1, P2).
func normalizePriority(priority string) (string, bool) {
	p := strings.ToLower(strings.TrimSpace(priority))
	if _, ok := allowedPriority[p]; ok {
		return p, true
	}
	return "", false
}

// normalizeIssueStatus returns a valid status string in lowercase if the input matches a known
// value. Accepts case-insensitive input (e.g., ACTIVE, active, Active).
func normalizeIssueStatus(status string) (string, bool) {
	s := strings.ToLower(strings.TrimSpace(status))
	if _, ok := allowedIssueStatus[s]; ok {
		return s, true
	}
	return "", false
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
// url.Values map for querying the assurance issues endpoint.
func buildAssuranceParamsFromQuery(q QueryModel, startTime, endTime int64, pageSize, offset int) url.Values {
	p := url.Values{}
	
	// Only set time parameters if they are non-zero
	if startTime > 0 {
		p.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		p.Set("endTime", strconv.FormatInt(endTime, 10))
	}

       if len(q.Priority) > 0 {
	       for _, prio := range q.Priority {
		       normPrio, ok := normalizePriority(prio)
		       if ok {
			       p.Add("priority", normPrio)
		       }
	       }
       }
       if len(q.Status) > 0 {
	       for _, stat := range q.Status {
		       normStatus, ok := normalizeIssueStatus(stat)
		       if ok {
			       p.Add("status", normStatus)
		       }
	       }
       }
	if q.NetworkDeviceID != "" {
		p.Set("networkDeviceId", q.NetworkDeviceID)
	}
	if q.MACAddress != "" {
		p.Set("macAddress", q.MACAddress)
	}
       // Send all site IDs as repeated keys, per API spec.
       if len(q.SiteID) > 0 {
	       for _, id := range q.SiteID {
		       if id != "" {
			       p.Add("siteId", id)
		       }
	       }
       }
	if q.IssueName != "" {
		// The 'IssueName' from the UI corresponds to the 'name' of the issue in the API
		p.Set("name", q.IssueName)
	}
	if q.AIDriven {
		p.Set("aiDriven", "true")
	}
	if q.IsGlobal {
		p.Set("isGlobal", "true")
	}

	// Clamp limit to sane values: default 100, min 1, max 500
	clampedLimit := clampLimit(pageSize, 100, 1, 500)
	p.Set("limit", strconv.Itoa(clampedLimit))
	
	// Ensure offset is at least 1 (API uses 1-based indexing)
	if offset < 1 {
		offset = 1
	}
	p.Set("offset", strconv.Itoa(offset))

	return p
}
