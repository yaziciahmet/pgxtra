package query_test

import (
	"testing"

	"github.com/yaziciahmet/typg/internal/query"
	"github.com/yaziciahmet/typg/internal/query/querytest"
)

func TestDeleteWhereReturning(t *testing.T) {
	sql, args := query.Delete(querytest.TPosts).
		Where(query.Eq(querytest.TPostPublished, false)).
		Where(query.Lte(querytest.TPostUserID, int64(100))).
		Returning(querytest.TPostUserID).
		Build()

	want := "DELETE FROM posts WHERE posts.published = $1 AND posts.user_id <= $2 RETURNING posts.user_id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != false || args[1] != int64(100) {
		t.Fatalf("args = %v", args)
	}
}

func TestDeleteWithCTE(t *testing.T) {
	stale := query.Select(querytest.TPostUserID).
		From(querytest.TPosts).
		Where(query.Eq(querytest.TPostPublished, false))

	sql, args := query.With("stale").
		As(stale).
		Delete(querytest.TUsers).
		Where(query.In(querytest.TUsersID, int64(1), int64(2))).
		Prefix("/* purge */ ").
		Build()

	want := "WITH stale AS (SELECT posts.user_id FROM posts WHERE posts.published = $1) /* purge */ DELETE FROM users WHERE users.id IN ($2, $3)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 || args[0] != false {
		t.Fatalf("args = %v", args)
	}
}
