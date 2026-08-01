# typg

**Typed Postgres for Go** — introspect your database and generate type-safe table refs, column methods, and row models. Build SQL with compile-time checks, no stringly-typed queries.

typg embeds a lightweight query builder into your generated package, so your app has **zero runtime dependency** on typg itself.

## Install

```bash
go get github.com/yaziciahmet/typg
```

## Generate

```go
import (
    "context"
    "os"

    "github.com/yaziciahmet/typg/generate"
)

err := generate.Generate(ctx, generate.Options{
    DSN:           os.Getenv("DATABASE_URL"),
    OutDir:        "./gen/db",
    Schema:        "public",
    ExcludeTables: []string{"schema_migrations"},
})
```

**Output:**

```
gen/
├── db/           # typed columns, joins, query builder facade
│   └── query/    # embedded SQL builder (copied at codegen time)
└── models/       # row structs with db tags + enum types
```

## Query

```go
import (
    "myapp/gen/db"
    "myapp/gen/models"
)

sql, args := db.Select(db.Users.Name, db.Posts.Title).
    From(db.Users).
    Join(db.Posts, db.Users.ID.EqCol(db.Posts.UserID)).
    Where(db.Users.Status.Eq(models.UserStatusActive)).
    Where(db.Users.Email.ILike("%@example.com")).
    Build()
```

Pass `sql` and `args` to pgx — typg handles generation and query building, not execution.

## Example

See [`examples/basic`](examples/basic) for a complete setup with schema, generate tool, and tests.

## License

MIT — see [LICENSE](LICENSE).
