package codegen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaziciahmet/pgxtra/internal/codegen"
	"github.com/yaziciahmet/pgxtra/internal/schema"
	"github.com/yaziciahmet/pgxtra/internal/testschema"
	"github.com/yaziciahmet/pgxtra/query"
)

func TestGenerateCompile(t *testing.T) {
	dir := t.TempDir()
	modRoot := filepath.Join(dir, "mod")
	genDir := filepath.Join(modRoot, "gen", "db")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(modRoot, "go.mod"), `module example.com/app

go 1.26.5

require (
	github.com/google/uuid v1.6.0
	github.com/shopspring/decimal v1.4.0
)
`)

	cfg := codegen.Config{
		Package:       "db",
		ImportPath:    "example.com/app/gen/db",
		PgxtraVersion: "test",
		QueryFS:       query.EmbeddedFS,
	}

	db := testschema.Database()
	files, err := codegen.Generate(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	if err := codegen.WriteFiles(genDir, files); err != nil {
		t.Fatal(err)
	}

	// Generated package must compile and expose nested column API.
	writeFile(t, filepath.Join(modRoot, "smoke_test.go"), `package app_test

import (
	"testing"

	"example.com/app/gen/db"
)

func TestSmoke(t *testing.T) {
	sql, args := db.Select(db.Users.ID, db.Posts.Title).
		From(db.Users).
		Join(db.Posts, db.Users.ID.EqCol(db.Posts.UserID)).
		Where(db.Users.Status.Eq("active")).
		Build()
	if sql == "" {
		t.Fatal("empty sql")
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}
`)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = modRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "test", "./...")
	cmd.Dir = modRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test ./...: %v\n%s", err, out)
	}
}

func TestGenerateUsersFileContents(t *testing.T) {
	cfg := codegen.Config{
		Package:       "db",
		ImportPath:    "example.com/app/gen/db",
		PgxtraVersion: "test",
	}
	db := testschema.Database()
	files, err := codegen.Generate(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	var users, modelUsers []byte
	for _, f := range files {
		switch {
		case f.Path == "users.go" && f.Dir == "":
			users = f.Content
		case f.Path == "users.go" && f.Dir == "models":
			modelUsers = f.Content
		}
	}
	if users == nil {
		t.Fatal("users.go not generated")
	}
	if modelUsers == nil {
		t.Fatal("models/users.go not generated")
	}
	tableBody := string(users)
	for _, want := range []string{
		"var Users = usersTable",
		"usersIDCol",
		"func (c usersIDCol) Eq(v int64)",
		"func (c usersIDCol) EqCol(other query.Column)",
		"func (c usersIDCol) In(vals ...int64)",
		"func (c usersExternalIDCol) In(vals ...uuid.UUID)",
		"func (c usersBalanceCol) In(vals ...decimal.Decimal)",
		"func (c usersStatusCol) In(vals ...models.UserStatus)",
		"func (c usersEmailCol) ILike(pattern string)",
		"func (t usersTable) Insert()",
	} {
		if !strings.Contains(tableBody, want) {
			t.Fatalf("users.go missing %q\n%s", want, tableBody)
		}
	}
	for _, bad := range []string{
		"func (c usersTagsCol) In(",
		"func (c usersTagsCol) Gt(",
		"func (c usersMetaCol) Gt(",
	} {
		if strings.Contains(tableBody, bad) {
			t.Fatalf("users.go should not contain %q", bad)
		}
	}
	if strings.Contains(tableBody, "type Users struct") {
		t.Fatalf("users.go should not contain model struct\n%s", tableBody)
	}

	modelBody := string(modelUsers)
	for _, want := range []string{
		"package models",
		"type Users struct",
		"LastLoginAt *time.Time",
		"`db:\"last_login_at\"`",
		"Tags",
		"[]string",
		"`db:\"tags\"`",
		"Status",
		"UserStatus",
		"`db:\"status\"`",
	} {
		if !strings.Contains(modelBody, want) {
			t.Fatalf("models/users.go missing %q\n%s", want, modelBody)
		}
	}
}

func TestGenerateEnumsFile(t *testing.T) {
	cfg := codegen.Config{
		Package:       "db",
		ImportPath:    "example.com/app/gen/db",
		PgxtraVersion: "test",
	}
	files, err := codegen.Generate(cfg, testschema.Database())
	if err != nil {
		t.Fatal(err)
	}
	var enums []byte
	for _, f := range files {
		if f.Path == "enums.go" && f.Dir == "models" {
			enums = f.Content
		}
	}
	if enums == nil {
		t.Fatal("models/enums.go not generated")
	}
	body := string(enums)
	for _, want := range []string{
		"package models",
		"type UserStatus string",
		"UserStatusActive",
		`UserStatusActive   UserStatus = "active"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("enums.go missing %q\n%s", want, body)
		}
	}
}

func TestGenerateNoTimeImport(t *testing.T) {
	cfg := codegen.Config{
		Package:       "db",
		ImportPath:    "example.com/app/gen/db",
		PgxtraVersion: "test",
	}
	db := schema.Database{
		Tables: []schema.Table{{
			Name: "tags",
			Columns: []schema.Column{
				{Name: "name", Go: schema.GoType{Name: "string"}},
				{Name: "slug", Go: schema.GoType{Name: "string"}},
			},
		}},
	}
	files, err := codegen.Generate(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	var tags []byte
	for _, f := range files {
		if f.Path == "tags.go" {
			tags = f.Content
		}
	}
	body := string(tags)
	if strings.Contains(body, `"time"`) {
		t.Fatalf("tags.go should not import time:\n%s", body)
	}
}

func TestEmbeddedQueryOmitsEmbedGo(t *testing.T) {
	cfg := codegen.Config{
		Package:       "db",
		ImportPath:    "example.com/app/gen/db",
		PgxtraVersion: "test",
		QueryFS:       query.EmbeddedFS,
	}
	files, err := codegen.Generate(cfg, schema.Database{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 8 {
		t.Fatalf("expected embedded query files, got %d", len(files))
	}
	for _, f := range files {
		if strings.HasSuffix(f.Path, "embed.go") {
			t.Fatalf("embed.go should not be copied: %s", f.Path)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
