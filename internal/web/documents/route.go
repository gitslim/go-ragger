package documents

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func SetupRoutes(rtr chi.Router, logger *slog.Logger, db *sqlc.Queries) error {
	rtr.Get("/documents", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)
		if user == nil {
			// http.Error(w, "forbidden", http.StatusForbidden)
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		signals := &ListSignals{
			Limit:  1000,
			Offset: 0}

		PageDocumentList(r, user, signals).Render(ctx, w)

	})

	rtr.Get("/documents/load", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("@get documents/load")
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)
		if u == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		signals := &ListSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			logger.Error("error unmarshalling signals", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
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

	rtr.Get("/documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)
		if user == nil {
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		docID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		doc, err := db.GetUserDocumentById(ctx, sqlc.GetUserDocumentByIdParams{ID: docID, UserID: user.ID})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Document not found", http.StatusNotFound)
			} else {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		if doc.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		PageDocumentView(r, user, doc).Render(ctx, w)
	})

	rtr.Get("/documents/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)
		if user == nil {
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		docID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		doc, err := db.GetUserDocumentById(ctx, sqlc.GetUserDocumentByIdParams{ID: docID, UserID: user.ID})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Document not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to get document", http.StatusInternalServerError)
			}
			return
		}

		if doc.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.FileName))
		w.Header().Set("Content-Type", doc.MimeType)
		w.Header().Set("Content-Length", strconv.FormatInt(doc.FileSize, 10))

		if _, err := w.Write(doc.FileData); err != nil {
			log.Printf("Failed to send file: %v", err)
		}
	})

	rtr.Get("/documents/preview/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)
		if user == nil {
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		docID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		doc, err := db.GetUserDocumentById(ctx, sqlc.GetUserDocumentByIdParams{ID: docID, UserID: user.ID})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Document not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to get document", http.StatusInternalServerError)
			}
			return
		}

		if doc.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !isPreviewable(doc.MimeType) {
			http.Error(w, "Preview not available for this file type", http.StatusBadRequest)
			return
		}

		if strings.HasPrefix(doc.MimeType, "image/") {
			w.Header().Set("Content-Type", doc.MimeType)
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Write(doc.FileData)
			return
		}

		if doc.MimeType == "application/pdf" {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", doc.FileName))
			w.Write(doc.FileData)
			return
		}

		if strings.HasPrefix(doc.MimeType, "text/") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write(doc.FileData)
			return
		}

		http.Error(w, "Unsupported file type for preview", http.StatusBadRequest)
	})

	return nil
}
