package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// SiteTranslator provides a thread-safe, caching mechanism for resolving
// site IDs to site names and vice-versa. It fetches all sites from the API
// and builds an in-memory lookup table to avoid repeated API calls.
type SiteTranslator struct {
	httpClient *http.Client
	inst       *dsInstance
	tm         *tokenManager

	// Mutex to protect concurrent access to the cache
	mu sync.RWMutex
	// Caches for quick lookups
	idToSite map[string]Site // Changed from idToName
	nameToID map[string]string
	allSites []Site
	// lastRefresh tracks when the cache was last populated.
	lastRefresh time.Time
	// cacheTTL defines how long the cache is considered valid.
	cacheTTL time.Duration
}

// NewSiteTranslator creates a new translator instance.
func NewSiteTranslator(httpClient *http.Client, inst *dsInstance, tm *tokenManager) *SiteTranslator {
	return &SiteTranslator{
		httpClient:  httpClient,
		inst:        inst,
		tm:          tm,
		idToSite:    make(map[string]Site), // Changed from idToName
		nameToID:    make(map[string]string),
		allSites:    make([]Site, 0),
		cacheTTL:    5 * time.Minute, // Cache is valid for 5 minutes
		lastRefresh: time.Time{},     // Zero time indicates never refreshed
	}
}

// EnsureCache ensures that the site cache is populated and not stale.
// If the cache is empty or has expired, it triggers a refresh.
func (st *SiteTranslator) EnsureCache(ctx context.Context) error {
	st.mu.RLock()
	isStale := time.Since(st.lastRefresh) > st.cacheTTL
	st.mu.RUnlock()

	if isStale {
		log.DefaultLogger.Info("Site cache is stale or empty, refreshing...")
		return st.refresh(ctx)
	}
	return nil
}

// refresh fetches all sites from the Catalyst Center API and rebuilds the cache.
func (st *SiteTranslator) refresh(ctx context.Context) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Double-check if another goroutine refreshed while we were waiting for the lock.
	if time.Since(st.lastRefresh) < st.cacheTTL {
		return nil
	}

	siteURL, err := SiteURL(st.inst.Settings.BaseURL)
	if err != nil {
		return fmt.Errorf("bad site baseUrl: %w", err)
	}

	token, err := st.tm.getToken(ctx, st.inst.UID, st.inst.Settings, st.httpClient)
	if err != nil {
		return fmt.Errorf("token for site lookup: %w", err)
	}

	var allSites []Site
	limit := 500 // Fetch 500 sites per page as per API docs
	offset := 1  // API is 1-based

	for {
		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))
		reqURL := siteURL + "?" + params.Encode()

		httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		httpReq.Header.Set("X-Auth-Token", token)
		httpReq.Header.Set("Accept", "application/json")

		httpResp, err := st.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("site request failed on page offset %d: %w", offset, err)
		}

		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			body, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
			return fmt.Errorf("site endpoint returned %s: %s", httpResp.Status, string(body))
		}

		var envelope SiteEnvelope
		if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
			httpResp.Body.Close()
			return fmt.Errorf("failed to decode site response: %w", err)
		}
		httpResp.Body.Close()

		allSites = append(allSites, envelope.Response...)

		if len(envelope.Response) < limit {
			break // This was the last page
		}
		offset += limit
	}

	// Rebuild the caches
	idToSite := make(map[string]Site) // Changed from idToName
	nameToID := make(map[string]string)
	for _, site := range allSites {
		if site.ID != "" && site.Name != "" {
			idToSite[site.ID] = site // Store the whole site object
			// Use lowercase for the name-to-ID map for case-insensitive lookups
			nameToID[strings.ToLower(site.Name)] = site.ID
		}
	}

	st.idToSite = idToSite // Changed from idToName
	st.nameToID = nameToID
	st.allSites = allSites
	st.lastRefresh = time.Now()

	log.DefaultLogger.Info("Site cache refreshed successfully", "sites_loaded", len(allSites))
	return nil
}

// GetSiteByID retrieves a full Site object by its ID.
func (st *SiteTranslator) GetSiteByID(ctx context.Context, id string) (Site, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if site, ok := st.idToSite[id]; ok {
		return site, nil
	}
	return Site{}, fmt.Errorf("site with ID '%s' not found in cache", id)
}

// GetSiteName resolves a site ID to its name using the cache.
func (st *SiteTranslator) GetSiteName(ctx context.Context, id string) (string, error) {
	site, err := st.GetSiteByID(ctx, id)
	if err != nil {
		return "", err
	}
	return site.Name, nil
}

// GetSiteID resolves a site name to its ID using the cache (case-insensitive).
func (st *SiteTranslator) GetSiteID(ctx context.Context, name string) (string, error) {
	if err := st.EnsureCache(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure site cache: %w", err)
	}
	st.mu.RLock()
	defer st.mu.RUnlock()
	if id, ok := st.nameToID[strings.ToLower(name)]; ok {
		return id, nil
	}
	return "", fmt.Errorf("site with name '%s' not found in cache", name)
}

// GetAllSites returns a copy of all sites from the cache.
func (st *SiteTranslator) GetAllSites(ctx context.Context) ([]Site, error) {
	if err := st.EnsureCache(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure site cache: %w", err)
	}
	st.mu.RLock()
	defer st.mu.RUnlock()
	// Return a copy to prevent modification of the internal slice
	sitesCopy := make([]Site, len(st.allSites))
	copy(sitesCopy, st.allSites)
	return sitesCopy, nil
}
