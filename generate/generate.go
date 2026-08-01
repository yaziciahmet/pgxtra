package generate

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaziciahmet/typg/internal/codegen"
	"github.com/yaziciahmet/typg/internal/gomod"
	"github.com/yaziciahmet/typg/internal/introspect"
	"github.com/yaziciahmet/typg/internal/query"
)

// Options configures schema introspection and code generation.
type Options struct {
	// DSN is a postgres connection string. Ignored when Pool is set.
	DSN string

	// Pool is an existing postgres pool. When nil, DSN must be set.
	Pool *pgxpool.Pool

	// OutDir is the output directory for the generated db package.
	// Models are written to a sibling models/ directory.
	OutDir string

	// Schema is the postgres schema to introspect. Default: public.
	Schema string

	// ExcludeTables skips tables by name.
	ExcludeTables []string

	// Package is the generated db package name. Default: basename of OutDir.
	Package string

	// ImportPath is the full Go import path for the generated db package.
	// Default: derived from the nearest go.mod relative to OutDir.
	ImportPath string

	// QuerySource overrides the embedded query builder (for typg development).
	QuerySource string

	// Version is stamped into generated file headers. Default: generate.Version.
	Version string
}

// Generate introspects postgres and writes generated db + models packages to OutDir.
func Generate(ctx context.Context, opts Options) error {
	if opts.OutDir == "" {
		return fmt.Errorf("generate: OutDir is required")
	}
	if opts.Pool == nil && opts.DSN == "" {
		return fmt.Errorf("generate: DSN or Pool is required")
	}

	pool := opts.Pool
	if pool == nil {
		var err error
		pool, err = pgxpool.New(ctx, opts.DSN)
		if err != nil {
			return fmt.Errorf("generate: connect: %w", err)
		}
		defer pool.Close()
	}

	db, err := introspect.Postgres(ctx, pool, introspect.Options{
		Schema:        opts.Schema,
		ExcludeTables: opts.ExcludeTables,
	})
	if err != nil {
		return fmt.Errorf("generate: introspect: %w", err)
	}

	cfg, err := codegenConfig(opts)
	if err != nil {
		return err
	}

	files, err := codegen.Generate(cfg, db)
	if err != nil {
		return fmt.Errorf("generate: codegen: %w", err)
	}
	if err := codegen.WriteFiles(opts.OutDir, files); err != nil {
		return fmt.Errorf("generate: write: %w", err)
	}
	return nil
}

func codegenConfig(opts Options) (codegen.Config, error) {
	importPath := opts.ImportPath
	if importPath == "" {
		var err error
		importPath, err = gomod.ModuleImportPath(opts.OutDir)
		if err != nil {
			return codegen.Config{}, fmt.Errorf("generate: import path: %w", err)
		}
	}

	pkg := opts.Package
	if pkg == "" {
		pkg = filepath.Base(filepath.Clean(opts.OutDir))
		if pkg == "" || pkg == "." {
			pkg = "db"
		}
	}

	version := opts.Version
	if version == "" {
		version = Version
	}

	cfg := codegen.Config{
		Package:       pkg,
		ImportPath:    importPath,
		TypgVersion: version,
	}
	if opts.QuerySource != "" {
		cfg.QuerySource = opts.QuerySource
	} else {
		cfg.QueryFS = query.EmbeddedFS
	}
	return cfg, nil
}
