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
	sql, args := query.Select(querytest.TUsersName, querytest.TPostTitle).
		From(querytest.TUsers).
		Join(querytest.TPosts, query.Eq(querytest.TUsersID, querytest.TPostUserID)).
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

func TestSelectDistinct(t *testing.T) {
	sql, args := query.Select(querytest.TUsersName).
		Distinct().
		From(querytest.TUsers).
		Build()

	want := "SELECT DISTINCT users.name FROM users"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
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

func TestSelectLeftJoin(t *testing.T) {
	sql, args := query.Select(querytest.TUsersName, querytest.TPostTitle).
		From(querytest.TUsers).
		LeftJoin(querytest.TPosts, query.Eq(querytest.TUsersID, querytest.TPostUserID)).
		OrderByAsc(querytest.TUsersName).
		Build()

	want := "SELECT users.name, posts.title FROM users LEFT JOIN posts ON users.id = posts.user_id ORDER BY users.name"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectAndOr(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(query.And(
			query.In(querytest.TUsersStatus, "active", "pending"),
			query.Or(
				query.Like(querytest.TUsersName, "a%"),
				query.Eq(querytest.TUsersEmail, "admin@example.com"),
			),
		)).
		Build()

	want := "SELECT users.id FROM users WHERE (users.status IN ($1, $2)) AND ((users.name LIKE $3) OR (users.email = $4))"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 4 {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectRightCrossJoin(t *testing.T) {
	sql, args := query.Select(querytest.TUsersName, querytest.TPostTitle).
		From(querytest.TUsers).
		RightJoin(querytest.TPosts, query.Eq(querytest.TUsersID, querytest.TPostUserID)).
		CrossJoin(query.Named("archived")).
		Build()

	want := "SELECT users.name, posts.title FROM users RIGHT JOIN posts ON users.id = posts.user_id CROSS JOIN archived"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}
