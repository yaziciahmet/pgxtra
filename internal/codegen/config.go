package codegen

import (
	"io/fs"
	"path"
)

// Config controls code generation output.
type Config struct {
	Package     string // generated package name, e.g. "db"
	ImportPath  string // full module import path, e.g. "myapp/gen/db"
	TypgVersion string
	QueryFS     fs.FS  // embedded query source (preferred)
	QuerySource string // filesystem path to query/ (dev override)
}

func (c Config) queryImport() string {
	return c.ImportPath + "/query"
}

func (c Config) modelsPackage() string {
	return "models"
}

func (c Config) modelsImport() string {
	return path.Join(path.Dir(c.ImportPath), c.modelsPackage())
}
