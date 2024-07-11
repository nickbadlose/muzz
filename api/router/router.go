package router

import (
	"github.com/nickbadlose/muzz/internal/tracer"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func New(h *handlers.Handlers, cfg *config.Config, au *auth.Authorizer, tp trace.TracerProvider) http.Handler {
	r := chi.NewRouter()

	// TODO
	//  - context middleware
	//  - 405/404 middleware
	//  - add request id / trace id
	//  - use timestamps for migrations

	r.Use(
		middleware.AllowContentType("application/json"),
		middleware.RealIP,
		middleware.Logger, // TODO custom logger?
		middleware.Recoverer,
		render.SetContentType(render.ContentTypeJSON),
		tracer.HTTPResponseHeaders,
	)

	if cfg.DebugEnabled() {
		r.Use(tracer.HTTPDebugMiddleware(tp))
	}

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
		r.Use(auth.NewHTTPMiddleware(au))
		r.Get("/discover", h.Discover)
		r.Post("/swipe", h.Swipe)
	})

	return otelhttp.NewHandler(r, "router")
}