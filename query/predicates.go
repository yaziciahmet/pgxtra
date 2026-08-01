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
