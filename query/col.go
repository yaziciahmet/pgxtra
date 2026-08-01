package query

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

type rawTable string

func (r rawTable) SQLName() string { return string(r) }
func (r rawTable) Alias() string   { return string(r) }

// Named returns a table reference by name (CTEs, ad-hoc).
func Named(name string) Table { return rawTable(name) }
