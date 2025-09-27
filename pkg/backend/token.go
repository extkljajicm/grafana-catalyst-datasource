package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// tokenManager handles the acquisition and caching of authentication tokens.
// It ensures that a valid token is available for API requests, refreshing it
// automatically when it expires. It supports both username/password credentials
// and manual token overrides. The cache is keyed by datasource instance UID
// to support multiple instances of the datasource.
type tokenManager struct {
	mu    sync.Mutex
	cache map[string]tokenEntry // key: instance UID
}

// newTokenManager creates a new token manager with an empty cache.
func newTokenManager() *tokenManager {
	return &tokenManager{
		cache: make(map[string]tokenEntry),
	}
}

// getToken retrieves a valid token for the given datasource instance.
// It follows this order of precedence:
//  1. Returns the manual API token from settings if provided.
//  2. Returns a valid, non-expired token from the cache.
//  3. If no valid token is found, it requests a new one using the provided
//     username and password, then caches it with its expiry time.
func (tm *tokenManager) getToken(ctx context.Context, instanceUID string, s *InstanceSettings, client *http.Client) (string, error) {
	// 1. Manual override: if the user has configured a specific token, always use it.
	if t := strings.TrimSpace(s.APIToken); t != "" {
		return t, nil
	}

	now := time.Now().Unix()

	// 2. Cache check: return a valid, non-expired token if one exists.
	tm.mu.Lock()
	if e, ok := tm.cache[instanceUID]; ok && now < e.ExpiresAt && strings.TrimSpace(e.Token) != "" {
		t := e.Token
		tm.mu.Unlock()
		return t, nil
	}
	tm.mu.Unlock()

	// 3. New token request: if no credentials, we can't proceed.
	if s.Username == "" || s.Password == "" {
		return "", errors.New("no username/password provided; cannot obtain token")
	}

	tokenURL, err := TokenURL(s.BaseURL)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.Username, s.Password)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("token endpoint returned non-2xx: " + resp.Status)
	}

	// 4. Token extraction: The token can be in a header or the response body.
	// Prefer the header if present.
	if tok := strings.TrimSpace(resp.Header.Get("X-Auth-Token")); tok != "" {
		if expAt, ok := parseExpiryFromHeaders(resp.Header); ok {
			tm.setWithExpiry(instanceUID, tok, expAt)
			return tok, nil
		}
		// Fallback to default TTL if headers don't specify expiry.
		tm.set(instanceUID, tok)
		return tok, nil
	}

	// 5. Fallback to body: The token and expiry hints can also be in the JSON body.
	var body struct {
		Token         string `json:"Token"`
		Token2        string `json:"token"`
		ExpiresIn     int64  `json:"expiresIn"`  // seconds
		ExpiresInAlt  int64  `json:"expires_in"` // seconds
		ExpiryEpoch   int64  `json:"expiry"`     // epoch seconds
		ExpiresAt     int64  `json:"expiresAt"`  // epoch seconds
		ExpireTimeRFC string `json:"expireTime"` // RFC3339 or RFC1123, if any
		Expiration    int64  `json:"expiration"` // seconds or epoch (varies by APIs)
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)

	tok := strings.TrimSpace(body.Token)
	if tok == "" {
		tok = strings.TrimSpace(body.Token2)
	}
	if tok == "" {
		log.DefaultLogger.Warn("DNAC token not found in header or JSON body")
		return "", errors.New("token not found in response")
	}

	// Prefer header-derived expiry if present; otherwise try JSON signals.
	if expAt, ok := parseExpiryFromHeaders(resp.Header); ok {
		tm.setWithExpiry(instanceUID, tok, expAt)
		return tok, nil
	}

	// Try common JSON fields for expiry.
	if expAt, ok := deriveExpiryFromJSON(body); ok {
		tm.setWithExpiry(instanceUID, tok, expAt)
		return tok, nil
	}

	// Last resort: if no expiry information is found, use a default TTL.
	tm.set(instanceUID, tok)
	return tok, nil
}

// set caches a token with a default TTL (Time To Live).
// This is used as a fallback when the API response doesn't provide expiry info.
func (tm *tokenManager) set(uid, token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Default TTL: 55 minutes, a safe duration for most token-based APIs.
	tm.cache[uid] = tokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(55 * time.Minute).Unix(),
	}
}

// setWithExpiry stores the token with an absolute expiry time (epoch seconds).
// If the provided expiry time is in the past or too close to the present,
// it applies a conservative minimum TTL to prevent caching an already-expired token.
func (tm *tokenManager) setWithExpiry(uid, token string, expAt int64) {
	const minTTL = 5 * time.Minute
	now := time.Now()
	if expAt <= now.Add(1*time.Minute).Unix() {
		// Guard: if server-provided expiry is missing, invalid, or too soon, default to a safe minimum TTL.
		expAt = now.Add(minTTL).Unix()
	}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cache[uid] = tokenEntry{
		Token:     token,
		ExpiresAt: expAt,
	}
}

// ---- helpers: expiry parsing ----

// parseExpiryFromHeaders attempts to determine the token's expiry time by inspecting
// various standard and non-standard HTTP headers. It checks for headers that
// specify a relative duration (e.g., Cache-Control: max-age) or an absolute time.
func parseExpiryFromHeaders(h http.Header) (int64, bool) {
	now := time.Now()

	// 1) Explicit “expires in seconds”
	for _, k := range []string{
		"X-Auth-Token-Expires-In", // hypothetical common name
		"X-Token-Expires-In",
	} {
		if v := strings.TrimSpace(h.Get(k)); v != "" {
			if sec, err := strconv.ParseInt(v, 10, 64); err == nil && sec > 0 {
				return now.Add(time.Duration(sec) * time.Second).Unix(), true
			}
		}
	}

	// 2) Explicit absolute epoch seconds
	for _, k := range []string{
		"X-Auth-Token-Expiry",
		"X-Token-Expiry",
	} {
		if v := strings.TrimSpace(h.Get(k)); v != "" {
			if epoch, err := strconv.ParseInt(v, 10, 64); err == nil && epoch > now.Unix() {
				return epoch, true
			}
		}
	}

	// 3) Cache-Control: max-age=NNNN
	if cc := h.Get("Cache-Control"); cc != "" {
		if secs, ok := parseMaxAge(cc); ok && secs > 0 {
			return now.Add(time.Duration(secs) * time.Second).Unix(), true
		}
	}

	// 4) Expires: HTTP-date (RFC7231)
	if exp := strings.TrimSpace(h.Get("Expires")); exp != "" {
		// Try common HTTP date formats
		for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.RFC850, time.ANSIC} {
			if t, err := time.Parse(layout, exp); err == nil {
				if t.After(now) {
					return t.Unix(), true
				}
			}
		}
	}

	return 0, false
}

// parseMaxAge extracts the 'max-age' value from a Cache-Control header string.
func parseMaxAge(cacheControl string) (int64, bool) {
	parts := strings.Split(cacheControl, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "max-age=") {
			val := strings.TrimSpace(p[len("max-age="):])
			if sec, err := strconv.ParseInt(val, 10, 64); err == nil {
				return sec, true
			}
		}
	}
	return 0, false
}

// deriveExpiryFromJSON attempts to determine the token's expiry time by inspecting
// various common fields in a JSON response body. It handles both relative durations
// (e.g., "expiresIn": 3600) and absolute timestamps.
func deriveExpiryFromJSON(body struct {
	Token         string `json:"Token"`
	Token2        string `json:"token"`
	ExpiresIn     int64  `json:"expiresIn"`
	ExpiresInAlt  int64  `json:"expires_in"`
	ExpiryEpoch   int64  `json:"expiry"`
	ExpiresAt     int64  `json:"expiresAt"`
	ExpireTimeRFC string `json:"expireTime"`
	Expiration    int64  `json:"expiration"`
}) (int64, bool) {
	now := time.Now()

	// seconds until expiry
	if body.ExpiresIn > 0 {
		return now.Add(time.Duration(body.ExpiresIn) * time.Second).Unix(), true
	}
	if body.ExpiresInAlt > 0 {
		return now.Add(time.Duration(body.ExpiresInAlt) * time.Second).Unix(), true
	}

	// absolute epoch seconds
	if body.ExpiresAt > now.Unix()+60 {
		return body.ExpiresAt, true
	}
	if body.ExpiryEpoch > now.Unix()+60 {
		return body.ExpiryEpoch, true
	}
	// Some APIs ambiguously name "expiration": treat > now as epoch, otherwise as seconds
	if body.Expiration > 0 {
		if body.Expiration > now.Unix()+60 {
			return body.Expiration, true
		}
		return now.Add(time.Duration(body.Expiration) * time.Second).Unix(), true
	}

	// RFC time string
	if ts := strings.TrimSpace(body.ExpireTimeRFC); ts != "" {
		for _, layout := range []string{time.RFC3339, time.RFC1123, time.RFC1123Z, time.RFC850, time.ANSIC} {
			if t, err := time.Parse(layout, ts); err == nil && t.After(now.Add(1*time.Minute)) {
				return t.Unix(), true
			}
		}
	}

	return 0, false
}
