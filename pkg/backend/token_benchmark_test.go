package backend

import (
	"net/http"
	"testing"
	"time"
)

func BenchmarkDeriveExpiryFromJSON_RFC(b *testing.B) {
	body := struct {
		Token         string `json:"Token"`
		Token2        string `json:"token"`
		ExpiresIn     int64  `json:"expiresIn"`
		ExpiresInAlt  int64  `json:"expires_in"`
		ExpiryEpoch   int64  `json:"expiry"`
		ExpiresAt     int64  `json:"expiresAt"`
		ExpireTimeRFC string `json:"expireTime"`
		Expiration    int64  `json:"expiration"`
	}{
		ExpireTimeRFC: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = deriveExpiryFromJSON(body)
	}
}

func BenchmarkParseExpiryFromHeaders_ExpiresHeader(b *testing.B) {
	h := http.Header{}
	h.Set("Expires", time.Now().Add(2*time.Hour).Format(time.RFC1123))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = parseExpiryFromHeaders(h)
	}
}
