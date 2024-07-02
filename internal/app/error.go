package app

type ErrorStatus uint8

const (
	// ErrorStatusBadRequest states an issue with the request.
	ErrorStatusBadRequest ErrorStatus = iota
	// ErrorStatusNotFound states that the requested resource could not be found.
	ErrorStatusNotFound
	// ErrorStatusInternal states an internal server issue.
	ErrorStatusInternal
	// ErrorStatusUnauthorised states a request was unauthorised.
	ErrorStatusUnauthorised
)

type Error interface {
	Status() ErrorStatus
	Error() string
}

type applicationError struct {
	status ErrorStatus
	error  error
}

func (e *applicationError) Status() ErrorStatus { return e.status }
func (e *applicationError) Error() string {
	if e.error == nil {
		return ""
	}
	return e.error.Error()
}

func errBadRequest(err error) Error {
	return &applicationError{
		status: ErrorStatusBadRequest,
		error:  err,
	}
}

func errInternal(err error) Error {
	return &applicationError{
		status: ErrorStatusInternal,
		error:  err,
	}
}

func errUnauthorised(err error) Error {
	return &applicationError{
		status: ErrorStatusUnauthorised,
		error:  err,
	}
}

func newErr(status ErrorStatus, err error) Error {
	return &applicationError{
		status: status,
		error:  err,
	}
}
