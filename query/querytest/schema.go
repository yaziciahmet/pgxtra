package querytest

import (
	"time"

	"github.com/yaziciahmet/pgxtra/query"
)

type Users struct{}

func (Users) SQLName() string { return "users" }
func (Users) Alias() string  { return "users" }

type UserID struct{ users Users }

func (UserID) SQLName() string     { return "id" }
func (UserID) Qualifier() string   { return "users.id" }
func (UserID) Table() query.Table  { return Users{} }
func (UserID) Eq(v int64) query.Expr {
	return query.RawExpr("users.id = ?", v)
}

type UserName struct{ users Users }

func (UserName) SQLName() string    { return "name" }
func (UserName) Qualifier() string  { return "users.name" }
func (UserName) Table() query.Table { return Users{} }
func (UserName) Eq(v string) query.Expr {
	return query.RawExpr("users.name = ?", v)
}

type UserEmail struct{ users Users }

func (UserEmail) SQLName() string    { return "email" }
func (UserEmail) Qualifier() string  { return "users.email" }
func (UserEmail) Table() query.Table { return Users{} }
func (UserEmail) Eq(v string) query.Expr {
	return query.RawExpr("users.email = ?", v)
}

type UserStatus struct{ users Users }

func (UserStatus) SQLName() string    { return "status" }
func (UserStatus) Qualifier() string  { return "users.status" }
func (UserStatus) Table() query.Table { return Users{} }
func (UserStatus) Eq(v string) query.Expr {
	return query.RawExpr("users.status = ?", v)
}

type UserLastLoginAt struct{ users Users }

func (UserLastLoginAt) SQLName() string    { return "last_login_at" }
func (UserLastLoginAt) Qualifier() string  { return "users.last_login_at" }
func (UserLastLoginAt) Table() query.Table { return Users{} }
func (UserLastLoginAt) Lt(v time.Time) query.Expr {
	return query.RawExpr("users.last_login_at < ?", v)
}

type UserUpdatedAt struct{ users Users }

func (UserUpdatedAt) SQLName() string    { return "updated_at" }
func (UserUpdatedAt) Qualifier() string  { return "users.updated_at" }
func (UserUpdatedAt) Table() query.Table { return Users{} }

var (
	TUsers            = Users{}
	TUsersID          = UserID{}
	TUsersName        = UserName{}
	TUsersEmail       = UserEmail{}
	TUsersStatus      = UserStatus{}
	TUsersLastLoginAt = UserLastLoginAt{}
	TUsersUpdatedAt   = UserUpdatedAt{}
)
