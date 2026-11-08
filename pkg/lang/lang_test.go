package lang

import (
	"testing"
)

func TestTranslator(t *testing.T) {
	trans := New("id", "en")

	idJSON := []byte(`{
		"welcome": "Sugeng rawuh, :name!",
		"auth": {
			"failed": "Kredensial ora cocog.",
			"throttle": "Keliwat kerep nyoba. Tulung ngenteni :seconds detik."
		}
	}`)

	enJSON := []byte(`{
		"welcome": "Welcome, :name!",
		"auth": {
			"failed": "These credentials do not match our records."
		},
		"goodbye": "Goodbye!"
	}`)

	if err := trans.LoadJSON("id", idJSON); err != nil {
		t.Fatalf("failed to load id json: %v", err)
	}
	if err := trans.LoadJSON("en", enJSON); err != nil {
		t.Fatalf("failed to load en json: %v", err)
	}

	// 1. Direct translation with placeholder map
	res1 := trans.Trans("welcome", map[string]any{"name": "Slamet"})
	if res1 != "Sugeng rawuh, Slamet!" {
		t.Fatalf("unexpected translation: got %q", res1)
	}

	// 2. Nested key translation
	res2 := trans.Trans("auth.failed")
	if res2 != "Kredensial ora cocog." {
		t.Fatalf("unexpected nested translation: got %q", res2)
	}

	// 3. Fallback translation (key exists only in en)
	res3 := trans.Trans("goodbye")
	if res3 != "Goodbye!" {
		t.Fatalf("unexpected fallback translation: got %q", res3)
	}

	// 4. Missing key returns key itself
	res4 := trans.Trans("unknown.key")
	if res4 != "unknown.key" {
		t.Fatalf("expected raw key for unknown: got %q", res4)
	}

	// 5. Variadic key-value pairs
	res5 := trans.Trans("auth.throttle", "seconds", 30)
	if res5 != "Keliwat kerep nyoba. Tulung ngenteni 30 detik." {
		t.Fatalf("unexpected variadic translation: got %q", res5)
	}

	// 6. Switch locale
	trans.SetLocale("en")
	if trans.GetLocale() != "en" {
		t.Fatalf("expected locale en, got %s", trans.GetLocale())
	}
	res6 := trans.Trans("welcome", "name", "John")
	if res6 != "Welcome, John!" {
		t.Fatalf("unexpected translation in en: got %q", res6)
	}
}
