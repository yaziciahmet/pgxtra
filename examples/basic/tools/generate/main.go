// Command generate applies schema.sql to postgres and regenerates gen/db + gen/models.
//
// Usage:
//
//	docker run -d --name pgxtra-basic -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=user \
//	  -e POSTGRES_DB=blog -p 5434:5432 postgres:16-alpine
//	psql postgres://user:pass@localhost:5434/blog -f schema.sql
//	DATABASE_URL='postgres://user:pass@localhost:5434/blog?sslmode=disable' go run ./tools/generate
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yaziciahmet/pgxtra/generate"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	if err := generate.Generate(context.Background(), generate.Options{
		DSN:           dsn,
		OutDir:        "./gen/db",
		ExcludeTables: []string{"_skip_me"},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
