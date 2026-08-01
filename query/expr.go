package query

// Expr is a SQL expression usable in WHERE, ON, HAVING, etc.
type Expr interface {
	fragment
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
