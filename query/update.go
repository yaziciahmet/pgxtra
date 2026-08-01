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
	name := col.SQLName()
	for i, s := range b.sets {
		if s.name == name {
			b.sets[i].val = val
			return b
		}
	}
	b.sets = append(b.sets, colVal{name: name, val: val})
	return b
}

// SetAny sets a column by name.
func (b *UpdateBuilder) SetAny(column string, val any) *UpdateBuilder {
	for i, s := range b.sets {
		if s.name == column {
			b.sets[i].val = val
			return b
		}
	}
	b.sets = append(b.sets, colVal{name: column, val: val})
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
	c := NewCompiler()
	b.compile(c)
	return c.SQL(), c.Args()
}

func (b *UpdateBuilder) compile(c *Compiler) {
	if b.with != nil {
		b.with.compilePrefix(c)
	}
	for _, p := range b.prefix {
		p.compile(c)
	}
	c.Write("UPDATE ")
	tableFrag{table: b.table}.compile(c)
	c.Write(" SET ")
	for i, s := range b.sets {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(s.name)
		c.Write(" = ")
		c.Write(c.Arg(s.val))
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
	if len(b.wheres) > 0 {
		c.Write(" WHERE ")
		for i, w := range b.wheres {
			if i > 0 {
				c.Write(" AND ")
			}
			w.compile(c)
		}
	}
	if len(b.returns) > 0 {
		c.Write(" RETURNING ")
		for i, col := range b.returns {
			if i > 0 {
				c.Write(", ")
			}
			col.compile(c)
		}
	}
	for _, p := range b.suffix {
		p.compile(c)
	}
}
