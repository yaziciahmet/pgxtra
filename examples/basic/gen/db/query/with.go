package query

type compileStmt interface {
	compile(c *Compiler)
}

type cteEntry struct {
	name string
	stmt compileStmt
}

// WithBuilder builds WITH (CTE) clauses.
type WithBuilder struct {
	ctes        []cteEntry
	pendingName string
}

// With starts a CTE chain. Pass the first CTE name; call As next.
func With(name string) *WithBuilder {
	return &WithBuilder{pendingName: name}
}

// As adds the pending CTE name with the given statement.
func (w *WithBuilder) As(stmt compileStmt) *WithBuilder {
	if w.pendingName == "" {
		return w
	}
	w.ctes = append(w.ctes, cteEntry{name: w.pendingName, stmt: stmt})
	w.pendingName = ""
	return w
}

// AsNext names and adds another CTE.
func (w *WithBuilder) AsNext(name string, stmt compileStmt) *WithBuilder {
	w.ctes = append(w.ctes, cteEntry{name: name, stmt: stmt})
	return w
}

func (w *WithBuilder) compilePrefix(c *Compiler) {
	if len(w.ctes) == 0 {
		return
	}
	c.Write("WITH ")
	for i, e := range w.ctes {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(e.name)
		c.Write(" AS (")
		e.stmt.compile(c)
		c.Write(")")
	}
	c.Write(" ")
}

// Select starts the main SELECT after CTEs.
func (w *WithBuilder) Select(cols ...Column) *SelectBuilder {
	s := Select(cols...)
	s.with = w
	return s
}

// Insert starts the main INSERT after CTEs.
func (w *WithBuilder) Insert(t Table) *InsertBuilder {
	b := Insert(t)
	b.with = w
	return b
}

// Update starts the main UPDATE after CTEs.
func (w *WithBuilder) Update(t Table) *UpdateBuilder {
	b := Update(t)
	b.with = w
	return b
}

// Delete starts the main DELETE after CTEs.
func (w *WithBuilder) Delete(t Table) *DeleteBuilder {
	b := Delete(t)
	b.with = w
	return b
}
