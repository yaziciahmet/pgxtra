# pgxtra basic example

Minimal blog schema (`users` + `posts`) showing the `generate` SDK.

## Regenerate

```bash
# start postgres
docker run -d --name pgxtra-basic \
  -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=blog \
  -p 5434:5432 postgres:16-alpine

# apply schema
psql postgres://user:pass@localhost:5434/blog -f schema.sql

# generate typed db + models
DATABASE_URL='postgres://user:pass@localhost:5434/blog?sslmode=disable' \
  go run ./tools/generate
```

Output:

- `gen/db/` — typed table/column refs + embedded query builder
- `gen/models/` — row structs + enums

## Usage

```go
import (
    "github.com/yaziciahmet/pgxtra/examples/basic/gen/db"
    "github.com/yaziciahmet/pgxtra/examples/basic/gen/models"
)

sql, args := db.Select(db.Users.Name, db.Posts.Title).
    From(db.Users).
    Join(db.Posts, db.Users.ID.EqCol(db.Posts.UserID)).
    Where(db.Users.Status.Eq(models.UserStatusActive)).
    Build()
// SELECT users.name, posts.title FROM users
// INNER JOIN posts ON users.id = posts.user_id
// WHERE users.status = $1
```

Run tests:

```bash
go test .
```
