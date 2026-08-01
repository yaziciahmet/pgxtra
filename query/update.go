package query

// UpdateBuilder builds UPDATE statements.
type UpdateBuilder struct {
	with    *WithBuilder
	prefix  []fragment
	suffix  []fragment
	table   Table
	sets    []colVal
	from    []string
	wheres  []Expr
	returns []fragment
}

// Update starts an UPDATE on table.
func Update(t Table) *UpdateBuilder {
	return &UpdateBuilder{table: t}
}

// Prefix injects SQL before UPDATE.
func (b *UpdateBuilder) Prefix(sql string, args ...any) *UpdateBuilder {
	b.prefix = append(b.prefix, Raw{sql: sql, args: args})
	return b
}

// Suffix injects SQL after the statement.
func (b *UpdateBuilder) Suffix(sql string, args ...any) *UpdateBuilder {
	b.suffix = append(b.suffix, Raw{sql: sql, args: args})
	return b
}

// Set adds or overwrites a column assignment.
func (b *UpdateBuilder) Set(col Column, val any) *UpdateBuilder {
	b.sets = upsertColVal(b.sets, col.SQLName(), colVal{name: col.SQLName(), val: val})
	return b
}

// SetAny sets a column by name.
func (b *UpdateBuilder) SetAny(column string, val any) *UpdateBuilder {
	b.sets = upsertColVal(b.sets, column, colVal{name: column, val: val})
	return b
}

// SetExpr sets col = sql, with ? placeholders bound to args.
func (b *UpdateBuilder) SetExpr(col Column, sql string, args ...any) *UpdateBuilder {
	b.sets = upsertColVal(b.sets, col.SQLName(), colVal{name: col.SQLName(), expr: sql, args: args})
	return b
}

// Where adds an AND condition.
func (b *UpdateBuilder) Where(e Expr) *UpdateBuilder {
	b.wheres = append(b.wheres, e)
	return b
}

// Returning adds a RETURNING clause.
func (b *UpdateBuilder) Returning(cols ...Column) *UpdateBuilder {
	for _, col := range cols {
		b.returns = append(b.returns, columnFrag{col: col})
	}
	return b
}

// From adds UPDATE ... FROM tables (postgres).
func (b *UpdateBuilder) From(tables ...string) *UpdateBuilder {
	b.from = append(b.from, tables...)
	return b
}

// Build renders SQL and arguments.
func (b *UpdateBuilder) Build() (string, []any) {
	return build(NewCompiler(), b)
}

func (b *UpdateBuilder) compile(c *Compiler) {
	compilePrefix(c, b.with, b.prefix)
	c.Write("UPDATE ")
	tableFrag{table: b.table}.compile(c)
	c.Write(" SET ")
	for i, s := range b.sets {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(s.name)
		c.Write(" = ")
		s.compileRHS(c)
	}
	if len(b.from) > 0 {
		c.Write(" FROM ")
		for i, t := range b.from {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(t)
		}
	}
	compileWhere(c, b.wheres)
	compileReturning(c, b.returns)
	compileSuffix(c, b.suffix)
}
