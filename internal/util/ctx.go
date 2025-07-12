package util

import (
	"context"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
)

// UserFromContext returns the user from the context
func UserFromContext(ctx context.Context) (*sqlc.User, bool) {
	user, ok := ctx.Value(UserKey).(*sqlc.User)
	return user, ok // TODO: refactor return?
}

// ContextWithUser returns a new context with the user
func ContextWithUser(ctx context.Context, user *sqlc.User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

// IsAuthenticated returns true if the user is authenticated
func IsAuthenticated(ctx context.Context) bool {
	_, ok := UserFromContext(ctx)
	return ok
}

// RequestIDFromContext returns the request ID from the context
func RequestIDFromContext(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(RequestIDKey).(string)
	return reqID, ok
}
