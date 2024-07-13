package muzz

import "errors"

// Match contains all the stored details of a match record.
type Match struct {
	// ID is the unique identifier of the match record,
	// empty if Matched is false as no record was created.
	ID int
	// MatchedUserID is the id of the user record that was matched.
	MatchedUserID int
	// Matched is whether there was a match.
	Matched bool
}

// CreateMatchInput to create a match record.
type CreateMatchInput struct {
	// UserID is the id of the user record.
	UserID int
	// MatchedUserID is the id of the record that the UserID record has matched with.
	MatchedUserID int
}

// Validate the CreateMatchInput fields.
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
