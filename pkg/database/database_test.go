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

func TestQueryBuilderDialects(t *testing.T) {
	// 1. PostgreSQL ($1, $2, ...)
	pgSQL, pgBindings := Table("products").
		SetDialect(PostgresDialect{}).
		Where("category", "elektronik").
		Where("price", "<=", 500000).
		Limit(5).
		Offset(10).
		ToSelectSQL()

	expectedPG := "SELECT * FROM products WHERE category = $1 AND price <= $2 LIMIT 5 OFFSET 10"
	if pgSQL != expectedPG {
		t.Fatalf("unexpected postgres sql:\nGot:  %s\nWant: %s", pgSQL, expectedPG)
	}
	if len(pgBindings) != 2 {
		t.Fatalf("unexpected pg bindings: %v", pgBindings)
	}

	// 2. SQL Server (@p1, @p2, OFFSET ROWS FETCH NEXT)
	msSQL, msBindings := Table("orders").
		SetDialect(SQLServerDialect{}).
		Where("status", "pending").
		OrderBy("order_date", "DESC").
		Limit(10).
		Offset(20).
		ToSelectSQL()

	expectedMS := "SELECT * FROM orders WHERE status = @p1 ORDER BY order_date DESC OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY"
	if msSQL != expectedMS {
		t.Fatalf("unexpected mssql sql:\nGot:  %s\nWant: %s", msSQL, expectedMS)
	}
	if len(msBindings) != 1 || msBindings[0] != "pending" {
		t.Fatalf("unexpected ms bindings: %v", msBindings)
	}

	// 3. MySQL
	mySQL, _ := Table("logs").
		SetDialect(MySQLDialect{}).
		Where("level", "error").
		Limit(15).
		Offset(30).
		ToSelectSQL()

	expectedMy := "SELECT * FROM logs WHERE level = ? LIMIT 30, 15"
	if mySQL != expectedMy {
		t.Fatalf("unexpected mysql sql:\nGot:  %s\nWant: %s", mySQL, expectedMy)
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

	// Update with Postgres Dialect ($1 = $1 ...)
	updateSQL, updateBindings := Table("users").
		SetDialect(PostgresDialect{}).
		Where("id", 5).
		ToUpdateSQL(map[string]any{
			"status": "blokir",
		})
	if !strings.HasPrefix(updateSQL, "UPDATE users SET status = $1 WHERE id = $2") {
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

func TestBlueprintDialects(t *testing.T) {
	// SQLite
	bpSQLite := NewBlueprint("articles", SQLiteDialect{})
	bpSQLite.ID()
	bpSQLite.String("title", 150)
	bpSQLite.Text("body")
	bpSQLite.Boolean("is_published")
	bpSQLite.Timestamps()

	sqlLite := bpSQLite.ToCreateTableSQL()
	if !strings.Contains(sqlLite, "id INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Fatalf("missing sqlite id: %s", sqlLite)
	}

	// PostgreSQL (BIGSERIAL)
	bpPG := NewBlueprint("articles", PostgresDialect{})
	bpPG.ID()
	bpPG.Boolean("is_published")
	bpPG.Timestamps()
	sqlPG := bpPG.ToCreateTableSQL()
	if !strings.Contains(sqlPG, "id BIGSERIAL PRIMARY KEY") {
		t.Fatalf("missing postgres bigserial: %s", sqlPG)
	}
	if !strings.Contains(sqlPG, "created_at TIMESTAMP WITH TIME ZONE") {
		t.Fatalf("missing postgres timestamp: %s", sqlPG)
	}

	// MySQL (BIGINT AUTO_INCREMENT)
	bpMy := NewBlueprint("articles", MySQLDialect{})
	bpMy.ID()
	bpMy.Boolean("is_published")
	sqlMy := bpMy.ToCreateTableSQL()
	if !strings.Contains(sqlMy, "id BIGINT AUTO_INCREMENT PRIMARY KEY") {
		t.Fatalf("missing mysql auto_increment: %s", sqlMy)
	}
	if !strings.Contains(sqlMy, "is_published TINYINT(1)") {
		t.Fatalf("missing mysql boolean: %s", sqlMy)
	}
}

func TestMultiDatabaseManager(t *testing.T) {
	mgr := NewManager()

	mgr.AddConnection("mysql_store", Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		Database: "toko_db",
		Username: "root",
	})

	mgr.AddConnection("pgsql_analytics", Config{
		Driver:   "postgres",
		Host:     "127.0.0.1",
		Port:     5432,
		Database: "analytics_db",
		Username: "postgres",
	})

	// Get mysql connection
	myConn, err := mgr.Connection("mysql_store")
	if err != nil {
		t.Fatalf("failed to get mysql connection: %v", err)
	}
	if myConn.Dialect().Name() != "mysql" {
		t.Fatalf("expected mysql dialect, got: %s", myConn.Dialect().Name())
	}

	// Table builder from mysql connection automatically uses MySQL dialect
	myQuery, _ := myConn.Table("transaksi").Where("total", ">", 1000).Limit(5).Offset(10).ToSelectSQL()
	if !strings.Contains(myQuery, "LIMIT 10, 5") {
		t.Fatalf("expected mysql limit offset in query: %s", myQuery)
	}

	// Get postgres connection
	pgConn, err := mgr.Connection("pgsql_analytics")
	if err != nil {
		t.Fatalf("failed to get postgres connection: %v", err)
	}
	if pgConn.Dialect().Name() != "postgres" {
		t.Fatalf("expected postgres dialect, got: %s", pgConn.Dialect().Name())
	}

	pgQuery, _ := pgConn.Table("events").Where("action", "click").ToSelectSQL()
	if !strings.Contains(pgQuery, "action = $1") {
		t.Fatalf("expected postgres $1 placeholder in query: %s", pgQuery)
	}
}

func TestQueryBuilderPaginateStructure(t *testing.T) {
	qb := Table("products").Where("status", "aktif")
	if qb.limitVal != -1 {
		t.Fatalf("expected default limitVal -1, got: %d", qb.limitVal)
	}
	qb.Limit(10).Offset(20)
	sql, _ := qb.ToSelectSQL()
	if !strings.Contains(sql, "LIMIT 10 OFFSET 20") {
		t.Fatalf("expected LIMIT 10 OFFSET 20 in query: %s", sql)
	}
}

func TestRelationships(t *testing.T) {
	conn := &Connection{dialect: SQLiteDialect{}}

	// Test HasMany
	user := map[string]any{"id": 42, "name": "Budi"}
	hasManyQB, err := HasMany(conn, "users", user, "articles", "", "")
	if err != nil {
		t.Fatalf("HasMany failed: %v", err)
	}
	sql, bindings := hasManyQB.ToSelectSQL()
	if !strings.Contains(sql, "WHERE user_id = ?") {
		t.Fatalf("expected user_id in sql, got: %s", sql)
	}
	if len(bindings) != 1 || bindings[0] != 42 {
		t.Fatalf("expected binding 42, got: %v", bindings)
	}

	// Test BelongsTo
	article := map[string]any{"id": 100, "user_id": 42, "title": "Larapak Relasi"}
	belongsToQB, err := BelongsTo(conn, article, "users", "", "")
	if err != nil {
		t.Fatalf("BelongsTo failed: %v", err)
	}
	bSQL, bBindings := belongsToQB.ToSelectSQL()
	if !strings.Contains(bSQL, "WHERE id = ?") {
		t.Fatalf("expected id = ? in sql, got: %s", bSQL)
	}
	if len(bBindings) != 1 || bBindings[0] != 42 {
		t.Fatalf("expected binding 42, got: %v", bBindings)
	}

	// Test HasOne
	hasOneQB, err := HasOne(conn, "users", user, "profiles", "", "")
	if err != nil {
		t.Fatalf("HasOne failed: %v", err)
	}
	hSQL, _ := hasOneQB.ToSelectSQL()
	if !strings.Contains(hSQL, "LIMIT 1") {
		t.Fatalf("expected LIMIT 1 in sql, got: %s", hSQL)
	}
}
