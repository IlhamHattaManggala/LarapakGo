package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionManager(t *testing.T) {
	mgr := NewManager("kunci_rahasia_banyumas", 2*time.Hour)

	// 1. Initial request creates new session and sets cookie
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()

	sess1 := mgr.Start(w1, req1)
	sess1.Set("user_id", 101)
	sess1.Flash("success", "Pendaftaran sukses!")

	cookies := w1.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "larapak_session" {
		t.Fatalf("expected session cookie, got %v", cookies)
	}

	sessionCookie := cookies[0]

	// 2. Second request reuses same session via cookie
	req2 := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()

	sess2 := mgr.Start(w2, req2)
	val, ok := sess2.Get("user_id")
	if !ok || val.(int) != 101 {
		t.Fatalf("expected user_id 101, got %v", val)
	}

	// 3. Flash message consumed
	flash := sess2.GetFlash("success")
	if flash != "Pendaftaran sukses!" {
		t.Fatalf("expected flash message, got %s", flash)
	}

	// 4. Flash message should now be empty
	flashAfter := sess2.GetFlash("success")
	if flashAfter != "" {
		t.Fatalf("expected flash to be consumed, got %s", flashAfter)
	}
}

