package backend

import (
	"encoding/json"
	"net/url"
	"strings"
)

type InstanceSettings struct {
	BaseURL            string
	InsecureSkipVerify bool
	Username           string
	Password           string
	APIToken           string
}

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

// dnacPrefix extracts any reverse-proxy prefix that appears BEFORE the /dna path.
// Example:
//   "/proxy/dnac/dna/intent/api/v1" -> "/proxy/dnac"
//   "/dna/intent/api/v1"            -> ""
//   "/something" (no /dna)          -> ""
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

// TokenURL always points to <prefix>/dna/system/api/v1/auth/token
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

// IssuesURL always points to <prefix>/dna/data/api/v1/assuranceIssues
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

// SiteURL always points to <prefix>/dna/intent/api/v1/site
func SiteURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	prefix := dnacPrefix(u.Path)
	u.Path = prefix + "/dna/intent/api/v1/site"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// StringOrBool accepts boolean or string JSON and preserves a normalized string form.
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

type QueryModel struct {
	QueryType   string       `json:"queryType"`
	SiteID      string       `json:"siteId,omitempty"`
	DeviceID    string       `json:"deviceId,omitempty"`
	MacAddress  string       `json:"macAddress,omitempty"`
	Priority    []string       `json:"priority,omitempty"`
	IssueStatus string       `json:"issueStatus,omitempty"`
	AIDriven    StringOrBool `json:"aiDriven,omitempty"`
	Limit       *int64       `json:"limit,omitempty"`
	RefID       string       `json:"refId,omitempty"`
	Enrich      bool         `json:"enrich,omitempty"`

	// Optional aliases for backward-compat in param builder (if used)
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
}

type tokenEntry struct {
	Token     string
	ExpiresAt int64 // epoch seconds
}

type IssuesEnvelope struct {
	Response []map[string]any `json:"response"`
}

// SiteEnvelope defines the structure for the site API response.
type SiteEnvelope struct {
	Response []Site `json:"response"`
}

// Site holds the relevant fields from the site API.
type Site struct {
	ID   string `json:"id"`
	Name string `json:"siteName"`
}
