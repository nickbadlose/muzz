package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/paulmach/orb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

const (
	renderingErrorMessage = "rendering error response"
)

// TODO move handlers to handlers sub package and service to service subpackage. Both can uses types from here, or do
//  a domain package for types
//  Have a check for nil items in constructors and log.FatalF if they are nil.

// TODO omitempty MatchID in SwipeResponse

// TODO restrict handlers that need authorizer? Pass in using handler adapter pattern.
//  No need for service interfaces in here. If the service needs editing,
//  so does this probably as it will be a business decision

type Config interface {
	DebugEnabled() bool
}

type Authorizer interface {
	// UserFromContext gets the authenticated user from context.
	UserFromContext(ctx context.Context) (userID int, err error)
}

type Locationer interface {
	ByIP(ctx context.Context, sourceIP string) (orb.Point, error)
}

type Handlers struct {
	config       Config
	authorizer   Authorizer
	location     Locationer
	authService  *service.AuthService
	userService  *service.UserService
	matchService *service.MatchService
}

func New(
	cfg *config.Config,
	auth Authorizer,
	l Locationer,
	as *service.AuthService,
	us *service.UserService,
	ms *service.MatchService,
) *Handlers {
	return &Handlers{cfg, auth, l, as, us, ms}
}

func (h *Handlers) decode() {}
func (h *Handlers) encode() {}

func encodeResponse[T render.Renderer](w http.ResponseWriter, r *http.Request, debug bool, v T) error {
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

// TODO force error and check

func decodeRequest[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return v, err
	}

	return v, nil
}
