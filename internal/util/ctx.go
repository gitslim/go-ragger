package util

import (
	"context"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
)

func UserFromContext(ctx context.Context) (*sqlc.User, bool) {
	user, ok := ctx.Value(UserKey).(*sqlc.User)
	return user, ok // TODO: refactor return?
}

func ContextWithUser(ctx context.Context, user *sqlc.User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

func IsAuthenticated(ctx context.Context) bool {
	_, ok := UserFromContext(ctx)
	return ok
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(RequestIDKey).(string)
	return reqID, ok
}
