package home

import (
	"net/http"

	"github.com/gitslim/go-ragger/internal/util"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router) {

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)

		PageHome(r, u).Render(ctx, w)
	})
}
