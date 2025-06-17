package home

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router) {

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		PageHome(r).Render(ctx, w)
	})
}
