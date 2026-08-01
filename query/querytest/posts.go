package querytest

import "github.com/yaziciahmet/pgxtra/query"

type Posts struct{}

func (Posts) SQLName() string { return "posts" }
func (Posts) Alias() string  { return "posts" }

type PostUserID struct{ posts Posts }

func (PostUserID) SQLName() string     { return "user_id" }
func (PostUserID) Qualifier() string   { return "posts.user_id" }
func (PostUserID) Table() query.Table  { return Posts{} }
func (PostUserID) Eq(v int64) query.Expr {
	return query.RawExpr("posts.user_id = ?", v)
}

type PostTitle struct{ posts Posts }

func (PostTitle) SQLName() string    { return "title" }
func (PostTitle) Qualifier() string  { return "posts.title" }
func (PostTitle) Table() query.Table { return Posts{} }
func (PostTitle) Eq(v string) query.Expr {
	return query.RawExpr("posts.title = ?", v)
}

type PostPublished struct{ posts Posts }

func (PostPublished) SQLName() string    { return "published" }
func (PostPublished) Qualifier() string  { return "posts.published" }
func (PostPublished) Table() query.Table { return Posts{} }
func (PostPublished) Eq(v bool) query.Expr {
	return query.RawExpr("posts.published = ?", v)
}

var (
	TPosts        = Posts{}
	TPostUserID   = PostUserID{}
	TPostTitle    = PostTitle{}
	TPostPublished = PostPublished{}
)
