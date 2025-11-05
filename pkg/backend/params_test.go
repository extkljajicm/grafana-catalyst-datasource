package backend

import (
	"net/url"
	"testing"
)

func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		priority string
		want     string
		ok       bool
	}{
		{"P1", "p1", true},
		{"p2", "p2", true},
		{"P3", "p3", true},
		{"p4", "p4", true},
		{"", "", false},
		{"weird", "", false},
		{"P5", "", false},  // Invalid priority
		{"  p1  ", "p1", true},  // Test trimming
	}
	for _, tt := range tests {
		got, ok := normalizePriority(tt.priority)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("normalizePriority(%q) = (%q,%v), want (%q,%v)", tt.priority, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNormalizeIssueStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
		ok     bool
	}{
		{"ACTIVE", "active", true},
		{"resolved", "resolved", true},
		{"ignored", "ignored", true},
		{"active", "active", true},
		{"RESOLVED", "resolved", true},
		{"IGNORED", "ignored", true},
		{"", "", false},
		{"bad", "", false},
		{"  active  ", "active", true},  // Test trimming
		{"Invalid", "", false},
	}
	for _, tt := range tests {
		got, ok := normalizeIssueStatus(tt.status)
		if got != tt.want || ok != tt.ok {
			t.Errorf("normalizeIssueStatus(%q) = (%q, %v), want (%q, %v)", tt.status, got, ok, tt.want, tt.ok)
		}
	}
}

func TestBuildAssuranceParamsFromQuery(t *testing.T) {
	q := QueryModel{
		SiteID:          []string{"site-123"},
		NetworkDeviceID: "dev-456",
		MACAddress:      "00:11:22:33:44:55",
		Priority:        []string{"p2"},
		Status:          []string{"resolved"},
	}

	params := buildAssuranceParamsFromQuery(q, 1700000000000, 1700003600000, 100, 1)

	want := url.Values{
		"siteId":          []string{"site-123"},
		"networkDeviceId": []string{"dev-456"},
		"macAddress":      []string{"00:11:22:33:44:55"},
		"priority":        []string{"p2"},
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
	if params.Get("priority") != "p3" {
		t.Fatalf("priority = %q, want p3", params.Get("priority"))
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
