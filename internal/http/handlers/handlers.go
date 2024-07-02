package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/app"
	"github.com/nickbadlose/muzz/internal/pkg/logger"
)

const (
	renderingErrorMessage = "rendering error response"
)

// TODO move handlers to handlers sub package and service to service subpackage. Both can uses types from here, or do
//  a domain package for types

type Handlers interface {
	CreateUser(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
}

type handlers struct {
	service app.Service
}

func NewHandlers(s app.Service) Handlers { return &handlers{s} }

// CreateUserRequest holds the information required to create a user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// UserView represents the user details to present to the client.
type UserView struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// UserResponse object to send to the client.
type UserResponse struct {
	Result *UserView `json:"result"`
}

// Render implements the render.Render interface.
func (u *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusCreated)
	return nil
}

// CreateUser forwards a create user request to the application.
func (h *handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := new(CreateUserRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding request", err)
		err = render.Render(w, r, errBadRequest(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	user, aErr := h.service.CreateUser(r.Context(), &app.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender,
		Age:      req.Age,
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, convertErr(aErr)),
		)
		return
	}

	err = render.Render(w, r, &UserResponse{
		Result: &UserView{
			ID:       user.ID,
			Email:    user.Email,
			Password: user.Password,
			Name:     user.Name,
			Gender:   user.Gender.String(),
			Age:      user.Age,
		},
	})
	if err != nil {
		logger.Error(r.Context(), "rendering response", err)
	}
}

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
func (l *LoginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Login forwards a log in request to the internal service.
func (h *handlers) Login(w http.ResponseWriter, r *http.Request) {
	req := new(app.LoginRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding request", err)
		err = render.Render(w, r, errBadRequest(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	token, aErr := h.service.Login(r.Context(), req)
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, convertErr(aErr)),
		)
		return
	}

	err = render.Render(w, r, &LoginResponse{Token: token})
	if err != nil {
		logger.Error(r.Context(), "rendering response", err)
	}
}
