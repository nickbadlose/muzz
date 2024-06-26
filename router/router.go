package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"net/http"
)

func New() http.Handler {
	r := chi.NewRouter()

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

	return r
}
