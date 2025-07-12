package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gitslim/go-ragger/internal/agent"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
	"github.com/gitslim/go-ragger/internal/web/auth"
	"github.com/gitslim/go-ragger/internal/web/documents"
	"github.com/gitslim/go-ragger/internal/web/home"
	"github.com/gitslim/go-ragger/internal/web/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"go.uber.org/fx"
)

func RegisterHTTPServerHooks(lc fx.Lifecycle, logger *slog.Logger, config *config.ServerConfig, q *sqlc.Queries, retrieverFactory milvus.MilvusRetrieverFactory, agentFactory agent.RagAgentFactory) {
	var srv *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			sessionStore := sessions.NewCookieStore([]byte(util.SessionKey))
			sessionStore.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteDefaultMode,
			}

			router := chi.NewRouter()
			router.Use(
				middleware.Logger,
				middleware.Recoverer,
				currentUserMiddleware(sessionStore, q),
				requestIDMiddleware(),
			)

			setupStaticRoute(router)
			home.SetupRoutes(router, logger, config, q, retrieverFactory, agentFactory)
			auth.SetupAuthRoutes(router, logger, q, sessionStore)
			documents.SetupRoutes(router, logger, q)
			upload.SetupFileUpload(router, logger, q)

			srv = &http.Server{
				Addr:    config.ServerAddress,
				Handler: router,
			}

			go func() {
				logger.Debug("starting web server", "config", config)
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
