package app

import (
	"errors"
)

const (
	// the minimum age a user can be.
	minimumAge = 18
)

// User information.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// UserResponse object to send to the client.
type UserResponse struct {
	Result *User `json:"result"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// TODO
//  - Have a password regex to validate
// 	- Validate email, see if there is a lib for it
//  - Document how we would break app into separate sections as it grows, user section, with user subrouter and handlers, then eventually it's own microservice

func (req *CreateUserRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is a required field")
	}
	// TODO validate valid email
	if req.Password == "" {
		return errors.New("password is a required field")
	}
	// TODO validate against password constraints
	if req.Name == "" {
		return errors.New("name is a required field")
	}
	if req.Gender == "" {
		return errors.New("gender is a required field")
	}
	// TODO validate against accepted genders
	if req.Age < minimumAge {
		return errors.New("the minimum age is 18")
	}

	return nil
}
