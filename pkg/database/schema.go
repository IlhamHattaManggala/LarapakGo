package database

import (
	"fmt"
	"strings"
)

// ColumnDefinition defines an individual database table column.
type ColumnDefinition struct {
	Name       string
	Type       string
	Nullable   bool
	IsPrimary  bool
	AutoIncr   bool
	DefaultVal any
}

// Blueprint builds table schema definitions.
type Blueprint struct {
	TableName string
	Columns   []ColumnDefinition
}

// NewBlueprint creates a new Blueprint for the given table.
func NewBlueprint(table string) *Blueprint {
	return &Blueprint{
		TableName: table,
		Columns:   make([]ColumnDefinition, 0),
	}
}

// ID creates an auto-incrementing primary key ID column.
func (b *Blueprint) ID() *Blueprint {
	b.Columns = append(b.Columns, ColumnDefinition{
		Name:      "id",
		Type:      "INTEGER",
		IsPrimary: true,
		AutoIncr:  true,
	})
	return b
}

// String creates a VARCHAR column.
func (b *Blueprint) String(name string, length ...int) *Blueprint {
	lenVal := 255
	if len(length) > 0 && length[0] > 0 {
		lenVal = length[0]
	}
	b.Columns = append(b.Columns, ColumnDefinition{
		Name: name,
		Type: fmt.Sprintf("VARCHAR(%d)", lenVal),
	})
	return b
}

// Integer creates an INTEGER column.
func (b *Blueprint) Integer(name string) *Blueprint {
	b.Columns = append(b.Columns, ColumnDefinition{
		Name: name,
		Type: "INTEGER",
	})
	return b
}

// Text creates a TEXT column.
func (b *Blueprint) Text(name string) *Blueprint {
	b.Columns = append(b.Columns, ColumnDefinition{
		Name: name,
		Type: "TEXT",
	})
	return b
}

// Boolean creates a BOOLEAN column.
func (b *Blueprint) Boolean(name string) *Blueprint {
	b.Columns = append(b.Columns, ColumnDefinition{
		Name: name,
		Type: "BOOLEAN",
	})
	return b
}

// Timestamps adds standard created_at and updated_at datetime columns.
func (b *Blueprint) Timestamps() *Blueprint {
	b.Columns = append(b.Columns,
		ColumnDefinition{Name: "created_at", Type: "DATETIME", Nullable: true},
		ColumnDefinition{Name: "updated_at", Type: "DATETIME", Nullable: true},
	)
	return b
}

// ToCreateTableSQL compiles the blueprint into standard CREATE TABLE SQL.
func (b *Blueprint) ToCreateTableSQL() string {
	colDefs := make([]string, 0, len(b.Columns))

	for _, col := range b.Columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.IsPrimary {
			def += " PRIMARY KEY"
		}
		if col.AutoIncr {
			def += " AUTOINCREMENT"
		}
		if !col.Nullable && !col.IsPrimary {
			def += " NOT NULL"
		}
		colDefs = append(colDefs, def)
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);",
		b.TableName,
		strings.Join(colDefs, ",\n  "),
	)
}

// ToDropTableSQL compiles a DROP TABLE IF EXISTS statement.
func ToDropTableSQL(table string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
}

