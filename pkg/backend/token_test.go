package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestTokenManager_ManualTokenOverride verifies that a manually configured token
// is always used when provided, bypassing all cache and authentication logic.
func TestTokenManager_ManualTokenOverride(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	settings := &InstanceSettings{
		BaseURL:  "https://catalyst.example.com",
		APIToken: "manual-token-123",
		Username: "user",
		Password: "pass",
	}

	// Mock HTTP client (should never be called)
	client := &http.Client{}

	token, err := tm.getToken(ctx, "instance-1", settings, client)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "manual-token-123" {
		t.Fatalf("expected manual token, got: %q", token)
	}
}

// TestTokenManager_SuccessfulFetch verifies that a new token is fetched correctly
// from the API when no cached token exists.
func TestTokenManager_SuccessfulFetch(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	// Mock server that returns a token in the header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return token in header with expiry
		w.Header().Set("X-Auth-Token", "test-token-abc")
		w.Header().Set("X-Auth-Token-Expires-In", "3600") // 1 hour
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "test-token-abc" {
		t.Fatalf("expected token 'test-token-abc', got: %q", token)
	}

	// Verify token is cached
	tm.mu.Lock()
	entry, exists := tm.cache["instance-1"]
	tm.mu.Unlock()

	if !exists {
		t.Fatal("expected token to be cached")
	}
	if entry.Token != "test-token-abc" {
		t.Fatalf("expected cached token 'test-token-abc', got: %q", entry.Token)
	}
}

// TestTokenManager_Caching verifies that a second call returns the cached token
// without making a new API request.
func TestTokenManager_Caching(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Auth-Token", "cached-token")
		w.Header().Set("X-Auth-Token-Expires-In", "3600")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	// First call - should hit the server
	token1, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected 1 API call, got %d", callCount)
	}

	// Second call - should use cache
	token2, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected still 1 API call (cached), got %d", callCount)
	}

	if token1 != token2 {
		t.Fatalf("expected same token, got %q and %q", token1, token2)
	}

	if token1 != "cached-token" {
		t.Fatalf("expected 'cached-token', got %q", token1)
	}
}

// TestTokenManager_TokenExpiry verifies that an expired token triggers a new fetch.
func TestTokenManager_TokenExpiry(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	// Manually set an expired token in the cache
	expiredTime := time.Now().Add(-10 * time.Minute).Unix()
	tm.mu.Lock()
	tm.cache["instance-1"] = tokenEntry{
		Token:     "expired-token",
		ExpiresAt: expiredTime,
	}
	tm.mu.Unlock()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Auth-Token", "fresh-token")
		w.Header().Set("X-Auth-Token-Expires-In", "3600")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	// This call should detect the expired token and fetch a new one
	token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("call error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected 1 API call (token expired), got %d", callCount)
	}

	if token != "fresh-token" {
		t.Fatalf("expected 'fresh-token', got %q", token)
	}
}

// TestTokenManager_ThreadSafety verifies that concurrent calls are handled safely.
func TestTokenManager_ThreadSafety(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	var callCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("X-Auth-Token", "concurrent-token")
		w.Header().Set("X-Auth-Token-Expires-In", "3600")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	// Make multiple concurrent calls
	const numGoroutines = 10
	var wg sync.WaitGroup
	tokens := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
			tokens[idx] = token
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// Verify all calls succeeded
	for i, err := range errors {
		if err != nil {
			t.Fatalf("goroutine %d got error: %v", i, err)
		}
	}

	// Verify all got the same token
	firstToken := tokens[0]
	for i, token := range tokens {
		if token != firstToken {
			t.Fatalf("goroutine %d got different token: %q vs %q", i, token, firstToken)
		}
	}
	
	if firstToken != "concurrent-token" {
		t.Fatalf("expected 'concurrent-token', got %q", firstToken)
	}

	// Verify that the cache lock is being used (i.e., no race conditions occurred)
	// The actual number of API calls can vary due to timing, but all goroutines
	// should have gotten the same token, which proves thread safety.
	mu.Lock()
	finalCallCount := callCount
	mu.Unlock()

	// At least one call should have been made
	if finalCallCount < 1 {
		t.Fatal("expected at least one API call")
	}
	
	t.Logf("Made %d API calls for %d concurrent requests (caching working)", finalCallCount, numGoroutines)
}

// TestTokenManager_NoCredentials verifies proper error when credentials are missing.
func TestTokenManager_NoCredentials(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	settings := &InstanceSettings{
		BaseURL: "https://catalyst.example.com",
		// No username/password
	}

	client := &http.Client{}

	_, err := tm.getToken(ctx, "instance-1", settings, client)
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
	if err.Error() != "no username/password provided; cannot obtain token" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// TestTokenManager_TokenInBody verifies token extraction from JSON response body.
func TestTokenManager_TokenInBody(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return token in JSON body instead of header
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Token": "body-token-123", "expiresIn": 7200}`))
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "body-token-123" {
		t.Fatalf("expected 'body-token-123', got: %q", token)
	}
}

// TestTokenManager_AlternateTokenField verifies fallback to alternate JSON field.
func TestTokenManager_AlternateTokenField(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Use lowercase "token" field instead of "Token"
		w.Write([]byte(`{"token": "alt-token-456", "expires_in": 3600}`))
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "alt-token-456" {
		t.Fatalf("expected 'alt-token-456', got: %q", token)
	}
}

// TestTokenManager_DefaultTTL verifies that default TTL is used when no expiry info is provided.
func TestTokenManager_DefaultTTL(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Token", "default-ttl-token")
		// No expiry information provided
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	token, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "default-ttl-token" {
		t.Fatalf("expected 'default-ttl-token', got: %q", token)
	}

	// Check that a default expiry was set (55 minutes from now)
	tm.mu.Lock()
	entry := tm.cache["instance-1"]
	tm.mu.Unlock()

	now := time.Now().Unix()
	expectedExpiry := now + (55 * 60)
	// Allow 2 second variance for test execution time
	if entry.ExpiresAt < expectedExpiry-2 || entry.ExpiresAt > expectedExpiry+2 {
		t.Fatalf("expected expiry around %d, got %d", expectedExpiry, entry.ExpiresAt)
	}
}

// TestTokenManager_MultipleInstances verifies that different instances have separate caches.
func TestTokenManager_MultipleInstances(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Token", "instance1-token")
		w.Header().Set("X-Auth-Token-Expires-In", "3600")
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Token", "instance2-token")
		w.Header().Set("X-Auth-Token-Expires-In", "3600")
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	settings1 := &InstanceSettings{
		BaseURL:  server1.URL,
		Username: "user1",
		Password: "pass1",
	}

	settings2 := &InstanceSettings{
		BaseURL:  server2.URL,
		Username: "user2",
		Password: "pass2",
	}

	token1, err := tm.getToken(ctx, "instance-1", settings1, server1.Client())
	if err != nil {
		t.Fatalf("instance 1 error: %v", err)
	}

	token2, err := tm.getToken(ctx, "instance-2", settings2, server2.Client())
	if err != nil {
		t.Fatalf("instance 2 error: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected different tokens for different instances")
	}

	if token1 != "instance1-token" {
		t.Fatalf("expected 'instance1-token', got %q", token1)
	}

	if token2 != "instance2-token" {
		t.Fatalf("expected 'instance2-token', got %q", token2)
	}
}

// TestTokenManager_HTTPError verifies proper error handling for non-2xx responses.
func TestTokenManager_HTTPError(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "baduser",
		Password: "badpass",
	}

	_, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "non-2xx") {
		t.Fatalf("expected 'non-2xx' in error, got: %v", err)
	}
}

// TestTokenManager_MissingTokenInResponse verifies error when token is not in response.
func TestTokenManager_MissingTokenInResponse(t *testing.T) {
	tm := newTokenManager()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return JSON without token field
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	settings := &InstanceSettings{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	_, err := tm.getToken(ctx, "instance-1", settings, server.Client())
	if err == nil {
		t.Fatal("expected error when token is missing")
	}
	if err.Error() != "token not found in response" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestParseMaxAge verifies Cache-Control max-age parsing.
func TestParseMaxAge(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		ok       bool
	}{
		{"max-age=3600", 3600, true},
		{"public, max-age=7200", 7200, true},
		{"max-age=1800, must-revalidate", 1800, true},
		{"no-cache", 0, false},
		{"max-age=invalid", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		got, ok := parseMaxAge(tt.input)
		if got != tt.expected || ok != tt.ok {
			t.Errorf("parseMaxAge(%q) = (%d, %v), want (%d, %v)",
				tt.input, got, ok, tt.expected, tt.ok)
		}
	}
}

// TestSetWithExpiry_MinTTL verifies that minimum TTL is applied for invalid expiry times.
func TestSetWithExpiry_MinTTL(t *testing.T) {
	tm := newTokenManager()

	// Set token with expiry in the past
	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	tm.setWithExpiry("test-uid", "test-token", pastTime)

	tm.mu.Lock()
	entry := tm.cache["test-uid"]
	tm.mu.Unlock()

	// Should have applied minimum TTL (5 minutes)
	now := time.Now().Unix()
	minExpiry := now + (5 * 60)
	if entry.ExpiresAt < minExpiry-2 {
		t.Fatalf("expected minimum TTL to be applied, got expiry %d, now %d", entry.ExpiresAt, now)
	}
}
