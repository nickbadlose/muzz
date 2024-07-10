package apperror

import (
	"errors"
	"net/http"
)

// NoResults represents when no requested records exist in the database.
var NoResults = errors.New("no results found in database")

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

func (e *Error) Status() Status { return e.status }
func (e *Error) Error() string {
	if e.error == nil {
		return ""
	}
	return e.error.Error()
}

func BadInput(err error) *Error {
	return &Error{
		status: StatusBadInput,
		error:  err,
	}
}

func NotFound(err error) *Error {
	return &Error{
		status: StatusNotFound,
		error:  err,
	}
}

func Internal(err error) *Error {
	return &Error{
		status: StatusInternal,
		error:  err,
	}
}

func Unauthorized(err error) *Error {
	return &Error{
		status: StatusUnauthorized,
		error:  err,
	}
}

func IncorrectCredentials() *Error {
	return &Error{
		status: StatusUnauthorized,
		error:  errors.New("incorrect credentials"),
	}
}

func NewErr(status Status, err error) *Error {
	return &Error{
		status: status,
		error:  err,
	}
}
