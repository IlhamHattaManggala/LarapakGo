package security

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRFManager(t *testing.T) {
	csrf := NewCSRFManager()

	token, err := csrf.GenerateToken()
	if err != nil || len(token) != 64 {
		t.Fatalf("failed to generate token: %v, len: %d", err, len(token))
	}

	if !csrf.Validate(token, token) {
		t.Fatal("expected matching tokens to be valid")
	}

	if csrf.Validate(token, "invalid_token_string") {
		t.Fatal("expected invalid token to fail")
	}

	if !csrf.IsSafeMethod("GET") || !csrf.IsSafeMethod("HEAD") {
		t.Fatal("expected GET/HEAD to be safe")
	}

	if csrf.IsSafeMethod("POST") || csrf.IsSafeMethod("DELETE") {
		t.Fatal("expected POST/DELETE to not be safe")
	}

	// Test header extraction
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("X-CSRF-TOKEN", token)
	if csrf.ExtractTokenFromRequest(req) != token {
		t.Fatal("failed to extract token from header")
	}

	// Test form extraction
	form := url.Values{}
	form.Set("_token", token)
	reqForm := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	reqForm.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = reqForm.ParseForm()
	if csrf.ExtractTokenFromRequest(reqForm) != token {
		t.Fatal("failed to extract token from form")
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "RahasiaNgapak123!"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !VerifyPassword(password, hashed) {
		t.Fatal("expected correct password to verify successfully")
	}

	if VerifyPassword("SalahPassword", hashed) {
		t.Fatal("expected wrong password to fail verification")
	}

	if NeedsRehash(hashed) {
		t.Fatal("freshly generated hash should not need rehash")
	}
}

