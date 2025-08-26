package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

	// Prefer header
	if v := resp.Header.Get("X-Auth-Token"); strings.TrimSpace(v) != "" {
		tm.set(instanceUID, v)
		return v, nil
	}

	// Fallback to body
	var body struct {
		Token  string `json:"Token"`
		Token2 string `json:"token"`
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

	tm.set(instanceUID, tok)
	return tok, nil
}

func (tm *tokenManager) set(uid, token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// TTL: 55 minutes
	tm.cache[uid] = tokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(55 * time.Minute).Unix(),
	}
}
