package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/internal/auth"
	httpMiddleware "github.com/nickbadlose/muzz/internal/middleware/http"
)

func New(h *handlers.Handlers, v *auth.Authorizer) http.Handler {
	r := chi.NewRouter()

	// TODO
	//  - context middleware
	//  - 405/404 middleware
	//  - add request id / trace id
	//  - use timestamps for migrations

	r.Use(
		middleware.AllowContentType("application/json"),
		middleware.Logger, // TODO custom logger?
		middleware.Recoverer,
	)
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
	)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, map[string]string{"status": "ok"})
		})
		r.Post("/user/create", h.CreateUser)
		r.Post("/login", h.Login)
	})

	// Private routes
	r.Group(func(r chi.Router) {
		r.Use(httpMiddleware.NewAuthMiddleware(v))
		r.Get("/discover", h.Discover)
		r.Post("/swipe", h.Swipe)
	})

	return r
}
