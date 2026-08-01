package schema

// Database is the introspection IR consumed by codegen.
type Database struct {
	Schema string
	Enums  []Enum
	Tables []Table
}

// Enum is a postgres enum type.
type Enum struct {
	Name   string
	Values []string
}

// Table is a database table.
type Table struct {
	Name    string
	Columns []Column
}

// Column is a table column with mapped Go type.
type Column struct {
	Name     string
	PGType   string
	Nullable bool
	Go       GoType
}

// GoType is the Go type used for a column in generated code.
type GoType struct {
	Name       string // e.g. "int64", "uuid.UUID", "UserStatus", "[]string"
	ImportPath string // empty for predeclared/builtin packages
	IsEnum     bool
}
