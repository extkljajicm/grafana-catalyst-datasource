package backend

import (
	"encoding/json"
	"testing"
)

// mimic the structure of the data in the loop
var allIssues []map[string]any

func init() {
	allIssues = make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		allIssues[i] = map[string]any{
			"issueId":   "ID-12345",
			"name":      "Issue Name",
			"timestamp": float64(1678886400000),
			"siteId":    "site-1",
			"priority":  "P1",
			"status":    "Open",
		}
	}
}

func BenchmarkAllocInLoop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, it := range allIssues {
			getStr := func(k string) string {
				if v, ok := it[k]; ok && v != nil {
					if s, ok2 := v.(string); ok2 {
						return s
					}
				}
				return ""
			}
			getNum := func(k string) int64 {
				if v, ok := it[k]; ok && v != nil {
					switch x := v.(type) {
					case float64:
						return int64(x)
					case int64:
						return x
					case json.Number:
						n, _ := x.Int64()
						return n
					}
				}
				return 0
			}

			_ = getStr("issueId")
			_ = getStr("name")
			_ = getNum("timestamp")
		}
	}
}

func getStrHelper(it map[string]any, k string) string {
	if v, ok := it[k]; ok && v != nil {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

func getNumHelper(it map[string]any, k string) int64 {
	if v, ok := it[k]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return int64(x)
		case int64:
			return x
		case json.Number:
			n, _ := x.Int64()
			return n
		}
	}
	return 0
}

func BenchmarkAllocOutLoop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, it := range allIssues {
			_ = getStrHelper(it, "issueId")
			_ = getStrHelper(it, "name")
			_ = getNumHelper(it, "timestamp")
		}
	}
}

var rawArrayJSON = []byte(`[{"issueId":"ID-12345","name":"Issue Name","timestamp":1678886400000,"siteId":"site-1","priority":"P1","status":"Open"}]`)
var envelopeJSON = []byte(`{"response":[{"issueId":"ID-12345","name":"Issue Name","timestamp":1678886400000,"siteId":"site-1","priority":"P1","status":"Open"}]}`)

func BenchmarkUnmarshalOld_Envelope(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var env IssuesEnvelope
		var arr []map[string]any
		if err := json.Unmarshal(envelopeJSON, &env); err == nil && len(env.Response) > 0 {
			arr = env.Response
		} else {
			_ = json.Unmarshal(envelopeJSON, &arr)
		}
		_ = arr
	}
}

func BenchmarkUnmarshalOld_Array(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var env IssuesEnvelope
		var arr []map[string]any
		if err := json.Unmarshal(rawArrayJSON, &env); err == nil && len(env.Response) > 0 {
			arr = env.Response
		} else {
			_ = json.Unmarshal(rawArrayJSON, &arr)
		}
		_ = arr
	}
}

func BenchmarkUnmarshalNew_Envelope(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var env IssuesEnvelope
		var arr []map[string]any

		// find first non-whitespace
		var firstByte byte
		for _, c := range envelopeJSON {
			if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
				firstByte = c
				break
			}
		}

		if firstByte == '{' {
			if err := json.Unmarshal(envelopeJSON, &env); err == nil && len(env.Response) > 0 {
				arr = env.Response
			}
		} else {
			_ = json.Unmarshal(envelopeJSON, &arr)
		}
		_ = arr
	}
}

func BenchmarkUnmarshalNew_Array(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var env IssuesEnvelope
		var arr []map[string]any

		// find first non-whitespace
		var firstByte byte
		for _, c := range rawArrayJSON {
			if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
				firstByte = c
				break
			}
		}

		if firstByte == '{' {
			if err := json.Unmarshal(rawArrayJSON, &env); err == nil && len(env.Response) > 0 {
				arr = env.Response
			}
		} else {
			_ = json.Unmarshal(rawArrayJSON, &arr)
		}
		_ = arr
	}
}
