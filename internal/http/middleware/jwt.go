package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nickbadlose/muzz/internal/http/handlers"
	"github.com/nickbadlose/muzz/internal/pkg/auth"
	"github.com/nickbadlose/muzz/internal/pkg/logger"
)

const (
	authorizationHeader   = "Authorization"
	renderingErrorMessage = "rendering error response"
)

// Authorization returns a middleware that authenticates the request JWT and sets the user ID on context.
func Authorization(validator auth.Validator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get(authorizationHeader)
			reqToken := strings.Split(bearerToken, " ")[1]
			token, claims, err := validator.ParseJWT(reqToken)
			if err != nil {
				if errors.Is(err, jwt.ErrSignatureInvalid) {
					logger.MaybeError(
						r.Context(),
						renderingErrorMessage,
						render.Render(w, r, handlers.ErrUnauthorised(err)),
					)
					return
				}

				logger.MaybeError(
					r.Context(),
					renderingErrorMessage,
					render.Render(w, r, handlers.ErrBadRequest(err)),
				)
				return
			}
			if !token.Valid {
				logger.MaybeError(
					r.Context(),
					renderingErrorMessage,
					render.Render(w, r, handlers.ErrUnauthorised(errors.New("invalid jwt"))),
				)
				return
			}

			err = validator.ValidateClaims(claims)
			if err != nil {
				logger.MaybeError(
					r.Context(),
					renderingErrorMessage,
					render.Render(w, r, handlers.ErrUnauthorised(err)),
				)
				return
			}

			ctx := context.WithValue(r.Context(), auth.UserKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
