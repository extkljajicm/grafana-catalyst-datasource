package backend

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// InstanceSettings holds the configuration for a single instance of the datasource.
// This includes the base URL for the Catalyst Center API and connection-specific
// settings like TLS verification and credentials.
type InstanceSettings struct {
	// BaseURL is the root URL of the Catalyst Center API.
	// Example: https://catalyst.example.com
	BaseURL string
	// InsecureSkipVerify allows the user to disable TLS certificate verification.
	// This should only be used in development or for self-signed certificates.
	InsecureSkipVerify bool
	// Username for API authentication.
	Username string
	// Password for API authentication.
	Password string
	// APIToken allows for manual override of the token, bypassing username/password auth.
	APIToken string
}

// ParseInstanceSettings unmarshals and validates the datasource instance settings
// from the Grafana plugin context.
func ParseInstanceSettings(jsonData json.RawMessage, secureData map[string]string) (*InstanceSettings, error) {
	var jd struct {
		BaseURL            string `json:"baseUrl"`
		InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	}
	_ = json.Unmarshal(jsonData, &jd)

	s := &InstanceSettings{
		BaseURL:            strings.TrimRight(jd.BaseURL, "/"),
		InsecureSkipVerify: jd.InsecureSkipVerify,
		Username:           secureData["username"],
		Password:           secureData["password"],
		APIToken:           secureData["apiToken"],
	}
	return s, nil
}

// dnacPrefix extracts any reverse-proxy prefix that appears BEFORE the /dna path segment.
// This is crucial for ensuring API calls are correctly routed when Catalyst Center
// is behind a reverse proxy.
// Example:
//
//	"/proxy/dnac/dna/intent/api/v1" -> "/proxy/dnac"
//	"/dna/intent/api/v1"            -> ""
//	"/something" (no /dna)          -> ""
func dnacPrefix(p string) string {
	if p == "" {
		return ""
	}
	segs := strings.Split(p, "/")
	idx := -1
	for i, s := range segs {
		if s == "dna" {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ""
	}
	prefix := strings.Join(segs[:idx], "/")
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return strings.TrimRight(prefix, "/")
}

// TokenURL constructs the full URL for the authentication token endpoint,
// preserving any reverse proxy prefix from the base URL.
// It always points to <prefix>/dna/system/api/v1/auth/token.
func TokenURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	prefix := dnacPrefix(u.Path)
	u.Path = prefix + "/dna/system/api/v1/auth/token"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// IssuesURL constructs the full URL for the issues/alerts endpoint,
// preserving any reverse proxy prefix.
// It always points to <prefix>/dna/data/api/v1/assuranceIssues.
func IssuesURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	prefix := dnacPrefix(u.Path)
	u.Path = prefix + "/dna/data/api/v1/assuranceIssues"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// SiteURL constructs the full URL for the site lookup endpoint,
// preserving any reverse proxy prefix.
// It always points to <prefix>/dna/intent/api/v2/site.
func SiteURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	prefix := dnacPrefix(u.Path)
	u.Path = prefix + "/dna/intent/api/v2/site"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// StringOrBool is a custom type that can unmarshal both boolean (true/false)
// and string ("true", "false", "yes", "no") values from JSON into a normalized
// string representation. This provides flexibility for API fields that might
// return booleans in different formats.
type StringOrBool string

func (v *StringOrBool) UnmarshalJSON(b []byte) error {
	// Try boolean first
	var bv bool
	if err := json.Unmarshal(b, &bv); err == nil {
		if bv {
			*v = StringOrBool("true")
		} else {
			*v = StringOrBool("false")
		}
		return nil
	}
	// Then string
	var sv string
	if err := json.Unmarshal(b, &sv); err == nil {
		s := strings.ToLower(strings.TrimSpace(sv))
		switch s {
		case "true", "yes", "1":
			*v = StringOrBool("true")
		case "false", "no", "0":
			*v = StringOrBool("false")
		default:
			*v = StringOrBool(sv)
		}
		return nil
	}
	// Be lenient on unknown types
	return nil
}

func (v StringOrBool) String() string { return string(v) }

// QueryModel represents the query structure sent from the frontend.
// It is used by the backend to construct API requests.
type QueryModel struct {
	TimeRange backend.TimeRange `json:"timeRange"`
	QueryType string            `json:"queryType,omitempty"`
	Limit     *int64            `json:"limit,omitempty"`
	// Filters for assurance issues
	Priority        []string `json:"priority,omitempty"`
	Status          []string `json:"status,omitempty"`
	NetworkDeviceID string   `json:"networkDeviceId,omitempty"`
	MACAddress      string   `json:"macAddress,omitempty"`
	SiteID          []string `json:"siteId,omitempty"`
	SiteName        []string `json:"siteName,omitempty"`
	IssueName       string   `json:"issueName,omitempty"`
	AIDriven        bool     `json:"aiDriven,omitempty"`
	IsGlobal        bool     `json:"isGlobal,omitempty"`

	// Filters for site-health
	SiteType       string   `json:"siteType,omitempty"`
	ParentSiteName string   `json:"parentSiteName,omitempty"`
	ParentSiteID   string   `json:"parentSiteId,omitempty"`
	Metrics        []string `json:"metrics,omitempty"`
}

// IssuesEnvelope is used to unmarshal the 'response' array from the issues API.
// The API response is sometimes wrapped in a `{"response": [...]}` object.
// This envelope allows for consistent handling of the response data.
type IssuesEnvelope struct {
	Response []map[string]any `json:"response"`
}

// SiteEnvelope defines the structure for the site API response.
type SiteEnvelope struct {
	Response []Site `json:"response"`
}

// Site represents a single site returned from the Catalyst Center API.
type Site struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId,omitempty"`
}
