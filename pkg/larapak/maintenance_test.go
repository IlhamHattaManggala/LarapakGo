package larapak

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaintenanceMode(t *testing.T) {
	tempDir := t.TempDir()
	downFile := filepath.Join(tempDir, "down")

	// 1. Initial state: not down
	isDown, _ := IsDown(downFile)
	if isDown {
		t.Fatal("expected app to not be down initially")
	}

	// 2. Put into maintenance mode
	data, err := Down("secret123", "Sistem lagi di-update sedela.", 120, downFile)
	if err != nil {
		t.Fatalf("failed to put app down: %v", err)
	}
	if data.Secret != "secret123" {
		t.Fatalf("unexpected secret: %s", data.Secret)
	}

	isDown, readData := IsDown(downFile)
	if !isDown || readData.Message != "Sistem lagi di-update sedela." {
		t.Fatalf("expected app to be down, got isDown=%v, msg=%s", isDown, readData.Message)
	}

	// 3. Test HTTP request blocked with 503
	app := New()
	app.Use(Maintenance(downFile))
	app.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "sukses")
	})
	app.GET("/api/users", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]any{"users": []any{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "503 Sedang Diopeni") {
		t.Fatalf("expected 503 HTML template, got %s", w.Body.String())
	}

	// 4. Test API request blocked with 503 JSON
	reqAPI := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	wAPI := httptest.NewRecorder()
	app.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for API, got %d", wAPI.Code)
	}
	var jsonResp map[string]any
	if err := json.Unmarshal(wAPI.Body.Bytes(), &jsonResp); err != nil {
		t.Fatalf("expected json response: %v", err)
	}
	if jsonResp["retry_after"] != float64(120) {
		t.Fatalf("expected retry_after 120, got %v", jsonResp["retry_after"])
	}

	// 5. Test bypass with valid secret
	reqBypass := httptest.NewRequest(http.MethodGet, "/test?bypass=secret123", nil)
	wBypass := httptest.NewRecorder()
	app.ServeHTTP(wBypass, reqBypass)

	if wBypass.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on bypass, got %d", wBypass.Code)
	}
	if wBypass.Body.String() != "sukses" {
		t.Fatalf("expected sukses response on bypass, got %s", wBypass.Body.String())
	}

	// 6. Turn off maintenance mode (Up)
	if err := Up(downFile); err != nil {
		t.Fatalf("failed to lift maintenance mode: %v", err)
	}
	if isDown, _ := IsDown(downFile); isDown {
		t.Fatal("expected app to be up after Up()")
	}

	// 7. Regular request now succeeds
	wNormal := httptest.NewRecorder()
	app.ServeHTTP(wNormal, req)
	if wNormal.Code != http.StatusOK {
		t.Fatalf("expected 200 OK after Up(), got %d", wNormal.Code)
	}

	_ = os.RemoveAll(tempDir)
}
