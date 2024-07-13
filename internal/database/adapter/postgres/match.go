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
	matchTable = "match"
)

// MatchAdapter adapts a *database.Database to the service.MatchRepository interface.
type MatchAdapter struct {
	database *database.Database
}

// NewMatchAdapter builds a new *MatchAdapter.
func NewMatchAdapter(d *database.Database) (*MatchAdapter, error) {
	if d == nil {
		return nil, errors.New("database cannot be nil")
	}
	return &MatchAdapter{database: d}, nil
}

// matchEntity represents a row in the match table.
//
// upper db tags attached for batch inserting match rows.
type matchEntity struct {
	ID            int `db:"id"`
	UserID        int `db:"user_id"`
	MatchedUserID int `db:"matched_user_id"`
}

// CreateSwipe adds a swipe record to the swipe table and if appropriate, adds corresponding match records
// to the match table too.
//
// This method runs as a transaction. If both the user record performing the action and the swiped user record have a
// preference of true, then two match records are created, one for each user.
func (ma *MatchAdapter) CreateSwipe(ctx context.Context, in *muzz.CreateSwipeInput) (*muzz.Match, error) {
	match := &muzz.Match{}
	err := ma.database.TxContext(ctx, func(tx db.Session) error {
		swipeInserted, tErr := createSwipeWithTx(ctx, tx.SQL(), in)
		if tErr != nil {
			logger.Error(ctx, "creating swipe", tErr)
			return tErr
		}

		if !swipeInserted.preference {
			return nil
		}

		swipeRetrieved, tErr := getSwipeWithTx(ctx, tx.SQL(), swipeInserted.swipedUserID, swipeInserted.userID)
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
		matches, tErr := createMatchWithTx(ctx, tx.SQL(), cmi)
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

// createMatchWithTx batch inserts a list of matches into the match table.
func createMatchWithTx(ctx context.Context, tx db.SQL, in []*muzz.CreateMatchInput) ([]*matchEntity, error) {
	columns := []string{"id", "user_id", "matched_user_id"}

	bi := tx.InsertInto(matchTable).
		Columns(columns[1:]...).
		Returning(columns...).
		Batch(len(in))

	for _, cmi := range in {
		err := cmi.Validate()
		if err != nil {
			logger.Error(ctx, "validating create match input", err)
			return nil, err
		}

		bi.Values(
			cmi.UserID,
			cmi.MatchedUserID,
		)
	}

	bi.Done()

	out := make([]*matchEntity, 0, len(in))
	for {
		batch := make([]*matchEntity, 0, len(in))
		if err := bi.Err(); err != nil {
			logger.Error(ctx, "batch inserting matches", err)
			return nil, err
		}
		if bi.NextResult(&batch) {
			out = append(out, batch...)
			continue
		}
		break
	}

	return out, bi.Wait()
}
