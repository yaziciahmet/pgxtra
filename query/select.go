package query

import "strconv"

type joinClause struct {
	kind  string
	table Table
	on    Expr
}

type orderClause struct {
	expr string
	desc bool
}

// SelectBuilder builds SELECT statements.
type SelectBuilder struct {
	with     *WithBuilder
	prefix   []fragment
	suffix   []fragment
	distinct bool
	cols     []fragment
	from     Table
	joins    []joinClause
	wheres   []Expr
	groupBy  []fragment
	having   []Expr
	orderBy  []orderClause
	limit    *int
	offset   *int
}

// Select starts a SELECT for the given columns.
func Select(cols ...Column) *SelectBuilder {
	s := &SelectBuilder{}
	for _, col := range cols {
		s.cols = append(s.cols, columnFrag{col: col})
	}
	return s
}

// Prefix injects SQL before SELECT.
func (s *SelectBuilder) Prefix(sql string, args ...any) *SelectBuilder {
	s.prefix = append(s.prefix, Raw{sql: sql, args: args})
	return s
}

// Suffix injects SQL after the statement.
func (s *SelectBuilder) Suffix(sql string, args ...any) *SelectBuilder {
	s.suffix = append(s.suffix, Raw{sql: sql, args: args})
	return s
}

// Distinct adds SELECT DISTINCT.
func (s *SelectBuilder) Distinct() *SelectBuilder {
	s.distinct = true
	return s
}

// From sets the FROM table.
func (s *SelectBuilder) From(t Table) *SelectBuilder {
	s.from = t
	return s
}

// Join adds an INNER JOIN.
func (s *SelectBuilder) Join(t Table, on Expr) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "INNER JOIN", table: t, on: on})
	return s
}

// LeftJoin adds a LEFT JOIN.
func (s *SelectBuilder) LeftJoin(t Table, on Expr) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "LEFT JOIN", table: t, on: on})
	return s
}

// Where adds an AND condition.
func (s *SelectBuilder) Where(e Expr) *SelectBuilder {
	s.wheres = append(s.wheres, e)
	return s
}

// GroupBy adds GROUP BY columns.
func (s *SelectBuilder) GroupBy(cols ...Column) *SelectBuilder {
	for _, col := range cols {
		s.groupBy = append(s.groupBy, columnFrag{col: col})
	}
	return s
}

// Having adds an AND HAVING condition.
func (s *SelectBuilder) Having(e Expr) *SelectBuilder {
	s.having = append(s.having, e)
	return s
}

// OrderByAsc adds ascending ORDER BY columns.
func (s *SelectBuilder) OrderByAsc(cols ...Column) *SelectBuilder {
	for _, col := range cols {
		s.orderBy = append(s.orderBy, orderClause{expr: col.Qualifier(), desc: false})
	}
	return s
}

// OrderByDesc adds descending ORDER BY columns.
func (s *SelectBuilder) OrderByDesc(cols ...Column) *SelectBuilder {
	for _, col := range cols {
		s.orderBy = append(s.orderBy, orderClause{expr: col.Qualifier(), desc: true})
	}
	return s
}

// Limit sets LIMIT.
func (s *SelectBuilder) Limit(n int) *SelectBuilder {
	s.limit = &n
	return s
}

// Offset sets OFFSET.
func (s *SelectBuilder) Offset(n int) *SelectBuilder {
	s.offset = &n
	return s
}

// Build renders SQL and arguments.
func (s *SelectBuilder) Build() (string, []any) {
	return build(NewCompiler(), s)
}

func (s *SelectBuilder) compile(c *Compiler) {
	compilePrefix(c, s.with, s.prefix)
	c.Write("SELECT ")
	if s.distinct {
		c.Write("DISTINCT ")
	}
	for i, col := range s.cols {
		if i > 0 {
			c.Write(", ")
		}
		col.compile(c)
	}
	if s.from != nil {
		c.Write(" FROM ")
		tableFrag{table: s.from}.compile(c)
	}
	for _, j := range s.joins {
		c.Write(" ")
		c.Write(j.kind)
		c.Write(" ")
		tableFrag{table: j.table}.compile(c)
		c.Write(" ON ")
		j.on.compile(c)
	}
	compileWhere(c, s.wheres)
	if len(s.groupBy) > 0 {
		c.Write(" GROUP BY ")
		for i, col := range s.groupBy {
			if i > 0 {
				c.Write(", ")
			}
			col.compile(c)
		}
	}
	if len(s.having) > 0 {
		c.Write(" HAVING ")
		for i, h := range s.having {
			if i > 0 {
				c.Write(" AND ")
			}
			h.compile(c)
		}
	}
	if len(s.orderBy) > 0 {
		c.Write(" ORDER BY ")
		for i, o := range s.orderBy {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(o.expr)
			if o.desc {
				c.Write(" DESC")
			}
		}
	}
	if s.limit != nil {
		c.Write(" LIMIT ")
		c.Write(strconv.Itoa(*s.limit))
	}
	if s.offset != nil {
		c.Write(" OFFSET ")
		c.Write(strconv.Itoa(*s.offset))
	}
	compileSuffix(c, s.suffix)
}
