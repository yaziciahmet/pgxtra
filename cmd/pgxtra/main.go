package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yaziciahmet/pgxtra/generate"
)

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
		fmt.Println(generate.Version)
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
	querySrc := fs.String("query-source", "", "path to internal/query (default: embedded builder)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dsn == "" {
		return fmt.Errorf("--dsn is required")
	}

	var excludeTables []string
	for _, t := range strings.Split(*exclude, ",") {
		if t = strings.TrimSpace(t); t != "" {
			excludeTables = append(excludeTables, t)
		}
	}

	return generate.Generate(context.Background(), generate.Options{
		DSN:           *dsn,
		OutDir:        *out,
		Schema:        *schemaName,
		ExcludeTables: excludeTables,
		QuerySource:   *querySrc,
	})
}
