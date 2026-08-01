package query_test

import (
	"testing"

	"github.com/yaziciahmet/pgxtra/query"
	"github.com/yaziciahmet/pgxtra/query/querytest"
)

func TestSelectBasic(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID, querytest.TUsersName).
		From(querytest.TUsers).
		Where(querytest.TUsersID.Eq(42)).
		Where(querytest.TUsersName.Eq("alice")).
		Build()

	want := "SELECT users.id, users.name FROM users WHERE users.id = $1 AND users.name = $2"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 || args[0] != int64(42) || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectPrefixSuffix(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		Prefix("/* hint */ ").
		From(querytest.TUsers).
		Suffix(" FOR UPDATE").
		Build()

	want := "/* hint */ SELECT users.id FROM users FOR UPDATE"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectJoin(t *testing.T) {
	on := query.RawExpr("users.id = posts.user_id")
	sql, args := query.Select(querytest.TUsersName, querytest.TPostTitle).
		From(querytest.TUsers).
		Join(querytest.TPosts, on).
		Where(querytest.TPostPublished.Eq(true)).
		Build()

	want := "SELECT users.name, posts.title FROM users INNER JOIN posts ON users.id = posts.user_id WHERE posts.published = $1"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != true {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectClauses(t *testing.T) {
	sql, args := query.Select(querytest.TUsersName).
		From(querytest.TUsers).
		GroupBy(querytest.TUsersName).
		Having(query.RawExpr("count(*) > ?", 1)).
		OrderByDesc(querytest.TUsersName).
		Limit(10).
		Offset(20).
		Build()

	want := "SELECT users.name FROM users GROUP BY users.name HAVING count(*) > $1 ORDER BY users.name DESC LIMIT 10 OFFSET 20"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Fatalf("args = %v", args)
	}
}
