package codegen

import "io/fs"

// Config controls code generation output.
type Config struct {
	Package       string // generated package name, e.g. "db"
	ImportPath    string // full module import path, e.g. "myapp/gen/db"
	PgxtraVersion string
	QueryFS       fs.FS  // embedded query source (preferred)
	QuerySource   string // filesystem path to query/ (dev override)
}

func (c Config) queryImport() string {
	return c.ImportPath + "/query"
}
