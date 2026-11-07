package larapak

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestThrottleMiddleware(t *testing.T) {
	app := New()
	// Limit: 2 requests per 500ms
	app.Use(Throttle(2, 500*time.Millisecond))

	app.GET("/test-throttle", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// 1st request -> 200
	req1 := httptest.NewRequest("GET", "/test-throttle", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on 1st request, got %d", w1.Code)
	}

	// 2nd request -> 200
	req2 := httptest.NewRequest("GET", "/test-throttle", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on 2nd request, got %d", w2.Code)
	}

	// 3rd request -> 429
	req3 := httptest.NewRequest("GET", "/test-throttle", nil)
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 3rd request, got %d", w3.Code)
	}

	if w3.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header on 429")
	}
}
