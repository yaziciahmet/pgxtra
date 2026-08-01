package query_test

import (
	"testing"

	"github.com/yaziciahmet/pgxtra/query"
	"github.com/yaziciahmet/pgxtra/query/querytest"
)

func TestIn(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(query.In(querytest.TUsersID, int64(1), int64(2), int64(3))).
		Build()

	want := "SELECT users.id FROM users WHERE users.id IN ($1, $2, $3)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 {
		t.Fatalf("args = %v", args)
	}
}

func TestLikeILike(t *testing.T) {
	sql, args := query.Select(querytest.TUsersName).
		From(querytest.TUsers).
		Where(query.Like(querytest.TUsersName, "%ali%")).
		Where(query.ILike(querytest.TUsersEmail, "%@EXAMPLE.COM")).
		Build()

	want := "SELECT users.name FROM users WHERE users.name LIKE $1 AND users.email ILIKE $2"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != "%ali%" || args[1] != "%@EXAMPLE.COM" {
		t.Fatalf("args = %v", args)
	}
}
