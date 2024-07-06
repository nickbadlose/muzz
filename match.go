package muzz

import "errors"

type Match struct {
	ID, MatchedUserID int
	Matched           bool
}

type CreateMatchInput struct {
	UserID, MatchedUserID int
}

func (in *CreateMatchInput) Validate() error {
	if in.UserID == 0 {
		return errors.New("user id is a required field")
	}
	if in.MatchedUserID == 0 {
		return errors.New("matched user id is a required field")
	}
	if in.UserID == in.MatchedUserID {
		return errors.New("user id and matched user id cannot be the same value")
	}

	return nil
}
