package database

import (
	"fmt"
	"strings"
)

// Dialect defines database-specific SQL formatting, placeholders, and data types.
type Dialect interface {
	Name() string
	QuoteIdentifier(name string) string
	Placeholder(index int) string
	CompileLimitOffset(limit, offset int) string
	AutoIncrementPK() string
	BooleanType() string
	TimestampType() string
}

// -----------------------------------------------------------------------------
// 1. SQLite Dialect
// -----------------------------------------------------------------------------

type SQLiteDialect struct{}

func (d SQLiteDialect) Name() string { return "sqlite" }

func (d SQLiteDialect) QuoteIdentifier(name string) string {
	if name == "*" || strings.Contains(name, ".") {
		return name
	}
	return fmt.Sprintf("`%s`", name)
}

func (d SQLiteDialect) Placeholder(index int) string {
	return "?"
}

func (d SQLiteDialect) CompileLimitOffset(limit, offset int) string {
	var sb strings.Builder
	if limit >= 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", offset))
	}
	return sb.String()
}

func (d SQLiteDialect) AutoIncrementPK() string {
	return "INTEGER PRIMARY KEY AUTOINCREMENT"
}

func (d SQLiteDialect) BooleanType() string {
	return "BOOLEAN"
}

func (d SQLiteDialect) TimestampType() string {
	return "DATETIME"
}

// -----------------------------------------------------------------------------
// 2. MySQL / MariaDB Dialect
// -----------------------------------------------------------------------------

type MySQLDialect struct{}

func (d MySQLDialect) Name() string { return "mysql" }

func (d MySQLDialect) QuoteIdentifier(name string) string {
	if name == "*" || strings.Contains(name, ".") {
		return name
	}
	return fmt.Sprintf("`%s`", name)
}

func (d MySQLDialect) Placeholder(index int) string {
	return "?"
}

func (d MySQLDialect) CompileLimitOffset(limit, offset int) string {
	var sb strings.Builder
	if limit >= 0 && offset > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d, %d", offset, limit))
	} else if limit >= 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	} else if offset > 0 {
		// In MySQL, offset requires limit
		sb.WriteString(fmt.Sprintf(" LIMIT %d, 18446744073709551615", offset))
	}
	return sb.String()
}

func (d MySQLDialect) AutoIncrementPK() string {
	return "BIGINT AUTO_INCREMENT PRIMARY KEY"
}

func (d MySQLDialect) BooleanType() string {
	return "TINYINT(1)"
}

func (d MySQLDialect) TimestampType() string {
	return "TIMESTAMP"
}

// -----------------------------------------------------------------------------
// 3. PostgreSQL Dialect
// -----------------------------------------------------------------------------

type PostgresDialect struct{}

func (d PostgresDialect) Name() string { return "postgres" }

func (d PostgresDialect) QuoteIdentifier(name string) string {
	if name == "*" || strings.Contains(name, ".") {
		return name
	}
	return fmt.Sprintf("\"%s\"", name)
}

func (d PostgresDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d PostgresDialect) CompileLimitOffset(limit, offset int) string {
	var sb strings.Builder
	if limit >= 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", offset))
	}
	return sb.String()
}

func (d PostgresDialect) AutoIncrementPK() string {
	return "BIGSERIAL PRIMARY KEY"
}

func (d PostgresDialect) BooleanType() string {
	return "BOOLEAN"
}

func (d PostgresDialect) TimestampType() string {
	return "TIMESTAMP WITH TIME ZONE"
}

// -----------------------------------------------------------------------------
// 4. SQL Server (MSSQL) Dialect
// -----------------------------------------------------------------------------

type SQLServerDialect struct{}

func (d SQLServerDialect) Name() string { return "sqlserver" }

func (d SQLServerDialect) QuoteIdentifier(name string) string {
	if name == "*" || strings.Contains(name, ".") {
		return name
	}
	return fmt.Sprintf("[%s]", name)
}

func (d SQLServerDialect) Placeholder(index int) string {
	return fmt.Sprintf("@p%d", index)
}

func (d SQLServerDialect) CompileLimitOffset(limit, offset int) string {
	if limit < 0 && offset <= 0 {
		return ""
	}
	off := offset
	if off < 0 {
		off = 0
	}
	if limit >= 0 {
		return fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", off, limit)
	}
	return fmt.Sprintf(" OFFSET %d ROWS", off)
}

func (d SQLServerDialect) AutoIncrementPK() string {
	return "BIGINT IDENTITY(1,1) PRIMARY KEY"
}

func (d SQLServerDialect) BooleanType() string {
	return "BIT"
}

func (d SQLServerDialect) TimestampType() string {
	return "DATETIME2"
}

// -----------------------------------------------------------------------------
// Dialect Resolver
// -----------------------------------------------------------------------------

// GetDialect returns the appropriate SQL dialect based on driver name.
// Supported: "sqlite", "sqlite3", "mysql", "mariadb", "postgres", "pgsql", "sqlserver", "mssql".
func GetDialect(driver string) Dialect {
	switch strings.ToLower(driver) {
	case "mysql", "mariadb":
		return MySQLDialect{}
	case "postgres", "postgresql", "pgsql":
		return PostgresDialect{}
	case "sqlserver", "mssql":
		return SQLServerDialect{}
	default:
		return SQLiteDialect{}
	}
}

