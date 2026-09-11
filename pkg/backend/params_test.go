package backend

import (
	"net/url"
	"testing"
)

func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		priority string
		severity string
		want     string
		ok       bool
	}{
		{"P1", "", "P1", true},
		{"p2", "", "P2", true},
		{"", "P3", "P3", true},
		{"", "p4", "P4", true},
		{"", "", "", false},
		{"weird", "nope", "", false},
		{"P1", "P2", "P1", true}, // Primary field (priority) wins
		{"  p1  ", "", "P1", true}, // Test trimming
	}
	for _, tt := range tests {
		got, ok := normalizePriority(tt.priority, tt.severity)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("normalizePriority(%q,%q) = (%q,%v), want (%q,%v)", tt.priority, tt.severity, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNormalizeIssueStatus(t *testing.T) {
	tests := []struct {
		issueStatus string
		status      string
		want        string
		ok          bool
	}{
		{"ACTIVE", "", "active", true},
		{"resolved", "", "resolved", true},
		{"", "ignored", "ignored", true},
		{"", "active", "active", true},
		{"", "", "", false},
		{"bad", "also_bad", "", false},
		{"ACTIVE", "RESOLVED", "active", true}, // Primary field (issueStatus) wins
		{"  resolved  ", "", "resolved", true},  // Test trimming
	}
	for _, tt := range tests {
		got, ok := normalizeIssueStatus(tt.issueStatus, tt.status)
		if got != tt.want || ok != tt.ok {
			t.Errorf("normalizeIssueStatus(%q, %q) = (%q, %v), want (%q, %v)", tt.issueStatus, tt.status, got, ok, tt.want, tt.ok)
		}
	}
}

func TestBuildAssuranceParamsFromQuery(t *testing.T) {
	q := QueryModel{
		SiteID:          []string{"site-123"},
		NetworkDeviceID: "dev-456",
		MACAddress:      "00:11:22:33:44:55",
		Priority:        []string{"P2"},
		Status:          []string{"resolved"},
	}

	params := buildAssuranceParamsFromQuery(q, 1700000000000, 1700003600000, 100, 1)

	want := url.Values{
		"siteId":          []string{"site-123"},
		"networkDeviceId": []string{"dev-456"},
		"macAddress":      []string{"00:11:22:33:44:55"},
		"priority":        []string{"P2"},
		"status":          []string{"resolved"},
		"limit":           []string{"100"},
		"offset":          []string{"1"},
		"startTime":       []string{"1700000000000"},
		"endTime":         []string{"1700003600000"},
	}

	if got := params.Encode(); got != want.Encode() {
		t.Fatalf("params mismatch\ngot:  %q\nwant: %q", got, want.Encode())
	}
}

func TestBuildAssuranceParams_SkipEmpties(t *testing.T) {
	q := QueryModel{
		Priority: []string{"P3"}, // use Priority field as per new struct
	}

	params := buildAssuranceParamsFromQuery(q, 0, 0, -5, 0) // bad page/offset should be clamped/fixed
	if _, ok := params["priority"]; !ok {
		t.Fatal("expected priority from severity")
	}
	if params.Get("priority") != "P3" {
		t.Fatalf("priority = %q, want P3", params.Get("priority"))
	}
	if params.Get("limit") != "100" { // default page size
		t.Fatalf("limit = %q, want 100", params.Get("limit"))
	}
	if params.Get("offset") != "1" {
		t.Fatalf("offset = %q, want 1", params.Get("offset"))
	}
	// No startTime/endTime unless non-zero
	if _, ok := params["startTime"]; ok {
		t.Fatal("startTime should be omitted")
	}
	if _, ok := params["endTime"]; ok {
		t.Fatal("endTime should be omitted")
	}
}
