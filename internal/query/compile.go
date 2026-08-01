package query

import (
	"strconv"
	"strings"
)

// fragment is anything that renders SQL through a Compiler.
type fragment interface {
	compile(c *Compiler)
}

// Expr is a boolean SQL expression for WHERE, ON, HAVING, etc.
type Expr interface {
	fragment
}

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
	return "$" + strconv.Itoa(c.index)
}

// SQL returns the rendered query string.
func (c *Compiler) SQL() string {
	return c.buf.String()
}

// Args returns collected arguments in placeholder order.
func (c *Compiler) Args() []any {
	return c.args
}

func writeInterpolated(c *Compiler, sql string, args []any) {
	argIdx := 0
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' && argIdx < len(args) {
			c.Write(c.Arg(args[argIdx]))
			argIdx++
			continue
		}
		c.Write(string(sql[i]))
	}
}

// Raw is a SQL fragment with ? placeholders replaced by numbered args.
type Raw struct {
	sql  string
	args []any
}

// RawExpr builds an expression from a SQL template and arguments.
func RawExpr(sql string, args ...any) Expr {
	return Raw{sql: sql, args: args}
}

func (r Raw) compile(c *Compiler) {
	writeInterpolated(c, r.sql, r.args)
}

type columnFrag struct {
	col Column
}

func (f columnFrag) compile(c *Compiler) {
	c.Write(f.col.Qualifier())
}

type tableFrag struct {
	table Table
}

func (f tableFrag) compile(c *Compiler) {
	c.Write(f.table.SQLName())
	if alias := f.table.Alias(); alias != "" && alias != f.table.SQLName() {
		c.Write(" AS ")
		c.Write(alias)
	}
}

// colVal is a column assignment: either a bound value or a SQL expression.
type colVal struct {
	name string
	val  any
	expr string
	args []any
}

func (cv colVal) compileRHS(c *Compiler) {
	if cv.expr != "" {
		writeInterpolated(c, cv.expr, cv.args)
		return
	}
	c.Write(c.Arg(cv.val))
}

func upsertColVal(sets []colVal, name string, cv colVal) []colVal {
	for i, s := range sets {
		if s.name == name {
			sets[i] = cv
			return sets
		}
	}
	return append(sets, cv)
}

func compilePrefix(c *Compiler, with *WithBuilder, prefix []fragment) {
	if with != nil {
		with.compilePrefix(c)
	}
	for _, p := range prefix {
		p.compile(c)
	}
}

func compileSuffix(c *Compiler, suffix []fragment) {
	for _, p := range suffix {
		p.compile(c)
	}
}

func compileWhere(c *Compiler, wheres []Expr) {
	if len(wheres) == 0 {
		return
	}
	c.Write(" WHERE ")
	for i, w := range wheres {
		if i > 0 {
			c.Write(" AND ")
		}
		w.compile(c)
	}
}

func compileReturning(c *Compiler, returns []fragment) {
	if len(returns) == 0 {
		return
	}
	c.Write(" RETURNING ")
	for i, col := range returns {
		if i > 0 {
			c.Write(", ")
		}
		col.compile(c)
	}
}

func build(c *Compiler, stmt fragment) (string, []any) {
	stmt.compile(c)
	return c.SQL(), c.Args()
}
