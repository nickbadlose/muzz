package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

// LoginRequest holds the information required to log in.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse object to send to the client.
type LoginResponse struct {
	// Token is the valid JWT for the client to use for further request authorization.
	Token string `json:"token"`
}

// Render implements the render.Render interface.
func (*LoginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Login logs the user into the application.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	req := new(LoginRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding login request", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	token, aErr := h.authService.Login(r.Context(), &muzz.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, aErr.ToHTTP()),
		)
		return
	}

	err = render.Render(w, r, &LoginResponse{Token: token})
	if err != nil {
		logger.Error(r.Context(), "rendering login response", err)
	}
}
