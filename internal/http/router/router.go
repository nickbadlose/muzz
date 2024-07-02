package router

import (
	"github.com/nickbadlose/muzz/internal/pkg/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/http/handlers"
	middlewareInternal "github.com/nickbadlose/muzz/internal/http/middleware"
)

func New(h handlers.Handlers, v auth.Validator) http.Handler {
	r := chi.NewRouter()

	// TODO
	//  - context middleware
	//  - 405/404 middleware
	//  - add request id / trace id

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
		r.Use(middlewareInternal.Authorization(v))
		//r.Post("/manage", CreateAsset)
	})

	return r
}
