package query

type colVal struct {
	name string
	val  any
}

// InsertBuilder builds INSERT statements.
type InsertBuilder struct {
	with     *WithBuilder
	prefix   []fragment
	suffix   []fragment
	table    Table
	sets     []colVal
	fromSel  *SelectBuilder
	returns  []fragment
	conflict *onConflictClause
}

// Insert starts an INSERT into table.
func Insert(t Table) *InsertBuilder {
	return &InsertBuilder{table: t}
}

// Prefix injects SQL before INSERT.
func (b *InsertBuilder) Prefix(sql string, args ...any) *InsertBuilder {
	b.prefix = append(b.prefix, Raw{sql: sql, args: args})
	return b
}

// Suffix injects SQL after the statement.
func (b *InsertBuilder) Suffix(sql string, args ...any) *InsertBuilder {
	b.suffix = append(b.suffix, Raw{sql: sql, args: args})
	return b
}

// Set adds or overwrites a column value.
func (b *InsertBuilder) Set(col Column, val any) *InsertBuilder {
	b.fromSel = nil
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
func (b *InsertBuilder) SetAny(column string, val any) *InsertBuilder {
	b.fromSel = nil
	for i, s := range b.sets {
		if s.name == column {
			b.sets[i].val = val
			return b
		}
	}
	b.sets = append(b.sets, colVal{name: column, val: val})
	return b
}

// Returning adds a RETURNING clause.
func (b *InsertBuilder) Returning(cols ...Column) *InsertBuilder {
	for _, col := range cols {
		b.returns = append(b.returns, columnFrag{col: col})
	}
	return b
}

// FromSelect uses INSERT ... SELECT instead of VALUES.
func (b *InsertBuilder) FromSelect(s *SelectBuilder) *InsertBuilder {
	b.sets = nil
	b.fromSel = s
	return b
}

type onConflictClause struct {
	cols    []string
	nothing bool
	updates []colVal
}

type OnConflictBuilder struct {
	insert *InsertBuilder
	clause onConflictClause
}

type assignment struct {
	col Column
	val any
}

// OnConflict starts an ON CONFLICT clause for the given columns.
func (b *InsertBuilder) OnConflict(cols ...Column) *OnConflictBuilder {
	oc := &OnConflictBuilder{insert: b}
	for _, col := range cols {
		oc.clause.cols = append(oc.clause.cols, col.SQLName())
	}
	return oc
}

// DoNothing sets ON CONFLICT DO NOTHING.
func (oc *OnConflictBuilder) DoNothing() *InsertBuilder {
	oc.clause.nothing = true
	oc.insert.conflict = &oc.clause
	return oc.insert
}

// DoUpdate sets ON CONFLICT DO UPDATE assignments.
func (oc *OnConflictBuilder) DoUpdate(assigns ...assignment) *InsertBuilder {
	for _, a := range assigns {
		name := a.col.SQLName()
		found := false
		for i, u := range oc.clause.updates {
			if u.name == name {
				oc.clause.updates[i].val = a.val
				found = true
				break
			}
		}
		if !found {
			oc.clause.updates = append(oc.clause.updates, colVal{name: name, val: a.val})
		}
	}
	oc.insert.conflict = &oc.clause
	return oc.insert
}

// Assign pairs a column with a value for ON CONFLICT DO UPDATE.
func Assign(col Column, val any) assignment {
	return assignment{col: col, val: val}
}

// Build renders SQL and arguments.
func (b *InsertBuilder) Build() (string, []any) {
	c := NewCompiler()
	b.compile(c)
	return c.SQL(), c.Args()
}

func (b *InsertBuilder) compile(c *Compiler) {
	if b.with != nil {
		b.with.compilePrefix(c)
	}
	for _, p := range b.prefix {
		p.compile(c)
	}
	c.Write("INSERT INTO ")
	tableFrag{table: b.table}.compile(c)
	if b.fromSel != nil {
		c.Write(" ")
		b.fromSel.compile(c)
	} else {
		c.Write(" (")
		for i, s := range b.sets {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(s.name)
		}
		c.Write(") VALUES (")
		for i, s := range b.sets {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(c.Arg(s.val))
		}
		c.Write(")")
	}
	if b.conflict != nil {
		c.Write(" ON CONFLICT (")
		for i, col := range b.conflict.cols {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(col)
		}
		c.Write(")")
		if b.conflict.nothing {
			c.Write(" DO NOTHING")
		} else {
			c.Write(" DO UPDATE SET ")
			for i, u := range b.conflict.updates {
				if i > 0 {
					c.Write(", ")
				}
				c.Write(u.name)
				c.Write(" = ")
				c.Write(c.Arg(u.val))
			}
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
