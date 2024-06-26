package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger, // TODO custom logger?
		middleware.Recoverer,
	)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return r
}
