package backend

import (
	"net/http"

	"github.com/go-chi/render"
)

func (backend *Backend) httpGetProviderList(w http.ResponseWriter, r *http.Request) {
	if len(backend.auth.Providers) > 0 {
		render.JSON(w, r, backend.auth.Providers)
		return
	}

	render.Status(r, http.StatusNotFound)
	render.JSON(w, r, JSON{"error": "authorization not configured"})
}
