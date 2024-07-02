package handlers

import (
	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/app"
	"net/http"
)

// ErrResponse represents the error format to respond with.
type ErrResponse struct {
	// Status is the http status code to respond with.
	Status int `json:"status"`
	// Error represents the error message to respond with.
	Error string `json:"error"`
}

// Render implements the render.Render interface.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Status)
	return nil
}

func errBadRequest(err error) *ErrResponse {
	return &ErrResponse{
		Status: http.StatusBadRequest,
		Error:  err.Error(),
	}
}

func convertErr(err app.Error) *ErrResponse {
	var status int
	switch err.Status() {
	case app.ErrorStatusBadRequest:
		status = http.StatusBadRequest
	case app.ErrorStatusNotFound:
		status = http.StatusNotFound
	default:
		status = http.StatusInternalServerError
	}

	return &ErrResponse{
		Status: status,
		Error:  err.Error(),
	}
}
