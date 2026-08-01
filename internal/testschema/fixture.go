package testschema

import "github.com/yaziciahmet/typg/internal/schema"

// Database returns a fixture schema for codegen tests.
func Database() schema.Database {
	return schema.Database{
		Schema: "public",
		Enums: []schema.Enum{{
			Name:   "user_status",
			Values: []string{"active", "inactive", "deleted"},
		}},
		Tables: []schema.Table{
			{
				Name: "users",
				Columns: []schema.Column{
					{Name: "id", PGType: "int8", Go: schema.GoType{Name: "int64"}},
					{Name: "email", PGType: "text", Go: schema.GoType{Name: "string"}},
					{Name: "name", PGType: "text", Go: schema.GoType{Name: "string"}},
					{Name: "status", PGType: "user_status", Go: schema.GoType{Name: "UserStatus", IsEnum: true}},
					{Name: "tags", PGType: "_text", Nullable: true, Go: schema.GoType{Name: "[]string"}},
					{Name: "meta", PGType: "jsonb", Nullable: true, Go: schema.GoType{Name: "json.RawMessage", ImportPath: "encoding/json"}},
					{Name: "balance", PGType: "numeric", Nullable: true, Go: schema.GoType{Name: "decimal.Decimal", ImportPath: "github.com/shopspring/decimal"}},
					{Name: "external_id", PGType: "uuid", Nullable: true, Go: schema.GoType{Name: "uuid.UUID", ImportPath: "github.com/google/uuid"}},
					{Name: "last_login_at", PGType: "timestamptz", Nullable: true, Go: schema.GoType{Name: "time.Time"}},
					{Name: "updated_at", PGType: "timestamptz", Go: schema.GoType{Name: "time.Time"}},
				},
			},
			{
				Name: "posts",
				Columns: []schema.Column{
					{Name: "id", PGType: "int8", Go: schema.GoType{Name: "int64"}},
					{Name: "user_id", PGType: "int8", Go: schema.GoType{Name: "int64"}},
					{Name: "title", PGType: "text", Go: schema.GoType{Name: "string"}},
					{Name: "published", PGType: "bool", Go: schema.GoType{Name: "bool"}},
					{Name: "created_at", PGType: "timestamptz", Go: schema.GoType{Name: "time.Time"}},
				},
			},
		},
	}
}
