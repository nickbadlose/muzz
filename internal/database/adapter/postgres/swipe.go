package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/upper/db/v4"
)

const (
	swipeTable = "swipe"
)

// SwipeAdapter adapts a *database.Database to the service.MatchRepository interface.
type SwipeAdapter struct {
	database *database.Database
}

// NewSwipeAdapter builds a new *SwipeAdapter.
func NewSwipeAdapter(d *database.Database) (*SwipeAdapter, error) {
	if d == nil {
		return nil, errors.New("database cannot be nil")
	}
	return &SwipeAdapter{database: d}, nil
}

// swipeEntity represents a row in the swipe table.
type swipeEntity struct {
	id, userID, swipedUserID int
	preference               bool
}

// CreateSwipe adds a swipe record to the swipe table and if appropriate, adds corresponding match records
// to the match table too.
//
// This method runs as a transaction. If both the user record performing the action and the swiped user record have a
// preference of true, then two match records are created, one for each user.
func (ma *SwipeAdapter) CreateSwipe(ctx context.Context, in *muzz.CreateSwipeInput) (*muzz.Match, error) {
	match := &muzz.Match{}
	err := ma.database.TxContext(ctx, func(tx db.Session) error {
		swipeInserted, tErr := createSwipeWithSession(ctx, tx.SQL(), in)
		if tErr != nil {
			logger.Error(ctx, "creating swipe", tErr)
			return tErr
		}

		if !swipeInserted.preference {
			return nil
		}

		swipeRetrieved, tErr := getSwipeWithSession(ctx, tx.SQL(), swipeInserted.swipedUserID, swipeInserted.userID)
		if tErr != nil {
			if errors.Is(tErr, sql.ErrNoRows) {
				return nil
			}

			logger.Error(ctx, "getting swipe of preferred user", tErr)
			return tErr
		}

		if !swipeRetrieved.preference {
			return nil
		}

		cmi := []*muzz.CreateMatchInput{
			{
				UserID:        swipeInserted.userID,
				MatchedUserID: swipeInserted.swipedUserID,
			},
			{
				UserID:        swipeRetrieved.userID,
				MatchedUserID: swipeRetrieved.swipedUserID,
			},
		}
		matches, tErr := createMatchWithSession(ctx, tx.SQL(), cmi)
		if tErr != nil {
			logger.Error(ctx, "creating first match record", tErr)
			return tErr
		}

		match = &muzz.Match{
			ID:            matches[0].ID,
			MatchedUserID: matches[0].MatchedUserID,
			Matched:       true,
		}

		return nil
	}, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	return match, nil
}

func createSwipeWithSession(ctx context.Context, s db.SQL, in *muzz.CreateSwipeInput) (*swipeEntity, error) {
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

func getSwipeWithSession(ctx context.Context, s db.SQL, userID, swipedUserID int) (*swipeEntity, error) {
	columns := []any{"id", "user_id", "swiped_user_id", "preference"}
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
