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

	p.Set("limit", strconv.Itoa(limit))
	p.Set("offset", strconv.Itoa(offset))

	return p
}

// Allowed value sets for validation and normalization.
var (
	// allowedPriority defines the valid priority values for the API.
	allowedPriority = map[string]struct{}{"P1": {}, "P2": {}, "P3": {}, "P4": {}}
	// allowedIssueStatus defines the valid status values for the API.
	allowedIssueStatus = map[string]struct{}{"ACTIVE": {}, "RESOLVED": {}, "IGNORED": {}}
)

// normalizePriority returns a valid priority string (P1-P4) if the input
// matches a known value.
func normalizePriority(priority string) (string, bool) {
	p := strings.ToUpper(strings.TrimSpace(priority))
	if _, ok := allowedPriority[p]; ok {
		return p, true
	}
	return "", false
}

// normalizeIssueStatus returns a valid status string if the input matches a known
// value.
func normalizeIssueStatus(status string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(status))
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
	p.Set("startTime", strconv.FormatInt(startTime, 10))
	p.Set("endTime", strconv.FormatInt(endTime, 10))

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

	p.Set("limit", strconv.Itoa(pageSize))
	p.Set("offset", strconv.Itoa(offset))

	return p
}
