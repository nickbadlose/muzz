package muzz

import "errors"

// LoginInput is the accepted request format to log in.
type LoginInput struct {
	Email    string
	Password string
}

// Validate the LoginInput fields.
func (in *LoginInput) Validate() error {
	if in.Email == "" {
		return errors.New("email is a required field")
	}
	if in.Password == "" {
		return errors.New("password is a required field")
	}

	return nil
}
