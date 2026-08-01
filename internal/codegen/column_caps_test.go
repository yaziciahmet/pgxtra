package codegen

import (
	"testing"

	"github.com/yaziciahmet/pgxtra/internal/schema"
)

func TestColumnCaps(t *testing.T) {
	tests := []struct {
		name string
		col  schema.Column
		want colCaps
	}{
		{
			name: "string",
			col:  schema.Column{Go: schema.GoType{Name: "string"}},
			want: colCaps{order: true, in: true, like: true},
		},
		{
			name: "array",
			col:  schema.Column{Go: schema.GoType{Name: "[]string"}},
			want: colCaps{contains: true},
		},
		{
			name: "json",
			col:  schema.Column{Go: schema.GoType{Name: "json.RawMessage", ImportPath: "encoding/json"}},
			want: colCaps{in: true},
		},
		{
			name: "bool",
			col:  schema.Column{Go: schema.GoType{Name: "bool"}},
			want: colCaps{in: true},
		},
		{
			name: "int64",
			col:  schema.Column{Go: schema.GoType{Name: "int64"}},
			want: colCaps{order: true, in: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := columnCapabilities(tt.col); got != tt.want {
				t.Fatalf("columnCaps() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
