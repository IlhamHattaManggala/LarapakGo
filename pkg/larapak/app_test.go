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

func TestFormInputsAndRedirect(t *testing.T) {
	app := New()

	app.POST("/submit", func(c *Context) error {
		name := c.FormValue("name")
		if name == "admin" {
			return c.Redirect(http.StatusSeeOther, "/admin/dashboard")
		}
		inputs := c.AllInputs()
		return c.JSON(http.StatusOK, inputs)
	})

	app.SetNotFoundHandler(func(c *Context) error {
		return c.String(http.StatusNotFound, "Kustom 404: %s", c.Path())
	})

	// 1. Test Redirect
	req := httptest.NewRequest(http.MethodPost, "/submit?name=admin", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 See Other, got: %d", w.Code)
	}
	if w.Header().Get("Location") != "/admin/dashboard" {
		t.Fatalf("expected Location /admin/dashboard, got: %s", w.Header().Get("Location"))
	}

	// 2. Test Custom 404
	req404 := httptest.NewRequest(http.MethodGet, "/kaca-ilang", nil)
	w404 := httptest.NewRecorder()
	app.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got: %d", w404.Code)
	}
	if w404.Body.String() != "Kustom 404: /kaca-ilang" {
		t.Fatalf("unexpected 404 body: %s", w404.Body.String())
	}
}

