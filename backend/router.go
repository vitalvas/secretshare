package backend

import (
	"net/http"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type JSON map[string]interface{}

func (backend *Backend) GetRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)

	authRoutes, avaRoutes := backend.auth.Service.Handlers()
	router.Mount("/auth", authRoutes)
	router.Mount("/avatar", avaRoutes)

	m := backend.auth.Service.Middleware()

	router.Group(func(r chi.Router) {
		if len(backend.auth.Providers) > 0 {
			r.Use(m.Auth)
		}

		r.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
		r.Post("/api/v1/message", backend.httpSaveMessage)
	})

	router.Get("/api/v1/message/{key}", backend.httpGetMessage)
	router.Get("/api/v1/message/{key}/{pin}", backend.httpGetMessage)

	router.Get("/api/v1/auth/providers", backend.httpGetProviderList)

	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "User-agent: *\nDisallow: /api/\nDisallow: /show/\n")
	})

	return router
}
