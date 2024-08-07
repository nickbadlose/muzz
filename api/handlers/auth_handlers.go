package handlers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

// LoginRequest holds the information required to log in.
type LoginRequest struct {
	// Email of the user to authenticate with.
	Email string `json:"email"`
	// Password of the user to authenticate with.
	Password string `json:"password"`
}

// LoginResponse object to send to the client.
type LoginResponse struct {
	// Token is the valid JWT for the client to use for further request authorization.
	Token string `json:"token"`
}

// Render implements the render.Render interface.
func (*LoginResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Login attempts to authenticate the provided user credentials and login to the application. The users location
// is retrieved via their IP address and our user record is updated on successful authentication.
//
// On success a valid JWT is returned for authorization of future requests.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest[*LoginRequest](w, r)
	if err != nil {
		logger.Error(r.Context(), "decoding login request", err)
		return
	}

	location, err := h.location.ByIP(r.Context(), r.RemoteAddr)
	if err != nil {
		logger.MaybeError(r.Context(), renderingErrorMessage, render.Render(w, r, apperror.InternalServerHTTP(err)))
		return
	}

	token, aErr := h.authService.Login(r.Context(), &muzz.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		Location: location,
	})
	if aErr != nil {
		logger.MaybeError(r.Context(), renderingErrorMessage, render.Render(w, r, aErr.ToHTTP()))
		return
	}

	err = encodeResponse(w, r, h.config.DebugEnabled(), &LoginResponse{Token: token})
	if err != nil {
		logger.Error(r.Context(), "rendering login response", err)
	}
}
