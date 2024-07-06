package muzz

import (
	"errors"
)

type Swipe struct {
	ID, UserID, SwipedUserID int
	Preference               bool
}

type CreateSwipeInput struct {
	UserID       int
	SwipedUserID int
	Preference   bool
}

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
