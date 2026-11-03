package database

import (
	"strings"
	"testing"
)

func TestQueryBuilderSelect(t *testing.T) {
	sql, bindings := Table("users").
		Select("id", "name", "email").
		Where("status", "aktif").
		Where("age", ">=", 21).
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(20).
		ToSelectSQL()

	expectedSQL := "SELECT id, name, email FROM users WHERE status = ? AND age >= ? ORDER BY created_at DESC LIMIT 10 OFFSET 20"
	if sql != expectedSQL {
		t.Fatalf("unexpected sql:\nGot:  %s\nWant: %s", sql, expectedSQL)
	}

	if len(bindings) != 2 || bindings[0] != "aktif" || bindings[1] != 21 {
		t.Fatalf("unexpected bindings: %v", bindings)
	}
}

func TestQueryBuilderInsertUpdateDelete(t *testing.T) {
	// Insert
	insertSQL, insertBindings := Table("users").ToInsertSQL(map[string]any{
		"name":  "Ilham Hatta",
		"email": "ilham@larapak.my.id",
	})
	if !strings.HasPrefix(insertSQL, "INSERT INTO users") || len(insertBindings) != 2 {
		t.Fatalf("unexpected insert sql: %s", insertSQL)
	}

	// Update
	updateSQL, updateBindings := Table("users").Where("id", 5).ToUpdateSQL(map[string]any{
		"status": "blokir",
	})
	if !strings.HasPrefix(updateSQL, "UPDATE users SET status = ? WHERE id = ?") {
		t.Fatalf("unexpected update sql: %s", updateSQL)
	}
	if len(updateBindings) != 2 || updateBindings[0] != "blokir" || updateBindings[1] != 5 {
		t.Fatalf("unexpected update bindings: %v", updateBindings)
	}

	// Delete
	deleteSQL, deleteBindings := Table("users").Where("id", 5).ToDeleteSQL()
	expectedDelete := "DELETE FROM users WHERE id = ?"
	if deleteSQL != expectedDelete || len(deleteBindings) != 1 || deleteBindings[0] != 5 {
		t.Fatalf("unexpected delete sql: %s", deleteSQL)
	}
}

func TestBlueprint(t *testing.T) {
	bp := NewBlueprint("articles")
	bp.ID()
	bp.String("title", 150)
	bp.Text("body")
	bp.Boolean("is_published")
	bp.Timestamps()

	createSQL := bp.ToCreateTableSQL()
	if !strings.Contains(createSQL, "CREATE TABLE IF NOT EXISTS articles") {
		t.Fatalf("failed create table prefix: %s", createSQL)
	}
	if !strings.Contains(createSQL, "id INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Fatalf("missing id column: %s", createSQL)
	}
	if !strings.Contains(createSQL, "title VARCHAR(150) NOT NULL") {
		t.Fatalf("missing title column: %s", createSQL)
	}
}

