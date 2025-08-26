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

func TokenURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = "/dna/system/api/v1/auth/token"
	return u.String(), nil
}

func IssuesURL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/issues"
	return u.String(), nil
}

type QueryModel struct {
	QueryType   string  `json:"queryType"`
	SiteID      string  `json:"siteId,omitempty"`
	DeviceID    string  `json:"deviceId,omitempty"`
	MacAddress  string  `json:"macAddress,omitempty"`
	Priority    string  `json:"priority,omitempty"`
	IssueStatus string  `json:"issueStatus,omitempty"`
	AIDriven    string  `json:"aiDriven,omitempty"`
	Limit       *int64  `json:"limit,omitempty"`
	RefID       string  `json:"refId,omitempty"`
}

type tokenEntry struct {
	Token     string
	ExpiresAt int64 // epoch seconds
}

type IssuesEnvelope struct {
	Response []map[string]any `json:"response"`
}
