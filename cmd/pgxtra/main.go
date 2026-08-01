package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaziciahmet/pgxtra/internal/codegen"
	"github.com/yaziciahmet/pgxtra/internal/gomod"
	"github.com/yaziciahmet/pgxtra/internal/introspect"
	"github.com/yaziciahmet/pgxtra/query"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "generate":
		if err := runGenerate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pgxtra: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println(version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: pgxtra generate --dsn=... --out=...\n")
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	dsn := fs.String("dsn", "", "postgres connection string")
	out := fs.String("out", "./gen/db", "output directory")
	schemaName := fs.String("schema", "public", "postgres schema")
	exclude := fs.String("exclude-tables", "", "comma-separated tables to skip")
	querySrc := fs.String("query-source", "", "path to pgxtra/query (default: embedded builder)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dsn == "" {
		return fmt.Errorf("--dsn is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	var excludeTables []string
	for _, t := range strings.Split(*exclude, ",") {
		if t = strings.TrimSpace(t); t != "" {
			excludeTables = append(excludeTables, t)
		}
	}

	db, err := introspect.Postgres(ctx, pool, introspect.Options{
		Schema:        *schemaName,
		ExcludeTables: excludeTables,
	})
	if err != nil {
		return err
	}

	importPath, err := gomod.ModuleImportPath(*out)
	if err != nil {
		return err
	}
	pkg := filepath.Base(filepath.Clean(*out))
	if pkg == "" || pkg == "." {
		pkg = "db"
	}

	cfg := codegen.Config{
		Package:       pkg,
		ImportPath:    importPath,
		PgxtraVersion: version,
	}
	if *querySrc != "" {
		cfg.QuerySource = *querySrc
	} else {
		cfg.QueryFS = query.EmbeddedFS
	}

	files, err := codegen.Generate(cfg, db)
	if err != nil {
		return err
	}
	return codegen.WriteFiles(*out, files)
}
