package validation

import (
	"testing"
)

func TestValidatorPasses(t *testing.T) {
	data := map[string]any{
		"nama":               "Ilham",
		"email":              "ilham@larapak.my.id",
		"sandi":              "rahasia123",
		"sandi_confirmation": "rahasia123",
		"umur":               "25",
		"peran":              "admin",
	}

	rules := map[string]string{
		"nama":  "required|alpha|min:3",
		"email": "required|email",
		"sandi": "required|min:8|confirmed",
		"umur":  "required|numeric",
		"peran": "required|in:admin,user,editor",
	}

	v := Make(data, rules)
	if !v.Passes() {
		t.Fatalf("expected validation to pass, got errors: %v", v.Errors())
	}
}

func TestValidatorFails(t *testing.T) {
	data := map[string]any{
		"email": "bukan-email",
		"umur":  "sepuluh",
	}

	rules := map[string]string{
		"nama":  "required",
		"email": "required|email",
		"umur":  "numeric",
	}

	v := Make(data, rules)
	if v.Passes() {
		t.Fatal("expected validation to fail")
	}

	errs := v.Errors()
	if len(errs["nama"]) == 0 {
		t.Fatal("expected required error for nama")
	}
	if len(errs["email"]) == 0 {
		t.Fatal("expected email error for email")
	}
	if len(errs["umur"]) == 0 {
		t.Fatal("expected numeric error for umur")
	}
}
