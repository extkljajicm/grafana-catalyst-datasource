package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	sdklog "github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type tokenManager struct {
	mu     sync.Mutex
	cache  map[string]tokenEntry // key: instance UID
	client *http.Client
}

func newTokenManager() *tokenManager {
	return &tokenManager{
		cache:  make(map[string]tokenEntry),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (tm *tokenManager) getToken(ctx context.Context, instanceUID string, s *InstanceSettings) (string, error) {
	// Manual override
	if t := strings.TrimSpace(s.APIToken); t != "" {
		return t, nil
	}

	now := time.Now()

	tm.mu.Lock()
	if e, ok := tm.cache[instanceUID]; ok && now.Before(e.ExpiresAt) && e.Token != "" {
		t := e.Token
		tm.mu.Unlock()
		return t, nil
	}
	tm.mu.Unlock()

	// Need to fetch
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
	// DNA Center token endpoint expects Basic Auth
	req.SetBasicAuth(s.Username, s.Password)

	resp, err := tm.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("token endpoint returned non-2xx: " + resp.Status)
	}

	// DNAC often returns token in header "X-Auth-Token"
	if v := resp.Header.Get("X-Auth-Token"); strings.TrimSpace(v) != "" {
		tm.set(instanceUID, v)
		return v, nil
	}

	// Some versions return {"Token": "..."} or {"token": "..."}
	var body struct {
		Token string `json:"Token"`
		Token2 string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	token := strings.TrimSpace(body.Token)
	if token == "" {
		token = strings.TrimSpace(body.Token2)
	}
	if token == "" {
		sdklog.DefaultLogger.Warn("DNAC token not found in header or JSON body")
		return "", errors.New("token not found in response")
	}

	tm.set(instanceUID, token)
	return token, nil
}

func (tm *tokenManager) set(uid, token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Soft TTL 55 minutes
	tm.cache[uid] = tokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(55 * time.Minute),
	}
}
