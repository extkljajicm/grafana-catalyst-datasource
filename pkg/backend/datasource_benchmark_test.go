package backend

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func generateBenchmarkDataEnvelope(n int) []byte {
	items := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		items[i] = map[string]any{
			"issueId":   fmt.Sprintf("ID-%d", i),
			"name":      "Issue Name",
			"timestamp": float64(1678886400000 + i),
			"siteId":    "site-1",
			"priority":  "P1",
			"status":    "Open",
		}
	}
	env := map[string]any{"response": items}
	b, _ := json.Marshal(env)
	return b
}

func generateBenchmarkDataArray(n int) []byte {
	items := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		items[i] = map[string]any{
			"issueId":   fmt.Sprintf("ID-%d", i),
			"name":      "Issue Name",
			"timestamp": float64(1678886400000 + i),
			"siteId":    "site-1",
			"priority":  "P1",
			"status":    "Open",
		}
	}
	b, _ := json.Marshal(items)
	return b
}

// Current/Old unmarshaling method used in datasource.go
func unmarshalIssuesCurrent(body []byte) []map[string]any {
	var env IssuesEnvelope
	var arr []map[string]any
	if err := json.Unmarshal(body, &env); err == nil && len(env.Response) > 0 {
		arr = env.Response
	} else {
		_ = json.Unmarshal(body, &arr)
	}
	return arr
}

func TestParseIssuesResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Envelope response",
			input:    `{"response": [{"issueId": "1"}, {"issueId": "2"}]}`,
			expected: 2,
		},
		{
			name:     "Array response",
			input:    `[{"issueId": "1"}, {"issueId": "2"}]`,
			expected: 2,
		},
		{
			name:     "Empty envelope",
			input:    `{"response": []}`,
			expected: 0,
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: 0,
		},
		{
			name:     "Whitespace padded array",
			input:    `   [{"issueId": "1"}]   `,
			expected: 1,
		},
		{
			name:     "Invalid JSON",
			input:    `invalid json`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOld := unmarshalIssuesCurrent([]byte(tt.input))
			gotNew := parseIssuesResponse([]byte(tt.input))

			if len(gotNew) != tt.expected {
				t.Errorf("parseIssuesResponse() returned %d items, want %d", len(gotNew), tt.expected)
			}

			// Validate equivalence with old implementation (for valid responses)
			if tt.input != "invalid json" && len(gotOld) != len(gotNew) {
				t.Errorf("mismatch with old implementation: old len=%d, new len=%d", len(gotOld), len(gotNew))
			}
		})
	}
}

func BenchmarkUnmarshalCurrent_Envelope25(b *testing.B) {
	body := generateBenchmarkDataEnvelope(25)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = unmarshalIssuesCurrent(body)
	}
}

func BenchmarkUnmarshalOptimized_Envelope25(b *testing.B) {
	body := generateBenchmarkDataEnvelope(25)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseIssuesResponse(body)
	}
}

func BenchmarkUnmarshalCurrent_Array25(b *testing.B) {
	body := generateBenchmarkDataArray(25)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = unmarshalIssuesCurrent(body)
	}
}

func BenchmarkUnmarshalOptimized_Array25(b *testing.B) {
	body := generateBenchmarkDataArray(25)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseIssuesResponse(body)
	}
}

func BenchmarkUnmarshalCurrent_EmptyEnvelope(b *testing.B) {
	body := []byte(`{"response": []}`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = unmarshalIssuesCurrent(body)
	}
}

func BenchmarkUnmarshalOptimized_EmptyEnvelope(b *testing.B) {
	body := []byte(`{"response": []}`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseIssuesResponse(body)
	}
}

func BenchmarkUnmarshalCurrent_EmptyArray(b *testing.B) {
	body := []byte(`[]`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = unmarshalIssuesCurrent(body)
	}
}

func BenchmarkUnmarshalOptimized_EmptyArray(b *testing.B) {
	body := []byte(`[]`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseIssuesResponse(body)
	}
}

// Silence unused import warning for reflect if needed
var _ = reflect.DeepEqual
