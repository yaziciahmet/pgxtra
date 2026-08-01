# pgxtra — API Design Plans (v2)

Three API designs for a codegen SQL builder. Focus is entirely on **how generated code is used at runtime** — not implementation.

## Hard requirements (all plans)

- **SQL-builder, not ORM.** No entity objects, no `user.SetName()`. You build SQL, execute with pgx/`database/sql`.
- **Sequential mutable builders** (go-sqlbuilder style).
- **Insert/update use `Set(col, val)`** — not `Cols(...).Values(...)`. Column and value stay paired; dynamic partial sets are natural.
- **Generic `Join(table, on)` only.** No `JoinPosts()`, no FK-derived join types, no relationship codegen. You write the ON clause yourself using column refs.
- **Joins, CTEs, ON CONFLICT, RETURNING** supported.
- **Type-safe by default, escape hatch when needed.**

## The join insight (why old Plan B is out)

Per-table select builders (`SelectUsers()`, `Users.Select()`) fight joins — a select always spans tables, but insert/update/delete target one table.

**All three plans use the same split:**

| Operation | Builder scope |
|-----------|---------------|
| `Select`, `With` (CTE) | **Schema-generic** — one builder, any tables/columns |
| `Insert`, `Update`, `Delete` | **Table-targeted** — knows which table is being written |
| `Join` | **Generic method** on select/update — `Join(table, onExpr)` |

Codegen never emits join helpers. It only emits **table refs** and **column refs**.

---

## Shared vocabulary

| Term | Meaning |
|------|---------|
| **Codegen CLI** | `pgxtra generate` — reads DB schema, writes Go files |
| **Generated package** | e.g. `myapp/gen/db` — tables, columns, optional row structs |
| **Runtime library** | `github.com/yaziciahmet/pgxtra/...` — imported at runtime only if the plan says so |
| **Column ref** | Generated value (`Users.Email`) — SQL name, Go type, table alias |
| **Table ref** | Generated singleton (`Users`, `Posts`) — table name, alias default |
| **Expr** | Boolean/SQL fragment from column methods (`.Eq()`, `.In()`, `.Lt()`, …) or escape hatch |

### Insert/update target shape (all plans)

```go
ib.Set(Users.Email, email)
ib.Set(Users.Name, name)
if avatarURL != "" {
    ib.Set(Users.AvatarURL, avatarURL) // optional col — no index bookkeeping
}
```

---

## Example domain (used in all examples below)

```
users  (id, email, name, status, last_login_at, created_at, updated_at)
posts  (id, user_id, title, published, created_at)
```

---

## Plan 1 — Schema in gen, builders in runtime

**Philosophy:** Codegen is dumb — tables and columns only. All builder logic lives in a small runtime package you import alongside `gen/db`.

**Best if:** Smallest generated output matters; you're fine with `pgxtra/builder` as a permanent dep.

### Generated vs runtime

| Generated (`gen/db`) | Runtime (`pgxtra/builder`) |
|----------------------|----------------------------|
| `Users`, `Posts` table refs | `Select()`, `Insert()`, `Update()`, `Delete()`, `With()` |
| Column refs + `.Eq()`, `.In()`, … | `Join()`, `From()`, `Where()`, `Set()`, `Build()` |
| Enums, composite types | SQL compilation, placeholders, dialect |
| Optional row structs for scan | CTE chaining, ON CONFLICT, RETURNING |

### Example 1 — Select with join + conditions

```go
import (
    "time"

    "github.com/yaziciahmet/pgxtra/builder"
    "myapp/gen/db"
)

since := time.Now().AddDate(0, -1, 0)

sel := builder.Select(
    db.Users.ID,
    db.Users.Name,
    db.Posts.Title,
    db.Posts.CreatedAt,
)
sel.From(db.Users)
sel.Join(db.Posts, db.Users.ID.Eq(db.Posts.UserID))
sel.Where(db.Users.Status.Eq("active"))
sel.Where(db.Posts.Published.Eq(true))
sel.Where(db.Posts.CreatedAt.Gte(since))
sel.OrderByDesc(db.Users.CreatedAt)
sel.Limit(50)

sql, args := sel.Build()
// SELECT users.id, users.name, posts.title, posts.created_at
// FROM users
// JOIN posts ON users.id = posts.user_id
// WHERE users.status = $1 AND posts.published = $2 AND posts.created_at >= $3
// ORDER BY users.created_at DESC
// LIMIT 50
```

### Example 2 — Insert ON CONFLICT DO UPDATE

```go
now := time.Now()

ins := builder.Insert(db.Users)
ins.Set(db.Users.Email, email)
ins.Set(db.Users.Name, name)
ins.Set(db.Users.Status, "active")
ins.Set(db.Users.CreatedAt, now)
ins.Set(db.Users.UpdatedAt, now)
ins.OnConflict(db.Users.Email).DoUpdate(
    builder.Assign(db.Users.Name, name),
    builder.Assign(db.Users.UpdatedAt, now),
)
ins.Returning(db.Users.ID)

sql, args := ins.Build()
// INSERT INTO users (email, name, status, created_at, updated_at) VALUES (...)
// ON CONFLICT (email) DO UPDATE SET name = $N, updated_at = $M
// RETURNING id
```

### Example 3 — Update with WHERE

```go
cutoff := time.Now().AddDate(0, 0, -30)
now := time.Now()

upd := builder.Update(db.Users)
upd.Set(db.Users.Status, "inactive")
upd.Set(db.Users.UpdatedAt, now)
upd.Where(db.Users.LastLoginAt.Lt(cutoff))
upd.Where(db.Users.Status.Neq("deleted"))
upd.Where(db.Users.Status.Neq("inactive")) // skip already-inactive

sql, args := upd.Build()
// UPDATE users SET status = $1, updated_at = $2
// WHERE last_login_at < $3 AND status != $4 AND status != $5
```

### Example 4 — CTE select, then update from CTE

```go
cutoff := time.Now().AddDate(0, 0, -90)
now := time.Now()

q := builder.With("stale")
q.As(
    builder.Select(db.Users.ID).
        From(db.Users).
        Where(db.Users.LastLoginAt.Lt(cutoff)).
        Where(db.Users.Status.Eq("active")),
)

upd := q.Update(db.Users)
upd.Set(db.Users.Status, "inactive")
upd.Set(db.Users.UpdatedAt, now)
upd.From("stale")
upd.Where(db.Users.ID.Eq(builder.Col("stale", "id")))

sql, args := upd.Build()
// WITH stale AS (
//   SELECT id FROM users WHERE last_login_at < $1 AND status = $2
// )
// UPDATE users SET status = $3, updated_at = $4
// FROM stale
// WHERE users.id = stale.id
```

### Escape hatch

```go
sel.Where(builder.Raw("users.metadata->>'tier' = ?", tier))
ins.SetRaw("legacy_col", val)
upd.Where(builder.Raw("updated_at < NOW() - INTERVAL '7 days'"))
```

### Runtime dependency

**Yes** — `github.com/yaziciahmet/pgxtra/builder` always imported.

### Tradeoffs

| Pros | Cons |
|------|------|
| Tiny `gen/` | Runtime dep forever |
| Builder fixes ship without regen | Two imports in every query file |
| Dialect can switch at runtime | Generated code not self-contained |

---

## Plan 2 — Schema + embedded builders in gen (zero runtime)

**Philosophy:** CLI generates column/table refs **and** copies a generic builder implementation into `gen/db/query/`. Your app imports only `gen/db`. No pgxtra at runtime.

**Best if:** Zero runtime dependency is the goal, but you don't want N copies of builder logic per table.

### Generated vs runtime

| Generated (`gen/db`) | Runtime |
|----------------------|---------|
| `Users`, `Posts`, column refs, expr methods | **Nothing** |
| `gen/db/query/` — embedded builder core (copied by CLI) | |
| Package-level `Select()`, `Insert(table)`, `Update(table)`, `With()` | |
| Table convenience: `Users.Insert()` → `Insert(Users)` | |

Layout:

```
gen/db/
  users.go          # Users table + columns
  posts.go
  query/            # embedded builder — DO NOT EDIT (regenerated)
  query.go          # package-level Select, Insert, With, …
```

### Example 1 — Select with join + conditions

```go
import (
    "time"

    "myapp/gen/db"
)

since := time.Now().AddDate(0, -1, 0)

sel := db.Select(
    db.Users.ID,
    db.Users.Name,
    db.Posts.Title,
    db.Posts.CreatedAt,
)
sel.From(db.Users)
sel.Join(db.Posts, db.Users.ID.Eq(db.Posts.UserID))
sel.Where(db.Users.Status.Eq("active"))
sel.Where(db.Posts.Published.Eq(true))
sel.Where(db.Posts.CreatedAt.Gte(since))
sel.OrderByDesc(db.Users.CreatedAt)
sel.Limit(50)

sql, args := sel.Build()
```

Same SQL as Plan 1. API is identical except import path — `db.Select` instead of `builder.Select`.

### Example 2 — Insert ON CONFLICT DO UPDATE

```go
now := time.Now()

ins := db.Insert(db.Users)
ins.Set(db.Users.Email, email)
ins.Set(db.Users.Name, name)
ins.Set(db.Users.Status, "active")
ins.Set(db.Users.CreatedAt, now)
ins.Set(db.Users.UpdatedAt, now)
ins.OnConflict(db.Users.Email).DoUpdate(
    db.Assign(db.Users.Name, name),
    db.Assign(db.Users.UpdatedAt, now),
)
ins.Returning(db.Users.ID)

sql, args := ins.Build()
```

Table-scoped entry also works:

```go
ins := db.Users.Insert() // sugar for db.Insert(db.Users)
```

### Example 3 — Update with WHERE

```go
cutoff := time.Now().AddDate(0, 0, -30)
now := time.Now()

upd := db.Update(db.Users)
upd.Set(db.Users.Status, "inactive")
upd.Set(db.Users.UpdatedAt, now)
upd.Where(db.Users.LastLoginAt.Lt(cutoff))
upd.Where(db.Users.Status.Neq("deleted"))
upd.Where(db.Users.Status.Neq("inactive"))

sql, args := upd.Build()
```

Or: `db.Users.Update().Set(...).Where(...)`

### Example 4 — CTE select, then update from CTE

```go
cutoff := time.Now().AddDate(0, 0, -90)
now := time.Now()

q := db.With("stale")
q.As(
    db.Select(db.Users.ID).
        From(db.Users).
        Where(db.Users.LastLoginAt.Lt(cutoff)).
        Where(db.Users.Status.Eq("active")),
)

upd := q.Update(db.Users)
upd.Set(db.Users.Status, "inactive")
upd.Set(db.Users.UpdatedAt, now)
upd.From("stale")
upd.Where(db.Users.ID.Eq(db.Col("stale", "id")))

sql, args := upd.Build()
```

### Escape hatch

```go
sel.Where(db.Raw("users.metadata->>'tier' = ?", tier))
ins.SetAny("legacy_col", val)
```

### Runtime dependency

**None.** Upgrade builder = bump CLI version + `pgxtra generate`.

### Tradeoffs

| Pros | Cons |
|------|------|
| Single import (`gen/db`) | `gen/db/query/` diff noise on regen |
| No pgxtra in prod `go.mod` | Dialect baked at codegen time |
| One builder copy per schema, not per table | Embedded core can drift from CLI version if you forget regen |

---

## Plan 3 — Table methods for writes, package funcs for reads (fluent gen)

**Philosophy:** Generated code is the primary API surface. Column refs have rich methods. Writes hang off the table ref; reads are package-level. Builder core is generated inline (no separate `query/` subfolder — flatter package). Still zero runtime.

**Best if:** You want generated code to read like a purpose-built library for *your* schema, without a visible "embedded copy" subdirectory.

### Generated vs runtime

| Generated (`gen/db`) | Runtime |
|----------------------|---------|
| `Users`, `Posts` with column fields | **Nothing** |
| Column methods: `.Eq()`, `.Set(val)`, `.Assign(val)` | |
| `Users.Insert()`, `Users.Update()`, `Users.Delete()` | |
| `Select(...)`, `With(name)` at package level | |
| Builder structs + SQL emit in same package | |

Key API difference: **`.Set(val)` on column ref** for insert/update shorthand.

```go
ins.Set(db.Users.Name.Set(name))       // column ref carries assignment
ins.Set(db.Users.Email, email)         // or classic pair form — both work
```

### Example 1 — Select with join + conditions

```go
import (
    "time"

    "myapp/gen/db"
)

since := time.Now().AddDate(0, -1, 0)

sel := db.Select(
    db.Users.ID,
    db.Users.Name,
    db.Posts.Title,
    db.Posts.CreatedAt,
)
sel.From(db.Users)
sel.Join(db.Posts, db.Users.ID.Eq(db.Posts.UserID))
sel.Where(db.Users.Status.Eq("active"))
sel.Where(db.Posts.Published.IsTrue())
sel.Where(db.Posts.CreatedAt.Gte(since))
sel.GroupBy(db.Users.ID, db.Users.Name) // if you need dedup from join
sel.Having(db.Count(db.Posts.ID).Gt(0))
sel.OrderByDesc(db.Users.CreatedAt)
sel.Limit(50)

sql, args := sel.Build()
```

Same join pattern. Extra aggregate lines show medium complexity without new concepts.

### Example 2 — Insert ON CONFLICT DO UPDATE

```go
now := time.Now()

ins := db.Users.Insert()
ins.Set(db.Users.Email.Set(email))
ins.Set(db.Users.Name.Set(name))
ins.Set(db.Users.Status.Set("active"))
ins.Set(db.Users.CreatedAt.Set(now))
ins.Set(db.Users.UpdatedAt.Set(now))
ins.OnConflict(db.Users.Email).DoUpdate(
    db.Users.Name.Set(name),
    db.Users.UpdatedAt.Set(now),
)
ins.Returning(db.Users.ID)

sql, args := ins.Build()
```

Dynamic partial insert — only set what you have:

```go
ins := db.Users.Insert()
ins.Set(db.Users.Email.Set(email))
ins.Set(db.Users.Name.Set(name))
if status != "" {
    ins.Set(db.Users.Status.Set(status))
}
```

### Example 3 — Update with WHERE

```go
cutoff := time.Now().AddDate(0, 0, -30)
now := time.Now()

upd := db.Users.Update()
upd.Set(db.Users.Status.Set("inactive"))
upd.Set(db.Users.UpdatedAt.Set(now))
upd.Where(db.Users.LastLoginAt.Lt(cutoff))
upd.Where(db.Users.Status.NotIn("deleted", "inactive"))

sql, args := upd.Build()
```

### Example 4 — CTE select, then update from CTE

```go
cutoff := time.Now().AddDate(0, 0, -90)
now := time.Now()

upd := db.With("stale").
    As(
        db.Select(db.Users.ID).
            From(db.Users).
            Where(db.Users.LastLoginAt.Lt(cutoff)).
            Where(db.Users.Status.Eq("active")),
    ).
    Update(db.Users).
    Set(db.Users.Status.Set("inactive")).
    Set(db.Users.UpdatedAt.Set(now)).
    From("stale").
    Where(db.Users.ID.Eq(db.Ident("stale", "id"))).
    Build()

sql, args := upd, updArgs // Build() returns (sql, args)
```

Chained fluent form for CTE → update in one expression. Sequential form (like Plan 1/2) also available:

```go
q := db.With("stale")
q.As(db.Select(db.Users.ID).From(db.Users).Where(...))
upd := q.Update(db.Users)
// ...
```

`db.Ident("stale", "id")` — untyped ref to a CTE column (not a generated column ref, since CTEs aren't in the schema).

### Escape hatch

```go
sel.Where(db.Expr("users.metadata->>'tier' = ?", tier))
ins.SetAny("legacy_col", val)
upd.Set(db.Users.Name.SetExpr("COALESCE($1, name)", fallback))
```

### Runtime dependency

**None.**

### Tradeoffs

| Pros | Cons |
|------|------|
| Most ergonomic generated API | Largest generated package (builders inline, not subfolder) |
| Table write entry (`Users.Insert()`) reads naturally | `.Set(col.Set(val))` vs `.Set(col, val)` — two forms to learn |
| Zero runtime, fluent CTE chains | Regen replaces all builder code in package |

---

## Feature matrix

| | Plan 1 | Plan 2 | Plan 3 |
|---|--------|--------|--------|
| Generic `Join(table, on)` | Yes | Yes | Yes |
| No join codegen | Yes | Yes | Yes |
| Generic `Select` (cross-table) | Yes | Yes | Yes |
| Table-scoped `Insert`/`Update` | Yes | Yes | Yes |
| `Set(col, val)` insert/update | Yes | Yes | Yes |
| CTE → UPDATE | Yes | Yes | Yes (fluent chain) |
| ON CONFLICT DO UPDATE | Yes | Yes | Yes |
| pgxtra runtime import | **Yes** | **No** | **No** |
| Generated size | Small | Medium | Medium–large |
| Write entry style | `builder.Insert(Users)` | `db.Insert(Users)` or `Users.Insert()` | `Users.Insert()` |
| Fix builder bug | Release lib | Regen | Regen |

---

## Do any plans fail the requirements?

**Old Plan B (per-table `SelectUsers`) — eliminated.** Joins need a cross-table select builder. None of the three plans above use it.

**All four example scenarios work in all three plans** with the same join model:

```go
sel.Join(db.Posts, db.Users.ID.Eq(db.Posts.UserID))
```

No plan requires generated join types.

---

## Comparison at a glance

```
Plan 1:  import builder + gen/db     →  builder.Select(...).Join(Posts, on)
Plan 2:  import gen/db only          →  db.Select(...).Join(Posts, on)
Plan 3:  import gen/db only          →  db.Select(...).Join(Posts, on)
                                      db.Users.Insert().Set(Users.Email.Set(x))
```

Plans 2 and 3 are nearly identical in capability. The difference is packaging and write-side ergonomics:

| | Plan 2 | Plan 3 |
|---|--------|--------|
| Builder location | `gen/db/query/` subpackage | inline in `gen/db` |
| Write API | `Set(col, val)` primarily | `Set(col.Set(val))` + pair form |
| CTE → update | sequential steps | sequential or fluent `.Build()` chain |
| Feels like | go-sqlbuilder + typed columns | schema-native mini library |

---

## Recommendation

Given your constraints (no join codegen, zero runtime if possible, go-sqlbuilder feel):

1. **Plan 2** if you want the closest match to go-sqlbuilder ergonomics with zero runtime — clean `Set(col, val)`, explicit builder subpackage.
2. **Plan 3** if you want writes to feel more codegen-native (`Users.Insert()`, `col.Set(val)`) and like fluent CTE chains.
3. **Plan 1** only if you explicitly want a living runtime library and smallest possible `gen/`.

---

## Open decisions (after picking a plan)

1. **Write entry:** `db.Insert(Users)` vs `Users.Insert()` — Plans 2 supports both; Plan 3 prefers table method.
2. **CTE column refs:** `Col("stale", "id")` vs `Ident("stale", "id")` vs string `"stale.id"` — pick one escape hatch.
3. **Scan structs:** generate row types for `pgx.RowToStructByName`, or builders only?
4. **Dialect:** fixed at codegen (Plans 2/3) vs runtime switch (Plan 1)?
5. **`Build()` return:** `(sql, args)` only, or also helpers like `BuildExec()`, `SQL()` fragment injection?

---

## Next step

Pick **Plan 1, 2, or 3** (or a mix — e.g. "Plan 2 API but Plan 3 fluent CTE"). Then we freeze v0 before writing any code.
