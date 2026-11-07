package studio

import (
	"fmt"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/database"
)

// TableSummary holds metadata about an inspected database table.
type TableSummary struct {
	Name string `json:"name"`
	Rows int64  `json:"rows"`
}

// TableData holds records and column headers for a specific table.
type TableData struct {
	Name    string           `json:"name"`
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Total   int64            `json:"total"`
}

// InspectTables queries the active database connection to list available user tables and row counts.
func InspectTables() ([]TableSummary, error) {
	conn, err := database.GetManager().Connection()
	if err != nil {
		return nil, err
	}
	db := conn.DB()
	if db == nil {
		return []TableSummary{}, nil
	}

	dialectName := conn.Dialect().Name()
	var query string
	switch dialectName {
	case "mysql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()"
	case "postgres":
		query = "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'"
	default: // sqlite & others
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE '_litestream%'"
	}

	rows, err := db.Query(query)
	if err != nil {
		return []TableSummary{}, nil
	}
	defer rows.Close()

	var tables []TableSummary
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			var count int64
			quote := conn.Dialect().QuoteIdentifier(name)
			_ = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", quote)).Scan(&count)
			tables = append(tables, TableSummary{
				Name: name,
				Rows: count,
			})
		}
	}

	return tables, nil
}

// InspectTableData retrieves column names and up to 50 rows for the specified table.
func InspectTableData(tableName string, limit, offset int) (TableData, error) {
	conn, err := database.GetManager().Connection()
	if err != nil {
		return TableData{Name: tableName}, err
	}
	db := conn.DB()
	if db == nil {
		return TableData{Name: tableName}, fmt.Errorf("database connection ora aktif")
	}

	// Basic safety sanitization
	safeTable := strings.ReplaceAll(tableName, ";", "")
	safeTable = strings.ReplaceAll(safeTable, " ", "")
	safeTable = strings.ReplaceAll(safeTable, "--", "")

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	quote := conn.Dialect().QuoteIdentifier(safeTable)
	var total int64
	_ = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", quote)).Scan(&total)

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", quote, limit, offset)
	rows, err := db.Query(query)
	if err != nil {
		return TableData{Name: tableName, Total: total}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return TableData{Name: tableName, Total: total}, err
	}

	resultRows := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err == nil {
			rowMap := make(map[string]any, len(columns))
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					rowMap[col] = string(b)
				} else {
					rowMap[col] = val
				}
			}
			resultRows = append(resultRows, rowMap)
		}
	}

	return TableData{
		Name:    safeTable,
		Columns: columns,
		Rows:    resultRows,
		Total:   total,
	}, nil
}
