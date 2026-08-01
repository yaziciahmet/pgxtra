package query

import (
	"strconv"
	"strings"
)

// Compiler renders SQL and collects placeholder arguments.
type Compiler struct {
	buf   strings.Builder
	args  []any
	index int
}

// NewCompiler returns a compiler for PostgreSQL ($1, $2, ...).
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Write appends literal SQL.
func (c *Compiler) Write(sql string) {
	c.buf.WriteString(sql)
}

// Arg registers a value and returns its PostgreSQL placeholder.
func (c *Compiler) Arg(v any) string {
	c.index++
	c.args = append(c.args, v)
	return placeholder(c.index)
}

func placeholder(n int) string {
	return "$" + strconv.Itoa(n)
}

// SQL returns the rendered query string.
func (c *Compiler) SQL() string {
	return c.buf.String()
}

// Args returns collected arguments in placeholder order.
func (c *Compiler) Args() []any {
	return c.args
}
