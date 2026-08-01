package query

// Table identifies a database table. Codegen implements this for each table.
type Table interface {
	SQLName() string
	Alias() string
}

// Column identifies a table column. Codegen implements this for each column.
type Column interface {
	Qualifier() string
	SQLName() string
	Table() Table
}

type qualColumn struct {
	table  string
	column string
}

// Col references a qualified column outside generated schema (e.g. CTE columns).
func Col(table, column string) Column {
	return qualColumn{table: table, column: column}
}

func (q qualColumn) SQLName() string   { return q.column }
func (q qualColumn) Qualifier() string { return q.table + "." + q.column }
func (q qualColumn) Table() Table      { return rawTable(q.table) }

// RawColumn is a select-list expression with ? placeholders replaced by numbered
// args at compile time. Use NewRawCol; do not use in WHERE/HAVING/JOIN ON.
type RawColumn struct {
	sql  string
	args []any
}

// NewRawCol builds a raw SQL select-list entry. Include AS alias in sql when needed.
func NewRawCol(sql string, args ...any) *RawColumn {
	return &RawColumn{sql: sql, args: args}
}

func (r *RawColumn) SQLName() string   { return r.sql }
func (r *RawColumn) Qualifier() string { return r.sql }
func (r *RawColumn) Table() Table      { return nil }

func (r *RawColumn) compile(c *Compiler) {
	writeInterpolated(c, r.sql, r.args)
}

type rawTable string

func (r rawTable) SQLName() string { return string(r) }
func (r rawTable) Alias() string   { return string(r) }

// Named returns a table reference by name (CTEs, ad-hoc).
func Named(name string) Table { return rawTable(name) }
