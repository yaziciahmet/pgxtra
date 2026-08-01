package generate_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaziciahmet/typg/generate"
)

func TestGenerateFromPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainers in short mode")
	}

	ctx := context.Background()
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	schemaSQL, err := os.ReadFile(filepath.Join("..", "internal", "introspect", "testdata", "schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		t.Fatal(err)
	}

	modRoot := t.TempDir()
	writeFile(t, filepath.Join(modRoot, "go.mod"), `module example.com/app

go 1.26.5

require (
	github.com/google/uuid v1.6.0
	github.com/shopspring/decimal v1.4.0
)
`)

	outDir := filepath.Join(modRoot, "gen", "db")
	if err := generate.Generate(ctx, generate.Options{
		Pool:          pool,
		OutDir:        outDir,
		ExcludeTables: []string{"_skip_me"},
	}); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(modRoot, "smoke_test.go"), `package app_test

import (
	"testing"

	"example.com/app/gen/db"
)

func TestSmoke(t *testing.T) {
	sql, args := db.Select(db.Users.ID).
		From(db.Users).
		Where(db.Users.Status.Eq("active")).
		Build()
	if sql == "" {
		t.Fatal("empty sql")
	}
	_ = args
}
`)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = modRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "test", "./...")
	cmd.Dir = modRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go test ./...: %v\n%s", err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
