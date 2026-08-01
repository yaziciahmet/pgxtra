package introspect

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaziciahmet/pgxtra/internal/schema"
)

// Options configures postgres introspection.
type Options struct {
	Schema        string
	ExcludeTables []string
}

// Postgres reads schema metadata from a live database.
func Postgres(ctx context.Context, pool *pgxpool.Pool, opts Options) (schema.Database, error) {
	schemaName := opts.Schema
	if schemaName == "" {
		schemaName = "public"
	}
	excluded := excludeSet(opts.ExcludeTables)

	enums, err := loadEnums(ctx, pool, schemaName)
	if err != nil {
		return schema.Database{}, err
	}
	enumMap := enumIndex(enums)

	rows, err := pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, schemaName)
	if err != nil {
		return schema.Database{}, err
	}
	defer rows.Close()

	var tables []schema.Table
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return schema.Database{}, err
		}
		if excluded[name] {
			continue
		}
		cols, err := columns(ctx, pool, schemaName, name, enumMap)
		if err != nil {
			return schema.Database{}, err
		}
		tables = append(tables, schema.Table{Name: name, Columns: cols})
	}
	if err := rows.Err(); err != nil {
		return schema.Database{}, err
	}

	return schema.Database{Schema: schemaName, Enums: enums, Tables: tables}, nil
}

func loadEnums(ctx context.Context, pool *pgxpool.Pool, schemaName string) ([]schema.Enum, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.typname, e.enumlabel
		FROM pg_type t
		JOIN pg_enum e ON e.enumtypid = t.oid
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = $1
		ORDER BY t.typname, e.enumsortorder
	`, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byName := map[string][]string{}
	var order []string
	for rows.Next() {
		var name, label string
		if err := rows.Scan(&name, &label); err != nil {
			return nil, err
		}
		if _, ok := byName[name]; !ok {
			order = append(order, name)
		}
		byName[name] = append(byName[name], label)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	enums := make([]schema.Enum, 0, len(order))
	for _, name := range order {
		enums = append(enums, schema.Enum{Name: name, Values: byName[name]})
	}
	return enums, nil
}

func enumIndex(enums []schema.Enum) map[string]schema.Enum {
	m := make(map[string]schema.Enum, len(enums))
	for _, e := range enums {
		m[e.Name] = e
	}
	return m
}

func columns(ctx context.Context, pool *pgxpool.Pool, schemaName, tableName string, enums map[string]schema.Enum) ([]schema.Column, error) {
	rows, err := pool.Query(ctx, `
		SELECT column_name, udt_name, is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []schema.Column
	for rows.Next() {
		var name, udt, nullable string
		if err := rows.Scan(&name, &udt, &nullable); err != nil {
			return nil, err
		}
		goType, err := ResolveType(udt, enums)
		if err != nil {
			return nil, fmt.Errorf("table %s column %s: %w", tableName, name, err)
		}
		cols = append(cols, schema.Column{
			Name:     name,
			PGType:   udt,
			Nullable: nullable == "YES",
			Go:       goType,
		})
	}
	return cols, rows.Err()
}

func excludeSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			out[n] = true
		}
	}
	return out
}

// SortTables sorts tables by name for stable output.
func SortTables(db *schema.Database) {
	sort.Slice(db.Tables, func(i, j int) bool {
		return db.Tables[i].Name < db.Tables[j].Name
	})
}
