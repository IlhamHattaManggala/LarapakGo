package factory

import (
	"strings"
	"testing"
)

func TestFactoryGenerators(t *testing.T) {
	f := New()

	name := f.Name()
	if name == "" || !strings.Contains(name, " ") {
		t.Fatalf("unexpected name: %q", name)
	}

	email := f.Email("Budi Santoso")
	if !strings.Contains(email, "budi.santoso@") {
		t.Fatalf("unexpected email: %q", email)
	}

	phone := f.Phone()
	if !strings.HasPrefix(phone, "08") || len(phone) < 10 {
		t.Fatalf("unexpected phone: %q", phone)
	}

	addr := f.Address()
	if !strings.Contains(addr, "Jl.") {
		t.Fatalf("unexpected address: %q", addr)
	}

	city := f.City()
	if city == "" {
		t.Fatalf("expected non-empty city")
	}

	text := f.Text(5)
	words := strings.Fields(text)
	if len(words) != 5 {
		t.Fatalf("expected 5 words, got %d: %q", len(words), text)
	}

	num := f.Number(10, 20)
	if num < 10 || num > 20 {
		t.Fatalf("number out of range: %d", num)
	}

	price := f.Price(10000, 50000)
	if price < 10000 || price > 50000 || price%1000 != 0 {
		t.Fatalf("price invalid: %d", price)
	}

	date := f.Date()
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
		t.Fatalf("invalid date format: %q", date)
	}

	choice := f.Choice([]any{"aktif", "nonaktif", "pending"})
	if choice != "aktif" && choice != "nonaktif" && choice != "pending" {
		t.Fatalf("unexpected choice: %v", choice)
	}

	uuid := f.UUID()
	if len(uuid) != 36 || uuid[14] != '4' {
		t.Fatalf("unexpected uuid: %q", uuid)
	}
}
