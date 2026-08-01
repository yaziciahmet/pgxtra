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
