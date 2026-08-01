# pgxtra — Implementation Plan (Plan 2)

Semi-locked API: **embedded builder in `gen/db`**, zero runtime pgxtra import, dual write entry (`db.Insert(Users)` and `Users.Insert()`).

**Scope of this doc:** codegen + SQL builder only. Execution (`Query`, `Exec`, pgx integration, row scanning) is a separate layer, designed later, and must not leak into builder internals.

---

## Goals

1. **Structured builders** for known SQL flows (select/from/where/join/limit, insert/set/on conflict, etc.).
2. **Type safety** via generated column/table refs — as far as Go allows without an ORM.
3. **Composable statements** — CTE wrapping insert/update/select, data-modifying CTEs, multi-CTE, without a mega-fluent API.
4. **Surgical escape hatches** — inject fragments at defined points or substitute sub-clauses, without abandoning arg numbering.
5. **Compile to `(sql, args)`** — builder never touches the DB.

---

## Mental model (short answer)

**Not** "everything is a string until the end."

**Yes:** each builder holds an **ordered list of typed parts** (select list, from table, join entries, where predicates, …). `Build()` walks those parts once and writes SQL + flattens args.

Unknown or unsupported syntax is not a separate universe — it enters the same pipeline as a **`Fragment`**: a small compile-time object that knows how to render SQL and register placeholders. Structured helpers and raw fragments share one arg registry.

```
User code
  → Builder (parts[])
  → Compile(ctx)  // single Args registry, dialect-aware
  → (sql string, args []any)
```

Execution layer (later) takes `(sql, args)` and runs it. Builder stops there.

---

## Architecture overview

```
pgxtra/                          # repo — CLI + templates only at dev time
  cmd/pgxtra/
  internal/
    introspect/                  # read PG schema
    codegen/                     # emit gen/db/*.go
  query/                         # source of truth for embedded builder
    args.go
    buf.go
    fragment.go
    expr.go
    ident.go
    select.go
    insert.go
    update.go
    delete.go
    with.go
    compile.go
    dialect_postgres.go          # v0: postgres only

gen/db/                          # GENERATED in consumer project
  users.go                       # Users table, columns, expr wrappers
  posts.go
  query/                         # copy of pgxtra/query @ version X.Y.Z
  db.go                          # Select(), Insert(), Update(), With(), Col(), Raw()
  version.go                     # // pgxtra/query v0.1.0
```

### What codegen emits vs what is templated

| Emitted per schema | Copied verbatim (embedded core) |
|--------------------|----------------------------------|
| `Users`, `Posts` table refs | All builder logic |
| Column refs + typed `.Eq()`, `.In()`, … | Args, buf, compile |
| `Users.Insert()` sugar | Select/Insert/Update/Delete/With |
| Optional row structs | Dialect (postgres v0) |
| `db.go` facades delegating to `query/` | Fragment/expr primitives |

Column expr methods on generated types are thin wrappers — they construct `query.Expr` values. No SQL logic in generated table files.

---

## Core types

### 1. `Table` / `Column` (generated)

```go
type Table interface {
    SQLName() string   // "public.users"
    Alias() string     // "users" — default, overridable via .As("u")
}

type Column interface {
    Table() Table
    SQLName() string   // "email"
    Qualifier() string // "users.email" or "u.email"
}
```

Typed columns (`ColumnOf[T]`) add compile-time `Set(val T)`, `Eq(val T)`, etc. Escape hatch columns use `Col(table, name string)`.

### 2. `Expr` (embedded core)

Boolean/scalar SQL expressions. All typed column methods return `Expr`.

```go
type Expr interface {
    compile(e *Compiler) // writes SQL + registers args
}

// helpers
func And(exprs ...Expr) Expr
func Or(exprs ...Expr) Expr
func Raw(sql string, args ...any) Expr      // predicate fragment
func SQL(sql string, args ...any) Fragment  // arbitrary injectable piece
```

### 3. `Fragment` (escape hatch backbone)

Anything that can appear inside a statement implements `Fragment`:

```go
type Fragment interface {
    compile(c *Compiler)
}
```

Built-in fragments:

| Fragment | Use |
|----------|-----|
| `RawExpr` | WHERE/HAVING/ON — `Where(Raw("x->>'k' = ?", v))` |
| `SQLPart` | Inject at marker — `sel.SQL("/* hint */")` |
| `NestedBuilder` | Subquery — `Where(Eq(col, Subquery(otherSelect)))` |
| `Ident` | CTE/table ref — `Col("stale", "id")` |

**Rule:** if it participates in `Build()`, it goes through `Compiler` so placeholders stay consistent (`$1`, `$2`, … for postgres).

### 4. `Statement` (compile root)

Top-level things that produce full SQL:

```go
type Statement interface {
    Fragment
    // Build() is sugar on root statements
}

func (s *SelectBuilder) Build() (string, []any)
```

A `SelectBuilder` is both a `Statement` and a valid subquery `Fragment` (parentheses + inner compile).

### 5. `Compiler` / `Args`

Single pass (with optional prefix for CTE):

```go
type Compiler struct {
    buf   bytes.Buffer
    args  []any
    dialect Dialect
}
```

All `compile()` methods append to shared `Compiler`. No string concat of user args into SQL — only placeholders.

---

## Builder internals: parts, not strings

Each builder owns slices of small structs. Example `SelectBuilder`:

```go
type SelectBuilder struct {
    distinct bool
    cols     []Fragment          // Column | Expr | Raw
    from     Fragment            // Table | Subquery
    joins    []joinPart          // {kind, table, on Expr}
    wheres   []Expr              // AND-combined
    groupBy  []Fragment
    having   []Expr
    orderBy  []orderPart
    limit    *int
    offset   *int
    prefix   []Fragment          // injection before SELECT
    suffix   []Fragment          // injection after whole stmt (before ;)
}
```

Methods append to parts:

```go
func (s *SelectBuilder) Where(e Expr) *SelectBuilder {
    s.wheres = append(s.wheres, e)
    return s
}
```

`Build()` = `compileSelect(s)` — fixed clause order mandated by SQL grammar. Unknown clause order is exactly why **`SQL()` injection markers** exist (see below).

Same pattern for Insert/Update/Delete.

---

## Injection markers (partial escape without full raw)

go-sqlbuilder-style **injection points** at stable grammar positions. User inserts a `Fragment` without breaking structure.

```go
sel := db.Select(db.Users.ID)
sel.SQL("/* planner hint */")           // before SELECT
sel.From(db.Users)
sel.SQL("TABLESAMPLE SYSTEM (1)")       // after FROM target — postgres-specific
sel.Where(...)
```

Internally each builder has named marker slots (enum), not free-form string concat:

```go
const (
    injectAfterFrom selectInject = iota
    injectAfterJoin
    injectBeforeWhere
    // ...
)
```

**v0:** small set of markers on Select/Insert/Update. Expand as real needs appear.

If a feature has no first-class method (e.g. `FOR UPDATE SKIP LOCKED`), preferred path:

1. `sel.SQL("FOR UPDATE SKIP LOCKED")` at suffix marker
2. If pattern repeats → promote to first-class method in next version

This keeps API small without trapping users in full raw SQL.

---

## Statement composition matrix

Not every op combines with every other. We model **valid roots** and **valid CTE bodies** explicitly instead of one god builder.

### Root statements (what `Build()` emits as top-level)

| Root | Builder |
|------|---------|
| Select | `SelectBuilder` |
| Insert | `InsertBuilder` |
| Update | `UpdateBuilder` |
| Delete | `DeleteBuilder` |
| With + X | `WithBuilder` wrapping any root above |

### CTE bodies (`With().As(...)`)

PostgreSQL allows data-modifying CTEs. Bodies can be:

| CTE body | Supported |
|----------|-----------|
| Select | yes |
| Insert | yes |
| Update | yes |
| Delete | yes |
| With (nested) | yes — nested `WithBuilder` |

`WithBuilder` holds:

```go
type WithBuilder struct {
    ctes []cteEntry          // name, optional col list, body Statement
    main Statement           // set via .Then() or .Select/.Insert/.Update/.Delete
    recursive bool           // WITH RECURSIVE
}

type cteEntry struct {
    name    string
    cols    []string         // optional explicit column names
    stmt    Statement
}
```

API (sequential — matches Plan 2):

```go
w := db.With("a", "b")       // multiple names = multiple .As calls, or chain
w.As(selectOrUpdateOrInsert)
w.As(anotherStmt)            // second CTE
main := w.Insert(db.Archive) // picks main statement kind
// ... configure main ...
sql, args := main.Build()    // WITH ... INSERT ...
```

Main statement is **always exactly one** root builder. CTEs are **ordered list**. No combinatorial explosion — composition is structural (WITH wraps one main stmt), not `InsertUpdateSelectBuilder`.

---

## Complex flow: CTE with UPDATE RETURNING → INSERT ON CONFLICT

Real postgres pattern:

```sql
WITH moved AS (
    UPDATE posts
    SET status = 'archived'
    WHERE created_at < $1
    RETURNING id, user_id
)
INSERT INTO archive_log (post_id, user_id)
SELECT id, user_id FROM moved
ON CONFLICT (post_id) DO NOTHING;
```

Plan 2 API:

```go
cutoff := time.Now().AddDate(-1, 0, 0)

w := db.With("moved")
w.As(
    db.Update(db.Posts).
        Set(db.Posts.Status, "archived").
        Where(db.Posts.CreatedAt.Lt(cutoff)).
        Returning(db.Posts.ID, db.Posts.UserID),
)

ins := w.Insert(db.ArchiveLog)
ins.FromSelect(                           // INSERT ... SELECT (not VALUES)
    db.Select(
        db.Col("moved", "id"),           // CTE column — untyped Ident
        db.Col("moved", "user_id"),
    ).From("moved"),
)
ins.OnConflict(db.ArchiveLog.PostID).DoNothing()
// or DoUpdate(...)

sql, args := ins.Build()
```

Why this works without a flexible mess:

- `UpdateBuilder.Returning()` is a normal part on update — works standalone or inside CTE.
- `InsertBuilder` has two value modes: **`Set`/`Values`** and **`FromSelect(SelectBuilder|Fragment)`** — mutually exclusive, validated at `Build()`.
- CTE columns use **`Col(cteName, colName)`** — untyped by necessity (CTEs aren't in schema).
- ON CONFLICT is a sub-builder on insert — same inside or outside WITH.

If postgres adds weird syntax: `ins.SQL("ON CONFLICT ...")` at suffix marker as fallback.

---

## Insert / Update: `Set` model

### Insert

```go
type InsertBuilder struct {
    table       Table
    sets        []colVal          // ordered pairs from Set()
    mode        insertMode        // values | query
    query       Fragment          // for INSERT ... SELECT
    onConflict  *onConflictClause
    returning   []Fragment
}
```

- `Set(col, val)` appends one column-value pair. Duplicate col → last wins or error (pick: **error at Build()**).
- Dynamic partial insert: call `Set` only for present fields.
- `FromSelect(sel)` switches mode; `Set` calls after that → error.

### Update

```go
type UpdateBuilder struct {
    table     Table
    sets      []assignment      // Set() or Assign(col, val)
    from      []Fragment        // UPDATE ... FROM (postgres)
    wheres    []Expr
    returning []Fragment
}
```

Same `Set` ergonomics. `From("stale")` for CTE join pattern.

### Dual entry (both supported)

```go
func Insert(t Table) *InsertBuilder { return &InsertBuilder{table: t} }

func (u usersTable) Insert() *InsertBuilder { return Insert(u) }
func (u usersTable) Update() *UpdateBuilder { return Update(u) }
func (u usersTable) Delete() *DeleteBuilder { return Delete(u) }
```

---

## Join

Generic only — no codegen join types:

```go
type joinKind int // INNER, LEFT, RIGHT, FULL, CROSS

func (s *SelectBuilder) Join(t Table, on Expr) *SelectBuilder
func (s *SelectBuilder) LeftJoin(t Table, on Expr) *SelectBuilder
// ...
```

`on` is user-supplied `Expr`. Table contributes `SQLName()` + alias. Cross join: `CrossJoin(t)` with no `on`.

Update also supports `From` + join-like refs for postgres `UPDATE ... FROM`.

---

## Type safety boundaries

| Safe (generated) | Untyped (escape) |
|------------------|------------------|
| `db.Users.Email` column ref | `db.Col("stale", "id")` |
| `col.Eq(val)` matching Go type | `db.Raw("...", args...)` |
| `Set(db.Users.Name, name)` | `SetAny("legacy_col", v)` |
| `Insert(db.Users)` | `From("moved")` string CTE name |
| Table refs in From/Join | `SQL("...")` injection fragments |

Principle: **schema-known → typed; computed/CTE/external → Ident/Raw**. No fake types for CTE columns.

---

## Validation strategy

Fail at **`Build()`**, not at each fluent call (keeps dynamic building easy):

| Check | When |
|-------|------|
| Insert has ≥1 value source | Build |
| `FromSelect` + `Set` mixed | Build |
| Update has ≥1 set clause | Build |
| Empty Where on Delete | optional warn, not error |
| Dialect-specific feature on wrong dialect | Build (v0: postgres only, moot) |

Return typed errors: `ErrNoColumns`, `ErrConflictValueModes`, etc.

---

## Dialect (v0)

- **Postgres only** at first — baked into embedded core at codegen (`--dialect=postgres` default).
- Placeholder format, `RETURNING`, `ON CONFLICT`, `UPDATE ... FROM` implemented once in `dialect_postgres.go`.
- CLI stamps `version.go` with query core version + dialect.
- Multi-dialect later = regen with different flag, not runtime switch (Plan 2 constraint).

---

## Codegen pipeline (step 1 scope)

```
pgxtra generate \
  --dsn=postgres://... \
  --schema=public \
  --out=./gen/db \
  --dialect=postgres
```

1. **Introspect** — tables, columns, types, enums, nullability, defaults (defaults inform optional scan types later, not builder v0).
2. **Emit schema files** — one file per table or grouped by schema.
3. **Emit `db.go`** — facades: `Select`, `Insert`, `Update`, `Delete`, `With`, `Raw`, `Col`, `Assign`.
4. **Copy `query/`** — embed from pgxtra repo, write `version.go`.
5. **Go format** — `go/format` all outputs.

No join/FK codegen. FK metadata optional later for docs/lints only.

---

## Implementation phases

### Phase 0 — Skeleton

- [ ] Repo layout: `query/`, `cmd/pgxtra`, `internal/introspect`, `internal/codegen`
- [ ] `Compiler`, `Args`, `Buf`, `Fragment`, `Expr` primitives
- [ ] Postgres `$N` placeholder numbering tests

### Phase 1 — Core statements (no CTE)

- [ ] `SelectBuilder` — cols, from, join, where, group, having, order, limit, offset
- [ ] `InsertBuilder` — Set, Returning, basic ON CONFLICT DO NOTHING / DO UPDATE
- [ ] `UpdateBuilder` — Set, Where, Returning
- [ ] `DeleteBuilder` — Where, Returning
- [ ] Codegen: tables, columns, typed expr wrappers
- [ ] Dual entry: `Insert(t)` + `t.Insert()`
- [ ] Golden SQL tests (builder → expected string + args)

### Phase 2 — CTE + composition

- [ ] `WithBuilder` — multiple CTEs, nested WITH
- [ ] Data-modifying CTE bodies (insert/update/delete as CTE stmt)
- [ ] `InsertBuilder.FromSelect`
- [ ] `UpdateBuilder.From` (postgres UPDATE FROM)
- [ ] `Col(name, column)` / `Ident` for CTE refs
- [ ] Complex golden tests (UPDATE RETURNING → INSERT ON CONFLICT)

### Phase 3 — Escape hatches + hard postgres

- [ ] Injection markers (`SQL()` at named positions)
- [ ] `SetAny`, `Raw`, nested subquery fragments
- [ ] ON CONFLICT DO UPDATE full assign list
- [ ] Union / Except / Intersect (if needed — as `SelectBuilder.Combine()` or separate entry)

### Phase 4 — Codegen polish

- [ ] Enum types
- [ ] Array column types
- [ ] Version stamp + drift warning in CLI
- [ ] Optional row struct generation (for future exec layer, fields tagged for scan)

### Later (explicitly out of scope now)

- Execution layer (`db.Exec(ctx, stmt)`, pgx pool wrapper)
- Full raw query helper (exec concern)
- Migration generator
- MySQL/SQLite dialects

---

## Execution layer interface (future, not built now)

Design constraint: builder output must be enough for any executor.

```go
// future — not in v0
type BuiltQuery struct {
    SQL  string
    Args []any
}

// pgx.Exec(ctx, q.SQL, q.Args...)
```

No `Query()` methods on builders in v0. Keeps builder pure and testable.

---

## Testing approach

1. **Unit golden tests** — `Build()` → want SQL + args (primary).
2. **Fragment tests** — arg renumbering when nesting subqueries + CTE.
3. **Codegen tests** — run generator against testcontainers postgres, assert files compile.
4. **Compile-only integration** — generated `gen/db` used in example module.

No DB execution required for builder correctness.

---

## Design principles (checklist)

| Principle | How |
|-----------|-----|
| Structured > stringly | Builders hold parts; one compile pass |
| Small API surface | Grammar-aligned methods; promote repeated Raw patterns |
| Composable, not combinatorial | WITH wraps one main stmt; CTE list; subqueries as Fragment |
| Type safe where schema exists | Generated Column/Table |
| CTE / dynamic → Ident/Raw | Explicit untyped tier |
| Unsupported → Fragment, not panic | Injection markers + Raw |
| No exec in builder | `(sql, args)` only |
| Plan 2 packaging | Core embedded in `gen/db/query/` |

---

## Open implementation choices (need decision before coding)

1. **Duplicate `Set` on same column** — error vs last-wins? (Recommend: **error at Build()**.)
2. **CTE API** — `With("a").As(x).As(y)` vs `With().As("a", x).As("y", y)`? (Recommend: **chain `.As(stmt)`**, name on `With(name)`.)
3. **Subquery API** — `Subquery(sel *SelectBuilder)` vs implicit (SelectBuilder is Fragment). (Recommend: **implicit** — select is Fragment, wraps in parens.)
4. **Injection marker set for v0** — minimal (prefix/suffix/after-from) or go-sqlbuilder-full? (Recommend: **minimal**, expand on demand.)

---

## Summary

Plan 2 builder is a **part-accumulating compiler**, not a string builder. Known SQL clauses are methods; unknown syntax enters as **`Fragment`** at defined injection points or as **`Raw`/`Col`/`SQL`** — always through the same `Args` registry. CTEs compose by wrapping a **single main Statement** with an ordered list of **any Statement bodies**, which covers UPDATE RETURNING inside a CTE feeding INSERT ON CONFLICT without a mega-API. Execution stays out until `(sql, args)` is solid.
