package backend

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

type InstanceSettings struct {
	BaseURL   string
	Username  string
	Password  string
	// Optional manual token override (if set, we’ll prefer it)
	APIToken string
}

// Read JSON/SecureJSON from Grafana settings
func ParseInstanceSettings(jsonData json.RawMessage, secureData map[string]string) (*InstanceSettings, error) {
	var jd struct {
		BaseURL string `json:"baseUrl"`
	}
	_ = json.Unmarshal(jsonData, &jd)

	s := &InstanceSettings{
		BaseURL:  strings.TrimRight(jd.BaseURL, "/"),
		Username: secureData["username"],
		Password: secureData["password"],
		APIToken: secureData["apiToken"],
	}
	return s, nil
}

// Build the absolute token URL from the configured base URL.
// If baseUrl is e.g. https://host/dna/intent/api/v1, token endpoint is https://host/dna/system/api/v1/auth/token
func TokenURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	// Replace path to token endpoint
	u.Path = "/dna/system/api/v1/auth/token"
	return u.String(), nil
}

// Issues URL: base + /issues
func IssuesURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/issues"
	return u.String(), nil
}

type QueryModel struct {
	QueryType   string  `json:"queryType"` // "alerts"
	SiteID      string  `json:"siteId,omitempty"`
	DeviceID    string  `json:"deviceId,omitempty"`
	MacAddress  string  `json:"macAddress,omitempty"`
	Priority    string  `json:"priority,omitempty"`    // P1,P2,P3,P4
	IssueStatus string  `json:"issueStatus,omitempty"` // ACTIVE,IGNORED,RESOLVED
	AIDriven    string  `json:"aiDriven,omitempty"`    // YES/NO
	Limit       *int64  `json:"limit,omitempty"`       // panel-side cap
	RefID       string  `json:"refId,omitempty"`
}

// Token cache entry
type tokenEntry struct {
	Token     string
	ExpiresAt time.Time // soft TTL; DNAC token usually ~60m
}

// Issues API response shape (very permissive)
type IssuesEnvelope struct {
	Response []map[string]any `json:"response"`
}
