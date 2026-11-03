package studio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStudioRegistrationAndOpenAPI(t *testing.T) {
	st := New("Toko Larapak", "1.2.2", "local", "http://localhost:8080")

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
}
