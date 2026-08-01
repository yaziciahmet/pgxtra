package introspect_test

import (
	"testing"

	"github.com/yaziciahmet/pgxtra/internal/introspect"
	"github.com/yaziciahmet/pgxtra/internal/schema"
)

func TestResolveType(t *testing.T) {
	enums := map[string]schema.Enum{
		"user_status": {Name: "user_status", Values: []string{"active", "inactive"}},
	}

	tests := []struct {
		udt    string
		want   schema.GoType
		wantOK bool
	}{
		{"int8", schema.GoType{Name: "int64"}, true},
		{"bool", schema.GoType{Name: "bool"}, true},
		{"text", schema.GoType{Name: "string"}, true},
		{"_text", schema.GoType{Name: "[]string"}, true},
		{"_int8", schema.GoType{Name: "[]int64"}, true},
		{"_uuid", schema.GoType{Name: "[]uuid.UUID", ImportPath: "github.com/google/uuid"}, true},
		{"timestamptz", schema.GoType{Name: "time.Time"}, true},
		{"user_status", schema.GoType{Name: "UserStatus", IsEnum: true}, true},
		{"jsonb", schema.GoType{Name: "json.RawMessage", ImportPath: "encoding/json"}, true},
		{"geometry", schema.GoType{}, false},
	}
	for _, tc := range tests {
		got, err := introspect.ResolveType(tc.udt, enums)
		if tc.wantOK && err != nil {
			t.Fatalf("ResolveType(%q): %v", tc.udt, err)
		}
		if !tc.wantOK && err == nil {
			t.Fatalf("ResolveType(%q): want error", tc.udt)
		}
		if tc.wantOK && (got.Name != tc.want.Name || got.ImportPath != tc.want.ImportPath || got.IsEnum != tc.want.IsEnum) {
			t.Fatalf("ResolveType(%q) = %+v, want %+v", tc.udt, got, tc.want)
		}
	}
}
