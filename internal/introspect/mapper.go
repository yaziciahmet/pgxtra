package introspect

import (
	"fmt"
	"strings"

	"github.com/yaziciahmet/pgxtra/internal/schema"
)

// ResolveType maps a postgres udt_name to a Go type.
func ResolveType(udtName string, enums map[string]schema.Enum) (schema.GoType, error) {
	if strings.HasPrefix(udtName, "_") {
		elem, err := ResolveType(udtName[1:], enums)
		if err != nil {
			return schema.GoType{}, err
		}
		return schema.GoType{
			Name:       "[]" + elem.Name,
			ImportPath: elem.ImportPath,
		}, nil
	}
	if _, ok := enums[udtName]; ok {
		return schema.GoType{Name: enumTypeName(udtName), IsEnum: true}, nil
	}
	return mapScalarType(udtName)
}

func mapScalarType(udtName string) (schema.GoType, error) {
	switch strings.ToLower(udtName) {
	case "int2", "int4", "int8", "serial", "bigserial", "int":
		return schema.GoType{Name: "int64"}, nil
	case "bool":
		return schema.GoType{Name: "bool"}, nil
	case "text", "varchar", "bpchar", "name":
		return schema.GoType{Name: "string"}, nil
	case "timestamptz", "timestamp", "date":
		return schema.GoType{Name: "time.Time"}, nil
	case "uuid":
		return schema.GoType{Name: "uuid.UUID", ImportPath: "github.com/google/uuid"}, nil
	case "json", "jsonb":
		return schema.GoType{Name: "json.RawMessage", ImportPath: "encoding/json"}, nil
	case "numeric", "decimal":
		return schema.GoType{Name: "decimal.Decimal", ImportPath: "github.com/shopspring/decimal"}, nil
	case "bytea":
		return schema.GoType{Name: "[]byte"}, nil
	case "float4", "float8":
		return schema.GoType{Name: "float64"}, nil
	default:
		return schema.GoType{}, fmt.Errorf("unsupported postgres type %q", udtName)
	}
}

func enumTypeName(pgName string) string {
	parts := strings.Split(pgName, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}
