package query

// DeleteBuilder builds DELETE statements.
type DeleteBuilder struct {
	with    *WithBuilder
	prefix  []fragment
	suffix  []fragment
	table   Table
	wheres  []Expr
	returns []fragment
}

// Delete starts a DELETE from table.
func Delete(t Table) *DeleteBuilder {
	return &DeleteBuilder{table: t}
}

// Prefix injects SQL before DELETE.
func (b *DeleteBuilder) Prefix(sql string, args ...any) *DeleteBuilder {
	b.prefix = append(b.prefix, Raw{sql: sql, args: args})
	return b
}

// Suffix injects SQL after the statement.
func (b *DeleteBuilder) Suffix(sql string, args ...any) *DeleteBuilder {
	b.suffix = append(b.suffix, Raw{sql: sql, args: args})
	return b
}

// Where adds an AND condition.
func (b *DeleteBuilder) Where(e Expr) *DeleteBuilder {
	b.wheres = append(b.wheres, e)
	return b
}

// Returning adds a RETURNING clause.
func (b *DeleteBuilder) Returning(cols ...Column) *DeleteBuilder {
	for _, col := range cols {
		b.returns = append(b.returns, columnFrag{col: col})
	}
	return b
}

// Build renders SQL and arguments.
func (b *DeleteBuilder) Build() (string, []any) {
	return build(NewCompiler(), b)
}

func (b *DeleteBuilder) compile(c *Compiler) {
	compilePrefix(c, b.with, b.prefix)
	c.Write("DELETE FROM ")
	tableFrag{table: b.table}.compile(c)
	compileWhere(c, b.wheres)
	compileReturning(c, b.returns)
	compileSuffix(c, b.suffix)
}
