# typg

Typed Postgres — schema codegen and SQL query builder for Go.

Generate typed table refs, column methods, and row models from your database, then build SQL with compile-time type safety. No runtime dependency on typg in your app; the query builder is embedded into your generated package.

## Install

```bash
go get github.com/yaziciahmet/typg
```

## Usage

```go
import "github.com/yaziciahmet/typg/generate"

err := generate.Generate(ctx, generate.Options{
    DSN:           os.Getenv("DATABASE_URL"),
    OutDir:        "./gen/db",
    ExcludeTables: []string{"schema_migrations"},
})
```

See [examples/basic](examples/basic) for a full walkthrough.

## License

MIT — see [LICENSE](LICENSE).
