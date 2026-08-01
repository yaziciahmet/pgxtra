package query_test

import (
	"testing"
	"time"

	"github.com/yaziciahmet/typg/internal/query"
	"github.com/yaziciahmet/typg/internal/query/querytest"
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

func TestCompareValue(t *testing.T) {
	cutoff := time.Unix(100, 0)
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(query.Eq(querytest.TUsersID, int64(42))).
		Where(query.Gt(querytest.TUsersLastLoginAt, cutoff)).
		Where(query.IsNull(querytest.TUsersEmail)).
		Build()

	want := "SELECT users.id FROM users WHERE users.id = $1 AND users.last_login_at > $2 AND users.email IS NULL"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 || args[0] != int64(42) || args[1] != cutoff {
		t.Fatalf("args = %v", args)
	}
}

func TestContains(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(query.Contains(query.Col("users", "tags"), "fraud")).
		Build()

	want := "SELECT users.id FROM users WHERE users.tags @> ARRAY[$1]"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "fraud" {
		t.Fatalf("args = %v", args)
	}
}

func TestCompareColumn(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Join(querytest.TPosts, query.EqCol(querytest.TUsersID, querytest.TPostUserID)).
		Where(query.NotNull(querytest.TUsersEmail)).
		Build()

	want := "SELECT users.id FROM users INNER JOIN posts ON users.id = posts.user_id WHERE users.email IS NOT NULL"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestNeGteLt(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(query.Ne(querytest.TUsersStatus, "deleted")).
		Where(query.Gte(querytest.TUsersID, int64(10))).
		Where(query.Lt(querytest.TUsersID, int64(100))).
		Build()

	want := "SELECT users.id FROM users WHERE users.status != $1 AND users.id >= $2 AND users.id < $3"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != "deleted" || args[1] != int64(10) || args[2] != int64(100) {
		t.Fatalf("args = %v", args)
	}
}
