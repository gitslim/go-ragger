package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/gitslim/go-ragger/internal/web/auth"
	"github.com/gitslim/go-ragger/internal/web/documents"
	"github.com/gitslim/go-ragger/internal/web/home"
	"github.com/gitslim/go-ragger/internal/web/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"go.uber.org/fx"
)

func RegisterHTTPServerHooks(lc fx.Lifecycle, log *slog.Logger, cfg *config.ServerConfig, db *sqlc.Queries) {
	var srv *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			sessionStore := sessions.NewCookieStore([]byte(util.SessionKey))
			sessionStore.MaxAge(int(24 * time.Hour / time.Second))

			router := chi.NewRouter()
			router.Use(
				middleware.Logger,
				middleware.Recoverer,
				func(next http.Handler) http.Handler {
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
				},
			)

			setupStaticRoute(router)
			home.SetupRoutes(router)
			auth.SetupAuthRoutes(router, log, db, sessionStore)
			documents.SetupRoutes(router, log, db)
			upload.SetupFileUpload(router, log, db)

			srv = &http.Server{
				Addr:    cfg.ServerAddress,
				Handler: router,
			}

			go func() {
				srv.ListenAndServe()
			}()
			return nil
		},

		OnStop: func(ctx context.Context) error {
			srv.Shutdown(ctx)
			return nil
		},
	})

}

func setupStaticRoute(r chi.Router) {
	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
}
