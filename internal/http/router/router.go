package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/http/handlers"
)

func New(h handlers.Handlers) http.Handler {
	r := chi.NewRouter()

	// TODO
	//  - context middleware
	//  - 405/404 middleware

	r.Use(
		middleware.AllowContentType("application/json"),
		middleware.Logger, // TODO custom logger?
		middleware.Recoverer,
	)
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
	)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]string{"status": "ok"})
	})

	r.Post("/user/create", h.CreateUser)
	r.Post("/login", h.Login)

	return r
}
