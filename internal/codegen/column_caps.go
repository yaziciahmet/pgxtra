package codegen

import (
	"strings"

	"github.com/yaziciahmet/pgxtra/internal/schema"
)

type colCaps struct {
	order    bool // Gt, Gte, Lt, Lte
	in       bool // In
	like     bool // Like, ILike
	contains bool // Contains (array columns)
}

func columnCapabilities(col schema.Column) colCaps {
	gt := col.Go
	if strings.HasPrefix(gt.Name, "[]") {
		return colCaps{contains: true}
	}
	switch gt.Name {
	case "json.RawMessage":
		return colCaps{in: true}
	case "bool":
		return colCaps{in: true}
	case "string":
		return colCaps{order: true, in: true, like: true}
	default:
		return colCaps{order: true, in: true}
	}
}
