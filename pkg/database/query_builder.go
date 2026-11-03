package database

import (
	"fmt"
	"strings"
)

// QueryBuilder provides fluent SQL query construction with parameter binding.
type QueryBuilder struct {
	table      string
	columns    []string
	wheres     []whereClause
	joins      []joinClause
	orderBys   []string
	limitVal   int
	offsetVal  int
	bindings   []any
}

type whereClause struct {
	boolean  string // "AND" or "OR"
	column   string
	operator string
	value    any
}

type joinClause struct {
	joinType string // "INNER", "LEFT", "RIGHT"
	table    string
	first    string
	operator string
	second   string
}

// Table initiates a new QueryBuilder instance for a specified table name.
func Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table:    name,
		columns:  []string{"*"},
		limitVal: -1,
	}
}

// Select specifies columns to retrieve.
func (q *QueryBuilder) Select(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		q.columns = columns
	}
	return q
}

// Where adds an AND WHERE condition.
// Syntax: Where("age", ">=", 18) or Where("status", "aktif") (defaults to "=")
func (q *QueryBuilder) Where(column string, args ...any) *QueryBuilder {
	return q.addWhere("AND", column, args...)
}

// OrWhere adds an OR WHERE condition.
func (q *QueryBuilder) OrWhere(column string, args ...any) *QueryBuilder {
	return q.addWhere("OR", column, args...)
}

func (q *QueryBuilder) addWhere(boolean, column string, args ...any) *QueryBuilder {
	op := "="
	var val any

	if len(args) == 1 {
		val = args[0]
	} else if len(args) >= 2 {
		op = fmt.Sprintf("%v", args[0])
		val = args[1]
	}

	q.wheres = append(q.wheres, whereClause{
		boolean:  boolean,
		column:   column,
		operator: op,
		value:    val,
	})
	return q
}

// Join adds an INNER JOIN clause.
func (q *QueryBuilder) Join(table, first, operator, second string) *QueryBuilder {
	q.joins = append(q.joins, joinClause{
		joinType: "INNER",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	return q
}

// LeftJoin adds a LEFT JOIN clause.
func (q *QueryBuilder) LeftJoin(table, first, operator, second string) *QueryBuilder {
	q.joins = append(q.joins, joinClause{
		joinType: "LEFT",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	return q
}

// OrderBy adds an ORDER BY clause.
func (q *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	dir := strings.ToUpper(direction)
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	q.orderBys = append(q.orderBys, fmt.Sprintf("%s %s", column, dir))
	return q
}

// Limit sets the maximum number of rows.
func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.limitVal = limit
	return q
}

// Offset sets the row skip offset.
func (q *QueryBuilder) Offset(offset int) *QueryBuilder {
	q.offsetVal = offset
	return q
}

// ToSelectSQL compiles the builder into a parameterized SELECT query and its bound arguments.
func (q *QueryBuilder) ToSelectSQL() (string, []any) {
	var sb strings.Builder
	var bindings []any

	sb.WriteString(fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.columns, ", "), q.table))

	// Joins
	for _, j := range q.joins {
		sb.WriteString(fmt.Sprintf(" %s JOIN %s ON %s %s %s", j.joinType, j.table, j.first, j.operator, j.second))
	}

	// Wheres
	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(" %s ", w.boolean))
			}
			sb.WriteString(fmt.Sprintf("%s %s ?", w.column, w.operator))
			bindings = append(bindings, w.value)
		}
	}

	// OrderBy
	if len(q.orderBys) > 0 {
		sb.WriteString(" ORDER BY " + strings.Join(q.orderBys, ", "))
	}

	// Limit & Offset
	if q.limitVal >= 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", q.limitVal))
	}
	if q.offsetVal > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", q.offsetVal))
	}

	return sb.String(), bindings
}

// ToInsertSQL compiles an INSERT statement.
func (q *QueryBuilder) ToInsertSQL(data map[string]any) (string, []any) {
	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	bindings := make([]any, 0, len(data))

	for k, v := range data {
		cols = append(cols, k)
		placeholders = append(placeholders, "?")
		bindings = append(bindings, v)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, bindings
}

// ToUpdateSQL compiles an UPDATE statement based on active WHERE clauses.
func (q *QueryBuilder) ToUpdateSQL(data map[string]any) (string, []any) {
	setClauses := make([]string, 0, len(data))
	bindings := make([]any, 0, len(data)+len(q.wheres))

	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", k))
		bindings = append(bindings, v)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(setClauses, ", ")))

	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(" %s ", w.boolean))
			}
			sb.WriteString(fmt.Sprintf("%s %s ?", w.column, w.operator))
			bindings = append(bindings, w.value)
		}
	}

	return sb.String(), bindings
}

// ToDeleteSQL compiles a DELETE statement based on active WHERE clauses.
func (q *QueryBuilder) ToDeleteSQL() (string, []any) {
	var sb strings.Builder
	var bindings []any

	sb.WriteString(fmt.Sprintf("DELETE FROM %s", q.table))

	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(" %s ", w.boolean))
			}
			sb.WriteString(fmt.Sprintf("%s %s ?", w.column, w.operator))
			bindings = append(bindings, w.value)
		}
	}

	return sb.String(), bindings
}

