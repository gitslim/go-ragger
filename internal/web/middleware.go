package web

import (
	"context"
	"net/http"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

// currentUserMiddleware fetches current user from session and adds it to the context
func currentUserMiddleware(sessionStore *sessions.CookieStore, db *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionStore.Get(r, util.SessionKey)
			if err != nil {
				http.Error(w, "failed to get session", http.StatusInternalServerError)
				return
			}

			userID, ok := session.Values[util.UserIDKey]
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			user, err := db.GetUserById(r.Context(), userID.(uuid.UUID))
			if err != nil {
				// Сбрасываем userID если не найден в бд
				delete(session.Values, util.UserIDKey)
				session.Save(r, w)
				http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
				return

			}
			ctx := util.ContextWithUser(r.Context(), &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}

// requestIDMiddleware fetches the request ID from the request header or creates new and adds it to the context
func requestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")

			if reqID == "" {
				reqID = uuid.New().String()
			}

			ctx := context.WithValue(r.Context(), util.RequestIDKey, reqID)
			w.Header().Set("X-Request-ID", reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
