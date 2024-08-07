package muzz

import (
	"errors"

	"github.com/paulmach/orb"
)

// LoginInput is the accepted request format to log in.
type LoginInput struct {
	// Email of the user record to authenticate.
	Email string
	// Password of the user record to authenticate.
	Password string
	// Location is the current location of the user to authenticate.
	Location orb.Point
}

// Validate the LoginInput fields.
func (in *LoginInput) Validate() error {
	if in.Email == "" {
		return errors.New("email is a required field")
	}
	if in.Password == "" {
		return errors.New("password is a required field")
	}

	err := validatePoint(in.Location)
	if err != nil {
		return err
	}

	return nil
}
