package studio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStudioRegistrationAndOpenAPI(t *testing.T) {
	st := New("Toko Larapak", "1.0.0", "local", "http://localhost:8080")

	st.RegisterRoute("GET", "/", "HomeController@index")
	st.RegisterRoute("GET", "/users/:id", "UserController@show")
	st.RegisterRoute("POST", "/users", "UserController@store")

	// Test HTML output
	html := st.RenderHTML()
	if !strings.Contains(html, "LarapakGo Route &amp; API Studio") {
		t.Fatal("missing studio brand title in HTML")
	}
	if !strings.Contains(html, "HomeController@index") {
		t.Fatal("missing registered route in HTML")
	}

	// Test OpenAPI output
	openApi := st.GenerateOpenAPI()
	if openApi["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi version: %v", openApi["openapi"])
	}

	paths := openApi["paths"].(map[string]any)
	if _, ok := paths["/users/{id}"]; !ok {
		t.Fatalf("expected /users/{id} path in OpenAPI, got paths: %v", paths)
	}

	// Test HTTP Handler
	req := httptest.NewRequest(http.MethodGet, "/_studio/openapi.json", nil)
	w := httptest.NewRecorder()
	st.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from openapi handler, got %d", w.Code)
	}

	var jsonSpec map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &jsonSpec); err != nil {
		t.Fatalf("invalid json from openapi handler: %v", err)
	}

	// Test Profiler queries endpoint
	reqQ := httptest.NewRequest(http.MethodGet, "/_studio/api/queries", nil)
	wQ := httptest.NewRecorder()
	st.Handler().ServeHTTP(wQ, reqQ)
	if wQ.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from queries handler, got %d", wQ.Code)
	}

	var jsonQueries map[string]any
	if err := json.Unmarshal(wQ.Body.Bytes(), &jsonQueries); err != nil {
		t.Fatalf("invalid json from queries handler: %v", err)
	}
	if _, ok := jsonQueries["queries"]; !ok {
		t.Fatal("expected 'queries' field in response")
	}
}
