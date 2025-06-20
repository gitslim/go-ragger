package util

import (
	"context"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
)

func UserFromContext(ctx context.Context) (*sqlc.User, bool) {
	user, ok := ctx.Value(UserKey).(*sqlc.User)
	return user, ok
}

func ContextWithUser(ctx context.Context, user *sqlc.User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}
