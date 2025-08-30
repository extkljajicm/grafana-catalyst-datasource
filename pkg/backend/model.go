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
// Examples:
//
//	"/"                                -> ""
//	"/dna/intent/api/v1"               -> ""
//	"/proxy/x/dna/intent/api/v1"       -> "/proxy/x"
//	"/something" (no /dna present)     -> ""
func dnacPrefix(p string) string {
	if p == "" {
		return ""
	}
	// Find the first occurrence of "/dna"
	i := strings.Index(p, "/dna")
	if i == -1 {
		// No /dna in the path; treat as root
		return ""
	}
	// Everything before "/dna" is considered a prefix; trim trailing slash for clean joins
	return strings.TrimRight(p[:i], "/")
}

// TokenURL always points to <prefix>/dna/system/api/v1/auth/token
// It ignores any trailing segments after /dna to avoid inconsistent base URL setups.
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

// IssuesURL always points to <prefix>/dna/intent/api/v1/issues
// It ignores any trailing segments after /dna to avoid inconsistent base URL setups.
func IssuesURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	prefix := dnacPrefix(u.Path)
	u.Path = prefix + "/dna/intent/api/v1/issues"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

type QueryModel struct {
	QueryType   string `json:"queryType"`
	SiteID      string `json:"siteId,omitempty"`
	DeviceID    string `json:"deviceId,omitempty"`
	MacAddress  string `json:"macAddress,omitempty"`
	Priority    string `json:"priority,omitempty"`
	IssueStatus string `json:"issueStatus,omitempty"`
	AIDriven    string `json:"aiDriven,omitempty"`
	Limit       *int64 `json:"limit,omitempty"`
	RefID       string `json:"refId,omitempty"`
}

type tokenEntry struct {
	Token     string
	ExpiresAt int64 // epoch seconds
}

type IssuesEnvelope struct {
	Response []map[string]any `json:"response"`
}
