package query_test

import (
	"testing"

	"github.com/yaziciahmet/pgxtra/query"
	"github.com/yaziciahmet/pgxtra/query/querytest"
)

func TestInsertSetOverwrite(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(querytest.TUsersName, "first").
		Set(querytest.TUsersName, "second").
		Set(querytest.TUsersID, int64(1)).
		Build()

	want := "INSERT INTO users (name, id) VALUES ($1, $2)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != "second" || args[1] != int64(1) {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertReturning(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(querytest.TUsersName, "alice").
		Returning(querytest.TUsersID).
		Build()

	want := "INSERT INTO users (name) VALUES ($1) RETURNING users.id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertOnConflictDoUpdate(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(querytest.TUsersEmail, "a@b.com").
		Set(querytest.TUsersName, "alice").
		OnConflict(querytest.TUsersEmail).
		DoUpdate(
			query.Assign(querytest.TUsersName, "alice"),
		).
		Returning(querytest.TUsersID).
		Build()

	want := "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = $3 RETURNING users.id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertFromSelectClearsSets(t *testing.T) {
	src := query.Select(querytest.TUsersID).From(querytest.TUsers)
	sql, args := query.Insert(querytest.TUsers).
		Set(querytest.TUsersName, "ignored").
		FromSelect(src).
		Build()

	want := "INSERT INTO users SELECT users.id FROM users"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertSetClearsFromSelect(t *testing.T) {
	src := query.Select(querytest.TUsersID).From(querytest.TUsers)
	sql, args := query.Insert(querytest.TUsers).
		FromSelect(src).
		Set(querytest.TUsersName, "alice").
		Build()

	want := "INSERT INTO users (name) VALUES ($1)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %v", args)
	}
}
