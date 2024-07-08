package apperror

import (
	"net/http"

	"github.com/go-chi/render"
)

// HTTPResponse represents the error format to respond with.
type HTTPResponse struct {
	// Status is the http status code to respond with.
	Status int `json:"status"`
	// Error represents the error message to respond with.
	Error string `json:"error"`
}

// Render implements the render.Render interface.
func (e *HTTPResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Status)
	return nil
}

// UnauthorisedHTTP returns an error wrapped in a http.StatusUnauthorized.
func UnauthorisedHTTP(err error) *HTTPResponse {
	return &HTTPResponse{
		Status: http.StatusUnauthorized,
		Error:  err.Error(),
	}
}

// BadRequestHTTP returns an error wrapped in a http.StatusBadRequest.
func BadRequestHTTP(err error) *HTTPResponse {
	return &HTTPResponse{
		Status: http.StatusBadRequest,
		Error:  err.Error(),
	}
}

// InternalServerHTTP returns an error wrapped in a http.StatusBadRequest.
func InternalServerHTTP(err error) *HTTPResponse {
	return &HTTPResponse{
		Status: http.StatusInternalServerError,
		Error:  err.Error(),
	}
}
