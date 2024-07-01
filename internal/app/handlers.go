package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/pkg/logger"
)

const (
	renderingErrorMessage = "rendering error response"
)

// TODO move handlers to handlers sub package and service to service subpackage. Both can uses types from here, or do
//  a domain package for types

type Handlers interface {
	CreateUser(http.ResponseWriter, *http.Request)
}

type handlers struct {
	service Service
}

func NewHandlers(s Service) Handlers { return &handlers{s} }

func (h *handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := new(CreateUserRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding request", err)
		err = render.Render(w, r, errBadRequest(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		err = render.Render(w, r, errInternal(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	//render.Status(r, http.StatusCreated)
	//render.JSON(w, r, &UserResponse{Result: user})
	err = render.Render(w, r, &UserResponse{Result: user})
	if err != nil {
		logger.Error(r.Context(), "rendering response", err)
	}
}
