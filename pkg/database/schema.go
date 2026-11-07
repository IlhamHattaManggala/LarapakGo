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

// Blueprint builds table schema definitions with multi-dialect support.
type Blueprint struct {
	TableName string
	Columns   []ColumnDefinition
	dialect   Dialect
}

// NewBlueprint creates a new Blueprint for the given table.
func NewBlueprint(table string, dialect ...Dialect) *Blueprint {
	d := Dialect(SQLiteDialect{})
	if len(dialect) > 0 && dialect[0] != nil {
		d = dialect[0]
	}
	return &Blueprint{
		TableName: table,
		Columns:   make([]ColumnDefinition, 0),
		dialect:   d,
	}
}

// SetDialect overrides the active dialect for this blueprint.
func (b *Blueprint) SetDialect(d Dialect) *Blueprint {
	if d != nil {
		b.dialect = d
	}
	return b
}

// ID creates an auto-incrementing primary key ID column.
func (b *Blueprint) ID() *Blueprint {
	b.Columns = append(b.Columns, ColumnDefinition{
		Name:      "id",
		Type:      "ID",
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
		ColumnDefinition{Name: "created_at", Type: "TIMESTAMP", Nullable: true},
		ColumnDefinition{Name: "updated_at", Type: "TIMESTAMP", Nullable: true},
	)
	return b
}

// ToCreateTableSQL compiles the blueprint into dialect-appropriate CREATE TABLE SQL.
func (b *Blueprint) ToCreateTableSQL() string {
	d := b.dialect
	if d == nil {
		d = SQLiteDialect{}
	}

	colDefs := make([]string, 0, len(b.Columns))

	for _, col := range b.Columns {
		var def string

		if col.AutoIncr && col.IsPrimary {
			def = fmt.Sprintf("%s %s", col.Name, d.AutoIncrementPK())
		} else {
			typeName := col.Type
			switch col.Type {
			case "BOOLEAN":
				typeName = d.BooleanType()
			case "TIMESTAMP":
				typeName = d.TimestampType()
			}

			def = fmt.Sprintf("%s %s", col.Name, typeName)
			if col.IsPrimary {
				def += " PRIMARY KEY"
			}
			if !col.Nullable && !col.IsPrimary {
				def += " NOT NULL"
			}
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
