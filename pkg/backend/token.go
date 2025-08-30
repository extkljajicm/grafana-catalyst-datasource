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

type tokenManager struct {
	mu    sync.Mutex
	cache map[string]tokenEntry // key: instance UID
}

func newTokenManager() *tokenManager {
	return &tokenManager{
		cache: make(map[string]tokenEntry),
	}
}

func (tm *tokenManager) getToken(ctx context.Context, instanceUID string, s *InstanceSettings, client *http.Client) (string, error) {
	// manual override
	if t := strings.TrimSpace(s.APIToken); t != "" {
		return t, nil
	}

	now := time.Now().Unix()

	tm.mu.Lock()
	if e, ok := tm.cache[instanceUID]; ok && now < e.ExpiresAt && strings.TrimSpace(e.Token) != "" {
		t := e.Token
		tm.mu.Unlock()
		return t, nil
	}
	tm.mu.Unlock()

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

	// Try to parse expiry from headers (optional)
	if tok := strings.TrimSpace(resp.Header.Get("X-Auth-Token")); tok != "" {
		if expAt, ok := parseExpiryFromHeaders(resp.Header); ok {
			tm.setWithExpiry(instanceUID, tok, expAt)
			return tok, nil
		}
		// Fallback to default if headers don’t tell us
		tm.set(instanceUID, tok)
		return tok, nil
	}

	// Fallback to body: token + optional expiry hints
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

	// Prefer header-derived expiry if present; otherwise try JSON signals
	if expAt, ok := parseExpiryFromHeaders(resp.Header); ok {
		tm.setWithExpiry(instanceUID, tok, expAt)
		return tok, nil
	}

	// Try common JSON fields
	if expAt, ok := deriveExpiryFromJSON(body); ok {
		tm.setWithExpiry(instanceUID, tok, expAt)
		return tok, nil
	}

	// Last resort: default TTL
	tm.set(instanceUID, tok)
	return tok, nil
}

func (tm *tokenManager) set(uid, token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Default TTL: 55 minutes
	tm.cache[uid] = tokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(55 * time.Minute).Unix(),
	}
}

// setWithExpiry stores the token with an absolute expiry time (epoch seconds).
// If expAt is in the past or too close, apply a conservative min TTL.
func (tm *tokenManager) setWithExpiry(uid, token string, expAt int64) {
	const minTTL = 5 * time.Minute
	now := time.Now()
	if expAt <= now.Add(1*time.Minute).Unix() {
		// Guard: if server-provided expiry is missing/invalid/too soon, default to min TTL
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
