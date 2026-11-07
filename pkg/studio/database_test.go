package studio

import (
	"testing"
)

func TestInspectTables(t *testing.T) {
	tables, err := InspectTables()
	if err != nil {
		t.Fatalf("InspectTables failed: %v", err)
	}
	// Even if database has no tables yet, it should return slice without error
	if tables == nil {
		t.Fatalf("expected non-nil tables slice")
	}
}

func TestInspectTableData(t *testing.T) {
	data, err := InspectTableData("users", 10, 0)
	// Even if table doesn't exist yet, it shouldn't panic
	if data.Name != "users" {
		t.Fatalf("expected table name 'users', got %s", data.Name)
	}
	_ = err
}
