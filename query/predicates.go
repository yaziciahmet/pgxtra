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

func compareVal(col Column, op string, val any) Expr {
	return compareExpr{left: col, op: op, rightVal: val}
}

func compareCol(col Column, op string, other Column) Expr {
	return compareExpr{left: col, op: op, rightCol: other, isCol: true}
}

// Eq builds col = val.
func Eq(col Column, val any) Expr { return compareVal(col, "=", val) }

// EqCol builds col = other.
func EqCol(col, other Column) Expr { return compareCol(col, "=", other) }

// Ne builds col != val.
func Ne(col Column, val any) Expr { return compareVal(col, "!=", val) }

// NeCol builds col != other.
func NeCol(col, other Column) Expr { return compareCol(col, "!=", other) }

// Gt builds col > val.
func Gt(col Column, val any) Expr { return compareVal(col, ">", val) }

// GtCol builds col > other.
func GtCol(col, other Column) Expr { return compareCol(col, ">", other) }

// Gte builds col >= val.
func Gte(col Column, val any) Expr { return compareVal(col, ">=", val) }

// GteCol builds col >= other.
func GteCol(col, other Column) Expr { return compareCol(col, ">=", other) }

// Lt builds col < val.
func Lt(col Column, val any) Expr { return compareVal(col, "<", val) }

// LtCol builds col < other.
func LtCol(col, other Column) Expr { return compareCol(col, "<", other) }

// Lte builds col <= val.
func Lte(col Column, val any) Expr { return compareVal(col, "<=", val) }

// LteCol builds col <= other.
func LteCol(col, other Column) Expr { return compareCol(col, "<=", other) }

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
