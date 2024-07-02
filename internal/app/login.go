package app

import (
	"errors"
)

// LoginRequest is the accepted request format to log in.
type LoginRequest struct {
	Email    string
	Password string
}

// Validate the LoginRequest fields.
func (req *LoginRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is a required field")
	}
	if req.Password == "" {
		return errors.New("password is a required field")
	}

	return nil
}
