package documents

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func isPreviewable(mimeType string) bool {
	previewableTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"text/plain",
	}
	for _, t := range previewableTypes {
		if strings.HasPrefix(mimeType, t) {
			return true
		}
	}
	return false
}

func renderDocumentPreview(doc sqlc.Document) templ.Component {
	switch {
	case strings.HasPrefix(doc.MimeType, "image/"):
		return templ.Raw(fmt.Sprintf(
			`<img src="/documents/preview/%s" class="img-fluid" alt="Preview">`,
			doc.ID,
		))
	case doc.MimeType == "application/pdf":
		return templ.Raw(fmt.Sprintf(
			`<iframe src="/documents/preview/%s" class="w-100" style="height: 600px; border: none;"></iframe>`,
			doc.ID,
		))
	default:
		return templ.Raw(`<p>Просмотр для этого типа файла не поддерживается</p>`)
	}
}

func formatTimestamp(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return "N/A"
	}
	return ts.Time.Format("02.01.2006 15:04")
}
