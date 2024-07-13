package muzz

import (
	"errors"
)

// CreateSwipeInput to create a swipe record.
type CreateSwipeInput struct {
	// UserID is the id of the user record performing the swipe action.
	UserID int
	// SwipedUserID is the id of the user record that the swipe action was performed against.
	SwipedUserID int
	// Preference is whether the user would prefer to match with the swiped user.
	Preference bool
}

// Validate the CreateSwipeInput fields.
func (in *CreateSwipeInput) Validate() error {
	if in.UserID == 0 {
		return errors.New("user id is a required field")
	}
	if in.SwipedUserID == 0 {
		return errors.New("swiped user id is a required field")
	}
	if in.UserID == in.SwipedUserID {
		return errors.New("user id and swiped user id cannot be the same value")
	}
	return nil
}
