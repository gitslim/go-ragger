package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/web/home"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"go.uber.org/fx"
)

const sessionKey = "ragger"

func RegisterHTTPServerHooks(lc fx.Lifecycle, cfg *config.ServerConfig) {
	var srv *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			sessionStore := sessions.NewCookieStore([]byte(sessionKey))
			sessionStore.MaxAge(int(24 * time.Hour / time.Second))

			router := chi.NewRouter()
			router.Use(
				middleware.Logger,
				middleware.Recoverer,
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						session, err := sessionStore.Get(r, sessionKey)
						if err != nil {
							http.Error(w, "failed to get session", http.StatusInternalServerError)
							return
						}

						// User from session
						userID, ok := session.Values["userID"].(int64)
						if !ok {
							next.ServeHTTP(w, r)
							return
						}
						slog.Debug("Request user", "userID", userID)

						next.ServeHTTP(w, r.WithContext(r.Context()))
					})
				},
			)

			setupStaticRoute(router)
			home.SetupRoutes(router)

			srv = &http.Server{
				Addr:    ":8080",
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
