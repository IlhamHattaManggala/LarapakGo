package larapak

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterAndContext(t *testing.T) {
	app := New()

	app.GET("/", func(c *Context) error {
		return c.String(http.StatusOK, "Selamat Teka neng LarapakGo!")
	})

	app.GET("/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]string{
			"user_id": id,
			"status":  "aktif",
		})
	})

	app.POST("/echo", func(c *Context) error {
		var body map[string]any
		if err := c.BindJSON(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, body)
	})

	// Test GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "Selamat Teka neng LarapakGo!" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}

	// Test GET /users/42
	req = httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if res["user_id"] != "42" || res["status"] != "aktif" {
		t.Fatalf("unexpected json content: %v", res)
	}

	// Test 404 Not Found
	req = httptest.NewRequest(http.MethodGet, "/ora-ana", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRouteGrouping(t *testing.T) {
	app := New()

	app.Group("/api/v1", func(r *Router) {
		r.GET("/status", func(c *Context) error {
			return c.JSON(http.StatusOK, map[string]string{"system": "ok"})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on grouped route, got %d", w.Code)
	}
}

