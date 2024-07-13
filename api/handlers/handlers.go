package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/paulmach/orb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	renderingErrorMessage = "rendering error response"
)

// Config is the interface to retrieve configuration secrets from the environment.
type Config interface {
	// DebugEnabled retrieves the debug state of the application.
	DebugEnabled() bool
}

// Authorizer is the interface to retrieve an authenticated users ID from.
type Authorizer interface {
	// UserFromContext gets the authenticated user from context.
	UserFromContext(ctx context.Context) (userID int, err error)
}

// Locationer is the interface to retrieve location information.
type Locationer interface {
	// ByIP retrieves a user's IP address.
	ByIP(ctx context.Context, sourceIP string) (orb.Point, error)
}

// Handlers holds the HTTP handlers for the valid server endpoints.
type Handlers struct {
	config       Config
	authorizer   Authorizer
	location     Locationer
	authService  *service.AuthService
	userService  *service.UserService
	matchService *service.SwipeService
}

// New builds a new *Handlers.
func New(
	cfg *config.Config,
	auth Authorizer,
	l Locationer,
	as *service.AuthService,
	us *service.UserService,
	ms *service.SwipeService,
) (*Handlers, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if auth == nil {
		return nil, errors.New("authorizer cannot be nil")
	}
	if l == nil {
		return nil, errors.New("locationer cannot be nil")
	}
	if as == nil {
		return nil, errors.New("auth service cannot be nil")
	}
	if us == nil {
		return nil, errors.New("user service cannot be nil")
	}
	if ms == nil {
		return nil, errors.New("match service cannot be nil")
	}
	return &Handlers{
		config:       cfg,
		authorizer:   auth,
		location:     l,
		authService:  as,
		userService:  us,
		matchService: ms,
	}, nil
}

func encodeResponse[T render.Renderer](w http.ResponseWriter, r *http.Request, debug bool, v T) error {
	// If debugging, decorate traces with the response data.
	if debug {
		span := trace.SpanFromContext(r.Context())
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(v)
		if err != nil {
			logger.Error(r.Context(), "tracing response body", err)
			span.SetStatus(codes.Error, "encoding response body")
		}

		span.SetAttributes(attribute.String("http.response_body", buf.String()))
	}

	return render.Render(w, r, v)
}

// decodeRequest renders a response error on failure.
func decodeRequest[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		logger.MaybeError(r.Context(), renderingErrorMessage, render.Render(w, r, apperror.BadRequestHTTP(err)))
		return v, err
	}

	return v, nil
}
