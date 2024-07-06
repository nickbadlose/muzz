package muzz

import "errors"

// LoginInput is the accepted request format to log in.
type LoginInput struct {
	Email    string
	Password string
}

// Validate the LoginInput fields.
func (lr *LoginInput) Validate() error {
	if lr.Email == "" {
		return errors.New("email is a required field")
	}
	if lr.Password == "" {
		return errors.New("password is a required field")
	}

	return nil
}
