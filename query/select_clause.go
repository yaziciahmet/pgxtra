package query

type orderClause struct {
	expr string
	desc bool
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
