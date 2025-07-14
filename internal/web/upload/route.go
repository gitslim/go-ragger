package upload

import (
	"encoding/base64"
	"log/slog"

	"fmt"
	"io"
	"net/http"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func SetupFileUpload(rtr chi.Router, log *slog.Logger, db *sqlc.Queries) error {
	rtr.Get("/upload", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)

		PageFileUpload(r, u).Render(ctx, w)
	})

	rtr.Get("/upload/data", func(w http.ResponseWriter, r *http.Request) {
		signals := &FileUploadSignals{
			FilesBase64: [][]string{},
			FileMimes:   [][]string{},
			FileNames:   [][]string{},
		}
		sse := datastar.NewSSE(w, r)
		sse.MergeFragmentTempl(FileUploadView(signals))
	})

	rtr.Post("/upload/upload", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)
		if u == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		maxBytesSize := 50 * 1024 * 1024
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytesSize))
		data, err := io.ReadAll(r.Body)
		if err != nil {
			if len(data) >= maxBytesSize {
				http.Error(w, "upload data is too large", http.StatusRequestEntityTooLarge)
				return
			}

			datastar.NewSSE(w, r).ConsoleError(fmt.Errorf("error reading body: %w", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		signals := &FileUploadSignals{}
		if err := json.Unmarshal(data, signals); err != nil {
			sse.ConsoleError(fmt.Errorf("error unmarshalling json: %w", err))
			return
		}

		for i, b64 := range signals.FilesBase64[0] {
			mimeType := signals.FileMimes[0][i]
			fileName := signals.FileNames[0][i]
			data, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				log.Error("decode file data error", "fileName", fileName, "error", err)
				continue
			}
			log.Debug("file decoded", "fileName", fileName)

			args := sqlc.CreateDocumentParams{
				UserID:   u.ID,
				FileName: fileName,
				MimeType: mimeType,
				FileData: data,
				FileSize: int64(len(data)),
				FileHash: util.CalculateBytesHash(data),
			}
			_, err = db.CreateDocument(ctx, args)
			if err != nil {
				log.Error("failed to create document", "error", err)
			}

		}

		sse.MergeFragmentTempl(FileUploadResults(signals))
	})

	return nil
}
