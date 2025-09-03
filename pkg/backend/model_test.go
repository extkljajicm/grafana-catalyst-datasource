package backend

import "testing"

func TestIssuesURL_DataAssurance(t *testing.T) {
	u, err := IssuesURL("https://example.local/dna/intent/api/v1")
	if err != nil {
		t.Fatalf("IssuesURL error: %v", err)
	}
	want := "https://example.local/dna/data/api/v1/assuranceIssues"
	if u != want {
		t.Fatalf("IssuesURL = %q, want %q", u, want)
	}
}

func TestIssuesURL_PrefixPreserved(t *testing.T) {
	u, err := IssuesURL("https://gw/proxy/dnac/dna")
	if err != nil {
		t.Fatalf("IssuesURL error: %v", err)
	}
	want := "https://gw/proxy/dnac/dna/data/api/v1/assuranceIssues"
	if u != want {
		t.Fatalf("IssuesURL = %q, want %q", u, want)
	}
}

