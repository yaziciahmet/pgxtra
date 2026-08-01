package query

type onConflictClause struct {
	cols    []string
	nothing bool
	updates []colVal
}

type OnConflictBuilder struct {
	insert *InsertBuilder
	clause onConflictClause
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

type assignment struct {
	col Column
	val any
}

// Assign pairs a column with a value for ON CONFLICT DO UPDATE.
func Assign(col Column, val any) assignment {
	return assignment{col: col, val: val}
}
