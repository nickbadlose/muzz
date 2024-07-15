package router

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	"github.com/nickbadlose/muzz/internal/tracer"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// New builds a new http.Handler.
func New(
	cfg *config.Config,
	h *handlers.Handlers,
	au *auth.Authoriser,
	tp trace.TracerProvider,
) (http.Handler, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if h == nil {
		return nil, errors.New("handlers cannot be nil")
	}
	if au == nil {
		return nil, errors.New("authoriser cannot be nil")
	}
	if tp == nil {
		return nil, errors.New("tracer provider cannot be nil")
	}

	r := chi.NewRouter()

	r.Use(
		middleware.AllowContentType("application/json"),
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		render.SetContentType(render.ContentTypeJSON),
		tracer.HTTPResponseHeaders,
	)

	if cfg.DebugEnabled() {
		r.Use(tracer.HTTPDebugMiddleware(tp))
	}

	addPublicRoutes(r, h)
	addPrivateRoutes(r, h, au)

	return otelhttp.NewHandler(r, "router"), nil
}
