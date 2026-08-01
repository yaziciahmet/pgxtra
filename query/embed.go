package query

import "embed"

// EmbeddedFS is the builder source copied into generated packages.
//
//go:embed compile.go delete.go insert.go predicates.go schema.go select.go update.go with.go
var EmbeddedFS embed.FS
