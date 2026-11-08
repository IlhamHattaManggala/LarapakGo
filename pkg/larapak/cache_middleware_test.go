package larapak

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/cache"
)

func TestCacheResponseMiddleware(t *testing.T) {
	cache.Default().Flush()

	app := New()
	counter := 0

	app.GET("/api/cached-time", func(c *Context) error {
		counter++
		return c.JSON(http.StatusOK, map[string]any{
			"counter": counter,
			"message": "hello cache",
		})
	}, CacheResponse(2*time.Second))

	// 1st request -> Cache MISS, counter = 1
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/api/cached-time", nil)
	app.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w1.Code)
	}
	if w1.Header().Get("X-Larapak-Cache") != "MISS" {
		t.Fatalf("expected MISS on first request, got: %s", w1.Header().Get("X-Larapak-Cache"))
	}
	if counter != 1 {
		t.Fatalf("expected counter 1, got %d", counter)
	}

	// 2nd request -> Cache HIT, counter still 1
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/cached-time", nil)
	app.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200 on hit, got %d", w2.Code)
	}
	if w2.Header().Get("X-Larapak-Cache") != "HIT" {
		t.Fatalf("expected HIT on second request, got: %s", w2.Header().Get("X-Larapak-Cache"))
	}
	if counter != 1 {
		t.Fatalf("expected counter to remain 1 on hit, got %d", counter)
	}
	if w2.Body.String() != w1.Body.String() {
		t.Fatalf("expected cached body %q, got %q", w1.Body.String(), w2.Body.String())
	}

	// 3rd request with different query param -> Cache MISS
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/api/cached-time?param=1", nil)
	app.ServeHTTP(w3, req3)
	if w3.Header().Get("X-Larapak-Cache") != "MISS" {
		t.Fatalf("expected MISS on different query, got: %s", w3.Header().Get("X-Larapak-Cache"))
	}
	if counter != 2 {
		t.Fatalf("expected counter 2, got %d", counter)
	}

	_ = fmt.Sprintf("OK")
}
