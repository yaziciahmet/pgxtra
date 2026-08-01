package query

type andExpr struct {
	exprs []Expr
}

// And joins expressions with AND.
func And(exprs ...Expr) Expr {
	return andExpr{exprs: exprs}
}

func (a andExpr) compile(c *Compiler) {
	for i, e := range a.exprs {
		if i > 0 {
			c.Write(" AND ")
		}
		c.Write("(")
		e.compile(c)
		c.Write(")")
	}
}

type orExpr struct {
	exprs []Expr
}

// Or joins expressions with OR.
func Or(exprs ...Expr) Expr {
	return orExpr{exprs: exprs}
}

func (o orExpr) compile(c *Compiler) {
	for i, e := range o.exprs {
		if i > 0 {
			c.Write(" OR ")
		}
		c.Write("(")
		e.compile(c)
		c.Write(")")
	}
}

type inExpr struct {
	col  Column
	vals []any
}

// In builds col IN ($1, $2, ...).
func In(col Column, values ...any) Expr {
	return inExpr{col: col, vals: values}
}

func (e inExpr) compile(c *Compiler) {
	c.Write(e.col.Qualifier())
	c.Write(" IN (")
	for i, v := range e.vals {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.Arg(v))
	}
	c.Write(")")
}

type likeExpr struct {
	col     Column
	pattern string
	ilike   bool
}

// Like builds col LIKE pattern.
func Like(col Column, pattern string) Expr {
	return likeExpr{col: col, pattern: pattern}
}

// ILike builds col ILIKE pattern (postgres).
func ILike(col Column, pattern string) Expr {
	return likeExpr{col: col, pattern: pattern, ilike: true}
}

func (e likeExpr) compile(c *Compiler) {
	c.Write(e.col.Qualifier())
	if e.ilike {
		c.Write(" ILIKE ")
	} else {
		c.Write(" LIKE ")
	}
	c.Write(c.Arg(e.pattern))
}

type compareExpr struct {
	left     Column
	op       string
	rightCol Column
	rightVal any
	isCol    bool
}

func compare(col Column, op string, rhs any) Expr {
	if other, ok := rhs.(Column); ok {
		return compareExpr{left: col, op: op, rightCol: other, isCol: true}
	}
	return compareExpr{left: col, op: op, rightVal: rhs}
}

// Eq builds col = rhs. rhs may be a Column or a bound value.
func Eq(col Column, rhs any) Expr { return compare(col, "=", rhs) }

// Ne builds col != rhs. rhs may be a Column or a bound value.
func Ne(col Column, rhs any) Expr { return compare(col, "!=", rhs) }

// Gt builds col > rhs. rhs may be a Column or a bound value.
func Gt(col Column, rhs any) Expr { return compare(col, ">", rhs) }

// Gte builds col >= rhs. rhs may be a Column or a bound value.
func Gte(col Column, rhs any) Expr { return compare(col, ">=", rhs) }

// Lt builds col < rhs. rhs may be a Column or a bound value.
func Lt(col Column, rhs any) Expr { return compare(col, "<", rhs) }

// Lte builds col <= rhs. rhs may be a Column or a bound value.
func Lte(col Column, rhs any) Expr { return compare(col, "<=", rhs) }

func (e compareExpr) compile(c *Compiler) {
	c.Write(e.left.Qualifier())
	c.Write(" ")
	c.Write(e.op)
	c.Write(" ")
	if e.isCol {
		c.Write(e.rightCol.Qualifier())
	} else {
		c.Write(c.Arg(e.rightVal))
	}
}

type nullExpr struct {
	col Column
	neg bool
}

// IsNull builds col IS NULL.
func IsNull(col Column) Expr { return nullExpr{col: col} }

// NotNull builds col IS NOT NULL.
func NotNull(col Column) Expr { return nullExpr{col: col, neg: true} }

func (e nullExpr) compile(c *Compiler) {
	c.Write(e.col.Qualifier())
	if e.neg {
		c.Write(" IS NOT NULL")
	} else {
		c.Write(" IS NULL")
	}
}
