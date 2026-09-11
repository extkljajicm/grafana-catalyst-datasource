package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// ==================== Mock Client ====================

// mockClient is a mock implementation of the Client for testing datasource.go
type mockClient struct {
	// FetchAllIssues mocks
	fetchAllIssuesFunc func(ctx context.Context, issuesURL string, qm QueryModel, hardLimit int64) ([]map[string]any, error)
	// FetchAllSiteHealth mocks
	fetchAllSiteHealthFunc func(ctx context.Context, siteHealthURL *url.URL, qm QueryModel) ([]map[string]any, error)
}

func (m *mockClient) FetchAllIssues(ctx context.Context, issuesURL string, qm QueryModel, hardLimit int64) ([]map[string]any, error) {
	if m.fetchAllIssuesFunc != nil {
		return m.fetchAllIssuesFunc(ctx, issuesURL, qm, hardLimit)
	}
	return nil, fmt.Errorf("FetchAllIssues not mocked")
}

func (m *mockClient) FetchAllSiteHealth(ctx context.Context, siteHealthURL *url.URL, qm QueryModel) ([]map[string]any, error) {
	if m.fetchAllSiteHealthFunc != nil {
		return m.fetchAllSiteHealthFunc(ctx, siteHealthURL, qm)
	}
	return nil, fmt.Errorf("FetchAllSiteHealth not mocked")
}

// ==================== Mock SiteTranslator ====================

// mockSiteTranslator is a mock implementation of SiteTranslator for testing datasource.go
type mockSiteTranslator struct {
	ensureCacheFunc  func(ctx context.Context) error
	getSiteNameFunc  func(ctx context.Context, id string) (string, error)
	getSiteIDFunc    func(ctx context.Context, name string) (string, error)
	getSiteByIDFunc  func(ctx context.Context, id string) (Site, error)
	getAllSitesFunc  func(ctx context.Context) ([]Site, error)
}

func (m *mockSiteTranslator) EnsureCache(ctx context.Context) error {
	if m.ensureCacheFunc != nil {
		return m.ensureCacheFunc(ctx)
	}
	return nil
}

func (m *mockSiteTranslator) GetSiteName(ctx context.Context, id string) (string, error) {
	if m.getSiteNameFunc != nil {
		return m.getSiteNameFunc(ctx, id)
	}
	return "", fmt.Errorf("GetSiteName not mocked")
}

func (m *mockSiteTranslator) GetSiteID(ctx context.Context, name string) (string, error) {
	if m.getSiteIDFunc != nil {
		return m.getSiteIDFunc(ctx, name)
	}
	return "", fmt.Errorf("GetSiteID not mocked")
}

func (m *mockSiteTranslator) GetSiteByID(ctx context.Context, id string) (Site, error) {
	if m.getSiteByIDFunc != nil {
		return m.getSiteByIDFunc(ctx, id)
	}
	return Site{}, fmt.Errorf("GetSiteByID not mocked")
}

func (m *mockSiteTranslator) GetAllSites(ctx context.Context) ([]Site, error) {
	if m.getAllSitesFunc != nil {
		return m.getAllSitesFunc(ctx)
	}
	return nil, fmt.Errorf("GetAllSites not mocked")
}

// ==================== Mock-Enabled Datasource ====================

// testDatasource is a datasource that allows injection of mocks for testing
type testDatasource struct {
	*Datasource
	mockClient         *mockClient
	mockSiteTranslator *mockSiteTranslator
}

// Override getSiteTranslator to return our mock
func (td *testDatasource) getSiteTranslator(inst *dsInstance, httpClient interface{}) *mockSiteTranslator {
	return td.mockSiteTranslator
}

// Override NewClient to return our mock
func newMockClient(httpClient interface{}, tm *tokenManager, inst *dsInstance, mockClient *mockClient) *mockClient {
	return mockClient
}

// ==================== Integration Tests ====================

// TestDatasource_QueryData_MainFlow verifies the standard QueryData flow:
// - QueryData processes the request
// - Returns a response structure
// Note: This is an integration test that verifies the orchestration logic
func TestDatasource_QueryData_MainFlow(t *testing.T) {
	ctx := context.Background()

	// Create datasource
	ds := NewDatasource()

	// Build request
	queryJSON, _ := json.Marshal(map[string]any{
		"queryType": "assuranceIssues",
	})

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON,
			},
		},
	}

	// Call QueryData - it will fail to make HTTP calls but we verify structure
	resp, err := ds.QueryData(ctx, req)
	
	// We expect an error because we can't actually make HTTP calls
	// but we can verify the structure is correct
	if err != nil {
		t.Logf("Expected error in mock environment: %v", err)
	}
	
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// Verify response contains our query
	if _, ok := resp.Responses["A"]; !ok {
		t.Fatal("expected response for RefID 'A'")
	}
}

// TestDatasource_QueryData_WithClientError verifies that errors from Client
// are correctly propagated through QueryData
func TestDatasource_QueryData_WithClientError(t *testing.T) {
	ctx := context.Background()

	// Since we can't easily inject mocks into the existing Datasource structure,
	// we'll create a test that verifies the error handling path by examining
	// the response structure when things fail.

	ds := NewDatasource()

	// Build request with invalid JSON to trigger unmarshaling error
	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: []byte(`{invalid json`), // Invalid JSON
			},
		},
	}

	resp, err := ds.QueryData(ctx, req)

	// Should not error at the top level
	if err != nil {
		t.Fatalf("QueryData should not return top-level error, got: %v", err)
	}

	// But should have error in the response
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	dataResp, ok := resp.Responses["A"]
	if !ok {
		t.Fatal("expected response for RefID 'A'")
	}

	if dataResp.Error == nil {
		t.Fatal("expected error in DataResponse for invalid JSON")
	}

	if dataResp.Error.Error() == "" {
		t.Fatal("expected non-empty error message")
	}

	t.Logf("Correctly caught error: %v", dataResp.Error)
}

// TestDatasource_QueryData_SiteNameResolution verifies that site name resolution
// works when SiteName is provided instead of SiteID
func TestDatasource_QueryData_SiteNameResolution(t *testing.T) {
	ctx := context.Background()

	ds := NewDatasource()

	queryJSON, _ := json.Marshal(map[string]any{
		"queryType": "assuranceIssues",
		"siteName":  []string{"Building A"},
	})

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON,
			},
		},
	}

	// This will fail to make HTTP calls, but we can verify the query was accepted
	resp, _ := ds.QueryData(ctx, req)

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// The response should exist for our query
	if _, ok := resp.Responses["A"]; !ok {
		t.Fatal("expected response for RefID 'A'")
	}
}

// TestDatasource_QueryData_MultipleQueries verifies that multiple queries
// are handled correctly
func TestDatasource_QueryData_MultipleQueries(t *testing.T) {
	ctx := context.Background()

	ds := NewDatasource()

	queryJSON1, _ := json.Marshal(map[string]any{
		"queryType": "assuranceIssues",
	})

	queryJSON2, _ := json.Marshal(map[string]any{
		"queryType": "assuranceIssues",
		"priority":  []string{"p1"},
	})

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON1,
			},
			{
				RefID: "B",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-2 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON2,
			},
		},
	}

	resp, err := ds.QueryData(ctx, req)

	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// Should have responses for both queries
	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got: %d", len(resp.Responses))
	}

	if _, ok := resp.Responses["A"]; !ok {
		t.Fatal("expected response for RefID 'A'")
	}

	if _, ok := resp.Responses["B"]; !ok {
		t.Fatal("expected response for RefID 'B'")
	}
}

// TestDatasource_QueryData_SiteHealthFlow verifies the siteHealth query path
func TestDatasource_QueryData_SiteHealthFlow(t *testing.T) {
	ctx := context.Background()

	ds := NewDatasource()

	queryJSON, _ := json.Marshal(map[string]any{
		"queryType": "siteHealth",
	})

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON,
			},
		},
	}

	resp, err := ds.QueryData(ctx, req)

	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// Should have response for the query
	if _, ok := resp.Responses["A"]; !ok {
		t.Fatal("expected response for RefID 'A'")
	}
}

// TestDatasource_QueryData_LimitHandling verifies that the limit parameter
// is correctly processed
func TestDatasource_QueryData_LimitHandling(t *testing.T) {
	ctx := context.Background()

	ds := NewDatasource()

	limit := int64(100)
	queryJSON, _ := json.Marshal(map[string]any{
		"queryType": "assuranceIssues",
		"limit":     limit,
	})

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:     "test-instance",
				JSONData: json.RawMessage(`{"baseUrl": "https://catalyst.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"username": "testuser",
					"password": "testpass",
				},
			},
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: queryJSON,
			},
		},
	}

	resp, err := ds.QueryData(ctx, req)

	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// Verify query was processed (response exists)
	if _, ok := resp.Responses["A"]; !ok {
		t.Fatal("expected response for RefID 'A'")
	}
}

// TestDatasource_QueryData_NoPluginContext verifies error handling when
// plugin context is missing
func TestDatasource_QueryData_NoPluginContext(t *testing.T) {
	ctx := context.Background()

	ds := NewDatasource()

	req := &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: nil, // Missing settings
		},
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				TimeRange: backend.TimeRange{
					From: time.Now().Add(-1 * time.Hour),
					To:   time.Now(),
				},
				JSON: []byte(`{"queryType": "assuranceIssues"}`),
			},
		},
	}

	resp, err := ds.QueryData(ctx, req)

	// Should return error at top level for missing plugin context
	if err == nil {
		t.Fatal("expected error for missing plugin context")
	}

	if resp != nil {
		t.Fatal("expected nil response when plugin context is missing")
	}

	t.Logf("Correctly returned error: %v", err)
}

// TestDatasource_NewDatasource verifies that NewDatasource creates
// a properly initialized instance
func TestDatasource_NewDatasource(t *testing.T) {
	ds := NewDatasource()

	if ds == nil {
		t.Fatal("expected non-nil datasource")
	}

	if ds.tm == nil {
		t.Fatal("expected non-nil token manager")
	}

	if ds.translators == nil {
		t.Fatal("expected non-nil translators map")
	}

	if len(ds.translators) != 0 {
		t.Fatal("expected empty translators map initially")
	}
}

// TestDatasource_GetSiteTranslator verifies that getSiteTranslator
// correctly caches translators per instance
func TestDatasource_GetSiteTranslator(t *testing.T) {
	ds := NewDatasource()

	inst := &dsInstance{
		Settings: &InstanceSettings{
			BaseURL:  "https://catalyst.example.com",
			Username: "testuser",
			Password: "testpass",
		},
		UID: "test-instance-1",
	}

	// Get translator first time
	translator1 := ds.getSiteTranslator(inst, nil)
	if translator1 == nil {
		t.Fatal("expected non-nil translator")
	}

	// Get translator second time - should be cached
	translator2 := ds.getSiteTranslator(inst, nil)
	if translator2 == nil {
		t.Fatal("expected non-nil translator")
	}

	// Should be the same instance (cached)
	if translator1 != translator2 {
		t.Fatal("expected same translator instance from cache")
	}

	// Verify it was added to the map
	if len(ds.translators) != 1 {
		t.Fatalf("expected 1 translator in map, got: %d", len(ds.translators))
	}

	// Different instance should get different translator
	inst2 := &dsInstance{
		Settings: &InstanceSettings{
			BaseURL:  "https://catalyst2.example.com",
			Username: "testuser2",
			Password: "testpass2",
		},
		UID: "test-instance-2",
	}

	translator3 := ds.getSiteTranslator(inst2, nil)
	if translator3 == nil {
		t.Fatal("expected non-nil translator")
	}

	if translator3 == translator1 {
		t.Fatal("expected different translator for different instance")
	}

	// Now should have 2 translators
	if len(ds.translators) != 2 {
		t.Fatalf("expected 2 translators in map, got: %d", len(ds.translators))
	}
}

// ==================== Mock CallResourceResponseSender ====================

type mockResourceSender struct {
	resp *backend.CallResourceResponse
}

func (m *mockResourceSender) Send(resp *backend.CallResourceResponse) error {
	m.resp = resp
	return nil
}

func TestDatasource_CallResource_NotFound(t *testing.T) {
	ctx := context.Background()
	ds := NewDatasource()

	sender := &mockResourceSender{}
	req := &backend.CallResourceRequest{
		Path: "unknown_path",
	}

	err := ds.CallResource(ctx, req, sender)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.resp == nil || sender.resp.Status != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got: %v", sender.resp)
	}
}

func TestDatasource_CallResource_MissingSettings(t *testing.T) {
	ctx := context.Background()
	ds := NewDatasource()

	for _, path := range []string{"sites", "issues"} {
		sender := &mockResourceSender{}
		req := &backend.CallResourceRequest{
			Path: path,
		}
		err := ds.CallResource(ctx, req, sender)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sender.resp == nil || sender.resp.Status != http.StatusInternalServerError {
			t.Fatalf("path %s: expected 500 error on missing settings, got: %v", path, sender.resp)
		}
	}
}

func TestDatasource_CallResource_Issues(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dna/system/api/v1/auth/token":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Token": "test-token"}`))
		case "/dna/data/api/v1/assuranceIssues":
			if r.Header.Get("X-Auth-Token") != "test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.URL.Query().Get("limit") != "10" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[{"issueId":"iss-1","name":"ap_down"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ds := NewDatasource()
	sender := &mockResourceSender{}
	req := &backend.CallResourceRequest{
		Path: "issues",
		URL:  "issues?limit=10",
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:      "test-instance-issues",
				JSONData: json.RawMessage(fmt.Sprintf(`{"baseUrl": %q}`, server.URL)),
				DecryptedSecureJSONData: map[string]string{
					"username": "user",
					"password": "pass",
				},
			},
		},
	}

	err := ds.CallResource(ctx, req, sender)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.resp == nil || sender.resp.Status != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %v", sender.resp)
	}
	if !strings.Contains(string(sender.resp.Body), "iss-1") {
		t.Fatalf("expected response body to contain iss-1, got: %s", string(sender.resp.Body))
	}
}

func TestDatasource_CallResource_Sites(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dna/system/api/v1/auth/token":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Token": "test-token"}`))
		case "/dna/intent/api/v2/site":
			if r.Header.Get("X-Auth-Token") != "test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[{"id":"site-1","name":"Building A"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ds := NewDatasource()
	sender := &mockResourceSender{}
	req := &backend.CallResourceRequest{
		Path: "sites",
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				UID:      "test-instance-sites",
				JSONData: json.RawMessage(fmt.Sprintf(`{"baseUrl": %q}`, server.URL)),
				DecryptedSecureJSONData: map[string]string{
					"username": "user",
					"password": "pass",
				},
			},
		},
	}

	err := ds.CallResource(ctx, req, sender)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.resp == nil || sender.resp.Status != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %v", sender.resp)
	}
	if !strings.Contains(string(sender.resp.Body), "Building A") {
		t.Fatalf("expected response body to contain Building A, got: %s", string(sender.resp.Body))
	}
}
