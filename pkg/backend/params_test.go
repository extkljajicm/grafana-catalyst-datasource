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
		{"ACTIVE", "", "ACTIVE", true},
		{"resolved", "", "RESOLVED", true},
		{"", "ignored", "IGNORED", true},
		{"", "active", "ACTIVE", true},
		{"", "", "", false},
		{"bad", "also_bad", "", false},
		{"ACTIVE", "RESOLVED", "ACTIVE", true}, // Primary field should win
	}
	for _, tt := range tests {
		got, ok := normalizeIssueStatus(tt.issueStatus, tt.status)
		if got != tt.want || ok != tt.ok {
			t.Errorf("normalizeIssueStatus(%q, %q) = (%q, %v), want (%q, %v)", tt.issueStatus, tt.status, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNormalizeBoolish(t *testing.T) {
	trueVals := []string{"true", "TRUE", "yes", "1"}
	falseVals := []string{"false", "FALSE", "no", "0"}

	for _, v := range trueVals {
		got, ok := normalizeBoolish(v)
		if got != "true" || !ok {
			t.Fatalf("normalizeBoolish(%q) = (%q,%v), want (true,true)", v, got, ok)
		}
	}
	for _, v := range falseVals {
		got, ok := normalizeBoolish(v)
		if got != "false" || !ok {
			t.Fatalf("normalizeBoolish(%q) = (%q,%v), want (false,true)", v, got, ok)
		}
	}
	if _, ok := normalizeBoolish("maybe"); ok {
		t.Fatal("normalizeBoolish(maybe) expected not ok")
	}
}

func TestBuildAssuranceParamsFromQuery(t *testing.T) {
	q := QueryModel{
		SiteID:      "site-123",
		DeviceID:    "dev-456",
		MacAddress:  "00:11:22:33:44:55",
		Priority:    "p2",
		IssueStatus: "resolved",
		AIDriven:    StringOrBool("YES"),
		RefID:       "A",
		Severity:    "",
		Status:      "",
	}

	params := buildAssuranceParamsFromQuery(q, 1700000000000, 1700003600000, 100, 1)

	want := url.Values{
		"siteId":     []string{"site-123"},
		"deviceId":   []string{"dev-456"},
		"macAddress": []string{"00:11:22:33:44:55"},
		"priority":   []string{"P2"},
		"status":     []string{"resolved"},
		"aiDriven":   []string{"true"},
		"limit":      []string{"100"},
		"offset":     []string{"1"},
		"startTime":  []string{"1700000000000"},
		"endTime":    []string{"1700003600000"},
	}

	if got := params.Encode(); got != want.Encode() {
		t.Fatalf("params mismatch\ngot:  %q\nwant: %q", got, want.Encode())
	}
}

func TestBuildAssuranceParams_SkipEmpties(t *testing.T) {
	q := QueryModel{
		Severity: "P3", // legacy alias only
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