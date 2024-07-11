package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

const (
	authorizationHeader   = "Authorization"
	renderingErrorMessage = "rendering error response"
)

// NewHTTPMiddleware returns a middleware that authenticates the request JWT and sets the user ID on context.
func NewHTTPMiddleware(authorizer *Authorizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get(authorizationHeader)
			splitBearer := strings.Split(bearerToken, " ")
			if len(splitBearer) != 2 {
				logger.MaybeError(
					r.Context(),
					renderingErrorMessage,
					render.Render(w, r, apperror.UnauthorisedHTTP(errors.New("no authorisation provided"))),
				)
				return
			}
			jwt := splitBearer[1]

			userID, err := authorizer.Authorize(jwt)
			if err != nil {
				logger.MaybeError(
					r.Context(),
					renderingErrorMessage,
					render.Render(w, r, err.ToHTTP()),
				)
				return
			}

			ctx := authorizer.UserOnContext(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
