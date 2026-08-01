package query_test

import (
	"testing"
	"time"

	"github.com/yaziciahmet/pgxtra/query"
	"github.com/yaziciahmet/pgxtra/query/querytest"
)

func TestCTEUpdateFrom(t *testing.T) {
	cutoff := time.Unix(100, 0)
	now := time.Unix(200, 0)

	staleSel := query.Select(querytest.TUsersID).
		From(querytest.TUsers).
		Where(querytest.TUsersLastLoginAt.Lt(cutoff)).
		Where(querytest.TUsersStatus.Eq("active"))

	sql, args := query.With("stale").
		As(staleSel).
		Update(querytest.TUsers).
		Set(querytest.TUsersStatus, "inactive").
		Set(querytest.TUsersUpdatedAt, now).
		From("stale").
		Where(query.EqCol(querytest.TUsersID, query.Col("stale", "id"))).
		Build()

	want := "WITH stale AS (SELECT users.id FROM users WHERE users.last_login_at < $1 AND users.status = $2) UPDATE users SET status = $3, updated_at = $4 FROM stale WHERE users.id = stale.id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 4 {
		t.Fatalf("args = %v", args)
	}
}

func TestCTEModifyingInsertFromSelect(t *testing.T) {
	cutoff := time.Unix(100, 0)

	moved := query.Update(querytest.TPosts).
		Set(querytest.TPostPublished, false).
		Where(query.RawExpr("posts.created_at < ?", cutoff)).
		Returning(querytest.TPostUserID)

	src := query.Select(query.Col("moved", "user_id")).From(query.Named("moved"))

	sql, args := query.With("moved").
		As(moved).
		Insert(querytest.TUsers).
		FromSelect(src).
		OnConflict(querytest.TUsersID).
		DoNothing().
		Build()

	wantPrefix := "WITH moved AS (UPDATE posts SET published = $1 WHERE posts.created_at < $2 RETURNING posts.user_id) INSERT INTO users SELECT moved.user_id FROM moved ON CONFLICT (id) DO NOTHING"
	if sql != wantPrefix {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v", args)
	}
}

func TestCTEMultiAsNextSelect(t *testing.T) {
	active := query.Select(querytest.TUsersID, querytest.TUsersName).
		From(querytest.TUsers).
		Where(query.Eq(querytest.TUsersStatus, "active"))

	recent := query.Select(querytest.TPostTitle).
		From(querytest.TPosts).
		Where(query.Eq(querytest.TPostPublished, true))

	sql, args := query.With("active_users").
		As(active).
		AsNext("recent_posts", recent).
		Select(querytest.TUsersName, querytest.TPostTitle).
		From(querytest.TUsers).
		Join(querytest.TPosts, query.EqCol(querytest.TUsersID, querytest.TPostUserID)).
		Build()

	want := "WITH active_users AS (SELECT users.id, users.name FROM users WHERE users.status = $1), recent_posts AS (SELECT posts.title FROM posts WHERE posts.published = $2) SELECT users.name, posts.title FROM users INNER JOIN posts ON users.id = posts.user_id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 || args[0] != "active" || args[1] != true {
		t.Fatalf("args = %v", args)
	}
}

func TestCTEInsertMain(t *testing.T) {
	src := query.Select(querytest.TUsersID, querytest.TUsersEmail).
		From(querytest.TUsers).
		Where(query.NotNull(querytest.TUsersEmail))

	sql, args := query.With("src").
		As(src).
		Insert(querytest.TUsers).
		FromSelect(query.Select(query.Col("src", "id"), query.Col("src", "email")).From(query.Named("src"))).
		Build()

	want := "WITH src AS (SELECT users.id, users.email FROM users WHERE users.email IS NOT NULL) INSERT INTO users SELECT src.id, src.email FROM src"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}
