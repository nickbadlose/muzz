package postgres

import (
	"context"

	"github.com/nickbadlose/muzz"
	"github.com/upper/db/v4"
)

const (
	swipeTable = "swipe"
)

type swipeEntity struct {
	id, userID, swipedUserID int
	preference               bool
}

func createSwipeWithTx(ctx context.Context, s db.SQL, in *muzz.CreateSwipeInput) (*swipeEntity, error) {
	columns := []string{"id", "user_id", "swiped_user_id", "preference"}
	row, err := s.InsertInto(swipeTable).
		Columns(columns[1:]...).
		Values(in.UserID, in.SwipedUserID, in.Preference).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(swipeEntity)
	err = row.Scan(&entity.id, &entity.userID, &entity.swipedUserID, &entity.preference)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func getSwipeWithTx(ctx context.Context, s db.SQL, userID, swipedUserID int) (*swipeEntity, error) {
	columns := []interface{}{"id", "user_id", "swiped_user_id", "preference"}
	row, err := s.Select(columns...).
		From(swipeTable).
		Where("user_id = ?", userID).
		And("swiped_user_id = ?", swipedUserID).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(swipeEntity)
	err = row.Scan(&entity.id, &entity.userID, &entity.swipedUserID, &entity.preference)
	if err != nil {
		return nil, err
	}

	return entity, nil
}
