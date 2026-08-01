package introspect_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaziciahmet/pgxtra/internal/introspect"
)

func TestPostgresIntrospect(t *testing.T) {
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

	schemaSQL, err := os.ReadFile(filepath.Join(testdataDir(t), "schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		t.Fatal(err)
	}

	db, err := introspect.Postgres(ctx, pool, introspect.Options{
		Schema:        "public",
		ExcludeTables: []string{"_skip_me"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(db.Tables) != 2 {
		t.Fatalf("tables = %d, want 2", len(db.Tables))
	}
	if db.Tables[0].Name != "posts" || db.Tables[1].Name != "users" {
		t.Fatalf("table order = %v, %v", db.Tables[0].Name, db.Tables[1].Name)
	}

	if len(db.Enums) != 1 || db.Enums[0].Name != "user_status" {
		t.Fatalf("enums = %+v", db.Enums)
	}

	var usersCols int
	for _, table := range db.Tables {
		if table.Name == "users" {
			usersCols = len(table.Columns)
			for _, col := range table.Columns {
				switch col.Name {
				case "status":
					if col.Go.Name != "UserStatus" || !col.Go.IsEnum {
						t.Fatalf("status go type = %+v", col.Go)
					}
				case "tags":
					if col.Go.Name != "[]string" || !col.Nullable {
						t.Fatalf("tags go type = %+v nullable=%v", col.Go, col.Nullable)
					}
				case "external_id":
					if col.Go.Name != "uuid.UUID" {
						t.Fatalf("external_id go type = %+v", col.Go)
					}
				case "meta":
					if col.Go.Name != "json.RawMessage" {
						t.Fatalf("meta go type = %+v", col.Go)
					}
				case "balance":
					if col.Go.Name != "decimal.Decimal" {
						t.Fatalf("balance go type = %+v", col.Go)
					}
				case "last_login_at":
					if col.Go.Name != "time.Time" {
						t.Fatalf("last_login_at go type = %+v", col.Go)
					}
				}
			}
		}
	}
	if usersCols != 10 {
		t.Fatalf("users columns = %d", usersCols)
	}

	// sanity: mapped types compile
	_ = time.Now()
}

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}
