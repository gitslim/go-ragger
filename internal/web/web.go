package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/gitslim/go-ragger/internal/agent"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/gitslim/go-ragger/internal/web/auth"
	"github.com/gitslim/go-ragger/internal/web/documents"
	"github.com/gitslim/go-ragger/internal/web/home"
	"github.com/gitslim/go-ragger/internal/web/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"go.uber.org/fx"
)

func RegisterHTTPServerHooks(lc fx.Lifecycle, logger *slog.Logger, cfg *config.ServerConfig, db *sqlc.Queries, retriever retriever.Retriever, agentFactory agent.RagAgentFactory) {
	var srv *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			sessionStore := sessions.NewCookieStore([]byte(util.SessionKey))
			sessionStore.MaxAge(int(24 * time.Hour / time.Second))

			router := chi.NewRouter()
			router.Use(
				middleware.Logger,
				middleware.Recoverer,
				currentUserMiddleware(sessionStore, db),
				requestIDMiddleware(),
			)

			setupStaticRoute(router)
			home.SetupRoutes(router, logger, retriever, agentFactory)
			auth.SetupAuthRoutes(router, logger, db, sessionStore)
			documents.SetupRoutes(router, logger, db)
			upload.SetupFileUpload(router, logger, db)

			srv = &http.Server{
				Addr:    cfg.ServerAddress,
				Handler: router,
			}

			go func() {
				logger.Debug("starting web server", "config", cfg)
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
