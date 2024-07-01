package app

import (
	"github.com/go-chi/render"
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

func errInternal(err error) *ErrResponse {
	return &ErrResponse{
		Status: http.StatusInternalServerError,
		Error:  err.Error(),
	}
}

func errBadRequest(err error) *ErrResponse {
	return &ErrResponse{
		Status: http.StatusBadRequest,
		Error:  err.Error(),
	}
}

func newErr(status int, err error) *ErrResponse {
	return &ErrResponse{
		Status: status,
		Error:  err.Error(),
	}
}
