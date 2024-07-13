package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/internal/auth"
	"net/http"
)

func addPublicRoutes(r *chi.Mux, h *handlers.Handlers) {
	r.Group(func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, map[string]string{"status": "ok"})
		})
		r.Post("/user/create", h.CreateUser)
		r.Post("/login", h.Login)
	})
}

// addPrivate routes applies authentication middleware to any routes in this group.
func addPrivateRoutes(r *chi.Mux, h *handlers.Handlers, au *auth.Authoriser) {
	r.Group(func(r chi.Router) {
		r.Use(auth.NewHTTPMiddleware(au))

		r.Get("/discover", h.Discover)
		r.Post("/swipe", h.Swipe)
	})
}
