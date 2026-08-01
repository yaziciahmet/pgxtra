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
