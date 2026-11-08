package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// QueryExecutor defines the database interface needed for QueryBuilder execution.
type QueryExecutor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

// QueryBuilder provides fluent SQL query construction with parameter binding and dialect support.
type QueryBuilder struct {
	table       string
	columns     []string
	wheres      []whereClause
	joins       []joinClause
	orderBys    []string
	limitVal    int
	offsetVal   int
	bindings    []any
	dialect     Dialect
	executor    QueryExecutor
	softDeletes bool
	withTrashed bool
	onlyTrashed bool
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

// Table initiates a new QueryBuilder instance for a specified table name with default SQLite dialect.
func Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table:    name,
		columns:  []string{"*"},
		limitVal: -1,
		dialect:  SQLiteDialect{},
	}
}

// SetDialect overrides the active SQL dialect for this query.
func (q *QueryBuilder) SetDialect(d Dialect) *QueryBuilder {
	if d != nil {
		q.dialect = d
	}
	return q
}

// Dialect returns the active SQL dialect.
func (q *QueryBuilder) Dialect() Dialect {
	if q.dialect == nil {
		q.dialect = SQLiteDialect{}
	}
	return q.dialect
}

// TableName returns the target database table name.
func (q *QueryBuilder) TableName() string {
	return q.table
}

// SetExecutor assigns an active database connection or transaction executor.
func (q *QueryBuilder) SetExecutor(e QueryExecutor) *QueryBuilder {
	q.executor = e
	return q
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

// SoftDeletes enables soft-delete filtering and marking for this query.
func (q *QueryBuilder) SoftDeletes() *QueryBuilder {
	q.softDeletes = true
	return q
}

// MbusakAlus is the Ngapak alias for SoftDeletes.
func (q *QueryBuilder) MbusakAlus() *QueryBuilder {
	return q.SoftDeletes()
}

// WithTrashed includes soft-deleted records in query results.
func (q *QueryBuilder) WithTrashed() *QueryBuilder {
	q.withTrashed = true
	return q
}

// KaroDibusek is the Ngapak alias for WithTrashed.
func (q *QueryBuilder) KaroDibusek() *QueryBuilder {
	return q.WithTrashed()
}

// OnlyTrashed limits results to soft-deleted records only.
func (q *QueryBuilder) OnlyTrashed() *QueryBuilder {
	q.onlyTrashed = true
	return q
}

// MungDibusek is the Ngapak alias for OnlyTrashed.
func (q *QueryBuilder) MungDibusek() *QueryBuilder {
	return q.OnlyTrashed()
}

// ToSelectSQL compiles the builder into a parameterized SELECT query and its bound arguments.
func (q *QueryBuilder) ToSelectSQL() (string, []any) {
	d := q.Dialect()
	var sb strings.Builder
	var bindings []any
	paramIdx := 1

	sb.WriteString(fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.columns, ", "), q.table))

	// Joins
	for _, j := range q.joins {
		sb.WriteString(fmt.Sprintf(" %s JOIN %s ON %s %s %s", j.joinType, j.table, j.first, j.operator, j.second))
	}

	// Prepare effective WHERE clauses including soft-delete filters
	effectiveWheres := append([]whereClause(nil), q.wheres...)
	if q.softDeletes {
		if q.onlyTrashed {
			effectiveWheres = append(effectiveWheres, whereClause{boolean: "AND", column: "deleted_at", operator: "IS NOT NULL"})
		} else if !q.withTrashed {
			effectiveWheres = append(effectiveWheres, whereClause{boolean: "AND", column: "deleted_at", operator: "IS NULL"})
		}
	}

	// Wheres
	if len(effectiveWheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range effectiveWheres {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(" %s ", w.boolean))
			}
			if w.operator == "IS NULL" || w.operator == "IS NOT NULL" {
				sb.WriteString(fmt.Sprintf("%s %s", w.column, w.operator))
			} else {
				ph := d.Placeholder(paramIdx)
				paramIdx++
				sb.WriteString(fmt.Sprintf("%s %s %s", w.column, w.operator, ph))
				bindings = append(bindings, w.value)
			}
		}
	}

	// OrderBy
	if len(q.orderBys) > 0 {
		sb.WriteString(" ORDER BY " + strings.Join(q.orderBys, ", "))
	}

	// Limit & Offset per dialect
	limitClause := d.CompileLimitOffset(q.limitVal, q.offsetVal)
	if limitClause != "" {
		sb.WriteString(limitClause)
	}

	return sb.String(), bindings
}

// ToInsertSQL compiles an INSERT statement.
func (q *QueryBuilder) ToInsertSQL(data map[string]any) (string, []any) {
	d := q.Dialect()
	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	bindings := make([]any, 0, len(data))
	paramIdx := 1

	for k, v := range data {
		cols = append(cols, k)
		placeholders = append(placeholders, d.Placeholder(paramIdx))
		paramIdx++
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
	d := q.Dialect()
	setClauses := make([]string, 0, len(data))
	bindings := make([]any, 0, len(data)+len(q.wheres))
	paramIdx := 1

	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", k, d.Placeholder(paramIdx)))
		paramIdx++
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
			ph := d.Placeholder(paramIdx)
			paramIdx++
			sb.WriteString(fmt.Sprintf("%s %s %s", w.column, w.operator, ph))
			bindings = append(bindings, w.value)
		}
	}

	return sb.String(), bindings
}

// ToDeleteSQL compiles a DELETE statement based on active WHERE clauses.
func (q *QueryBuilder) ToDeleteSQL() (string, []any) {
	d := q.Dialect()
	var sb strings.Builder
	var bindings []any
	paramIdx := 1

	sb.WriteString(fmt.Sprintf("DELETE FROM %s", q.table))

	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(" %s ", w.boolean))
			}
			ph := d.Placeholder(paramIdx)
			paramIdx++
			sb.WriteString(fmt.Sprintf("%s %s %s", w.column, w.operator, ph))
			bindings = append(bindings, w.value)
		}
	}

	return sb.String(), bindings
}

// -----------------------------------------------------------------------------
// Live Database Execution Methods
// -----------------------------------------------------------------------------

func (q *QueryBuilder) checkExecutor() error {
	if q.executor == nil {
		return errors.New("database: no active database connection assigned to query builder")
	}
	return nil
}

// Get executes the compiled SELECT query and returns all matching rows as maps.
func (q *QueryBuilder) Get() ([]map[string]any, error) {
	if err := q.checkExecutor(); err != nil {
		return nil, err
	}

	query, args := q.ToSelectSQL()
	start := time.Now()
	rows, err := q.executor.Query(query, args...)
	GetProfiler().Record(query, args, time.Since(start), err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*any)
			// Unwrap byte slices into strings for cleaner JSON output
			if b, ok := (*val).([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = *val
			}
		}
		results = append(results, rowMap)
	}

	return results, rows.Err()
}

// First executes the compiled query with LIMIT 1 and returns the first row.
func (q *QueryBuilder) First() (map[string]any, error) {
	q.Limit(1)
	results, err := q.Get()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return results[0], nil
}

// Count returns the total number of matching rows.
func (q *QueryBuilder) Count() (int64, error) {
	if err := q.checkExecutor(); err != nil {
		return 0, err
	}
	origCols := q.columns
	origLimit := q.limitVal
	origOffset := q.offsetVal
	q.columns = []string{"COUNT(*) as count"}
	q.limitVal = -1
	q.offsetVal = 0
	defer func() {
		q.columns = origCols
		q.limitVal = origLimit
		q.offsetVal = origOffset
	}()

	query, args := q.ToSelectSQL()
	start := time.Now()
	row := q.executor.QueryRow(query, args...)

	var count int64
	err := row.Scan(&count)
	GetProfiler().Record(query, args, time.Since(start), err)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Paginate executes query with page and perPage parameters, returning a standardized pagination result map.
func (q *QueryBuilder) Paginate(page, perPage int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		lastPage = 1
	}

	offset := (page - 1) * perPage
	q.Limit(perPage)
	q.Offset(offset)

	data, err := q.Get()
	if err != nil {
		return nil, err
	}

	from := offset + 1
	to := offset + len(data)
	if len(data) == 0 {
		from = 0
		to = 0
	}

	return map[string]any{
		"data":         data,
		"total":        total,
		"per_page":     perPage,
		"current_page": page,
		"last_page":    lastPage,
		"from":         from,
		"to":           to,
		"has_more":     page < lastPage,
	}, nil
}

// Insert executes an INSERT statement with the provided key-value map.
func (q *QueryBuilder) Insert(data map[string]any) (sql.Result, error) {
	if err := q.checkExecutor(); err != nil {
		return nil, err
	}
	query, args := q.ToInsertSQL(data)
	start := time.Now()
	res, err := q.executor.Exec(query, args...)
	GetProfiler().Record(query, args, time.Since(start), err)
	return res, err
}

// Update executes an UPDATE statement with the provided key-value map.
func (q *QueryBuilder) Update(data map[string]any) (sql.Result, error) {
	if err := q.checkExecutor(); err != nil {
		return nil, err
	}
	query, args := q.ToUpdateSQL(data)
	start := time.Now()
	res, err := q.executor.Exec(query, args...)
	GetProfiler().Record(query, args, time.Since(start), err)
	return res, err
}

// Delete executes a DELETE statement, or marks deleted_at timestamp if softDeletes is enabled.
func (q *QueryBuilder) Delete() (sql.Result, error) {
	if q.softDeletes {
		return q.Update(map[string]any{
			"deleted_at": time.Now().Format("2006-01-02 15:04:05"),
		})
	}
	if err := q.checkExecutor(); err != nil {
		return nil, err
	}
	query, args := q.ToDeleteSQL()
	start := time.Now()
	res, err := q.executor.Exec(query, args...)
	GetProfiler().Record(query, args, time.Since(start), err)
	return res, err
}

// ForceDelete permanently removes records from the database using raw SQL DELETE.
func (q *QueryBuilder) ForceDelete() (sql.Result, error) {
	if err := q.checkExecutor(); err != nil {
		return nil, err
	}
	query, args := q.ToDeleteSQL()
	start := time.Now()
	res, err := q.executor.Exec(query, args...)
	GetProfiler().Record(query, args, time.Since(start), err)
	return res, err
}

// BusakPermanen is the Ngapak alias for ForceDelete.
func (q *QueryBuilder) BusakPermanen() (sql.Result, error) {
	return q.ForceDelete()
}

// Restore restores soft-deleted records by setting deleted_at to NULL.
func (q *QueryBuilder) Restore() (sql.Result, error) {
	return q.Update(map[string]any{
		"deleted_at": nil,
	})
}

// Balekna is the Ngapak alias for Restore.
func (q *QueryBuilder) Balekna() (sql.Result, error) {
	return q.Restore()
}
