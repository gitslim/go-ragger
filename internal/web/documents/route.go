package documents

import (
	"log/slog"
	"net/http"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func SetupRoutes(rtr chi.Router, logger *slog.Logger, db *sqlc.Queries) error {
	rtr.Get("/documents", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)
		if u == nil {
			// http.Error(w, "Forbidden", http.StatusForbidden)
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		signals := &ListSignals{
			Limit:  1000,
			Offset: 0}

		PageDocumentList(r, u, signals).Render(ctx, w)

	})

	rtr.Get("/documents/load", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)
		if u == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		signals := &ListSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if signals.Limit < 1 {
			signals.Limit = 10
		} else if signals.Limit > 100 {
			signals.Limit = 100
		}
		if signals.Offset < 0 {
			signals.Offset = 0
		}

		sse := datastar.NewSSE(w, r)

		args := sqlc.ListDocumentsParams{
			UserID:       u.ID,
			SearchQuery:  pgtype.Text{String: "", Valid: true},
			MimeFilter:   "",
			ResultLimit:  1000,
			ResultOffset: 0,
			StatusFilter: "",
		}

		docs, err := db.ListDocuments(ctx, args)
		if err != nil {
			logger.Error("failed to list documents", "error", err)
		}

		sse.MergeFragmentTempl(DocumentRows(docs))

	})

	return nil
}
