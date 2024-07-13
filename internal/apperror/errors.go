package apperror

import (
	"errors"
	"net/http"
)

// ErrNoResults represents when no requested records exist in the database.
var ErrNoResults = errors.New("no results found in database")

// Status code of an Error, represents the general reason for the error.
type Status uint8

const (
	// StatusBadInput states an issue with the request.
	StatusBadInput Status = iota
	// StatusNotFound states that the requested resource could not be found.
	StatusNotFound
	// StatusInternal states an internal server issue.
	StatusInternal
	// StatusUnauthorized states a request was unauthorised.
	StatusUnauthorized
)

// Error represents a muzz application error, the status can be pivoted on to convert to generic interface errors,
// such as HTTP and gRPC.
type Error struct {
	status Status
	error  error
}

// ToHTTP takes a *Error and converts it into a *HTTPResponse.
func (e *Error) ToHTTP() *HTTPResponse {
	var status int
	switch e.Status() {
	case StatusBadInput:
		status = http.StatusBadRequest
	case StatusNotFound:
		status = http.StatusNotFound
	case StatusUnauthorized:
		status = http.StatusUnauthorized
	default:
		status = http.StatusInternalServerError
	}

	return &HTTPResponse{
		Status: status,
		Error:  e.Error(),
	}
}

// Status code of the error.
func (e *Error) Status() Status { return e.status }

// Error implements the error interface.
func (e *Error) Error() string {
	if e.error == nil {
		return ""
	}
	return e.error.Error()
}

// BadInput decorates the given error with a StatusBadInput.
func BadInput(err error) *Error {
	return &Error{
		status: StatusBadInput,
		error:  err,
	}
}

// NotFound decorates the given error with a StatusNotFound.
func NotFound(err error) *Error {
	return &Error{
		status: StatusNotFound,
		error:  err,
	}
}

// Internal decorates the given error with a StatusInternal.
func Internal(err error) *Error {
	return &Error{
		status: StatusInternal,
		error:  err,
	}
}

// Unauthorised decorates the given error with a StatusUnauthorized.
func Unauthorised(err error) *Error {
	return &Error{
		status: StatusUnauthorized,
		error:  err,
	}
}

// IncorrectCredentials Error for when a users authentication credentials are incorrect.
func IncorrectCredentials() *Error {
	return &Error{
		status: StatusUnauthorized,
		error:  errors.New("incorrect credentials"),
	}
}

// NewErr build a new *Error.
func NewErr(status Status, err error) *Error {
	return &Error{
		status: status,
		error:  err,
	}
}
