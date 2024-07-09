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

// TODO
//  what is cleanest way to handle match table? one row that holds both users or 2 rows, one for each user?
//  return errors from constructors or panic if params are nil?

// MatchAdapter adapts a *database.Database to the service.MatchRepository interface.
type MatchAdapter struct {
	database *database.Database
}

func NewMatchAdapter(d *database.Database) *MatchAdapter {
	return &MatchAdapter{database: d}
}

type matchEntity struct {
	id, userID, matchedUserID int
}

func (ma *MatchAdapter) CreateSwipe(ctx context.Context, in *muzz.CreateSwipeInput) (*muzz.Match, error) {
	match := &muzz.Match{}
	err := ma.database.TxContext(ctx, func(tx db.Session) error {
		swipe, tErr := createSwipeWithTx(ctx, tx.SQL(), in)
		if tErr != nil {
			logger.Error(ctx, "creating swipe", tErr)
			return tErr
		}

		if !swipe.preference {
			return nil
		}

		swipe, tErr = getSwipeWithTx(ctx, tx.SQL(), in.SwipedUserID, in.UserID)
		if tErr != nil {
			if errors.Is(tErr, sql.ErrNoRows) {
				return nil
			}

			logger.Error(ctx, "getting swipe of preferred user", tErr)
			return tErr
		}

		if !swipe.preference {
			return nil
		}

		// TODO
		//  batch insert matches if using 2 rows for one match
		//  have createMatchWithTx take in multiple cmis

		cmi := &muzz.CreateMatchInput{
			UserID:        in.UserID,
			MatchedUserID: in.SwipedUserID,
		}
		tErr = cmi.Validate()
		if tErr != nil {
			logger.Error(ctx, "validating create match input", tErr)
			return tErr
		}

		matchEnt, tErr := createMatchWithTx(ctx, tx.SQL(), cmi)
		if tErr != nil {
			logger.Error(ctx, "creating first match record", tErr)
			return tErr
		}

		cmi = &muzz.CreateMatchInput{
			UserID:        in.SwipedUserID,
			MatchedUserID: in.UserID,
		}
		tErr = cmi.Validate()
		if tErr != nil {
			logger.Error(ctx, "validating create match input", tErr)
			return tErr
		}

		_, tErr = createMatchWithTx(ctx, tx.SQL(), cmi)
		if tErr != nil {
			logger.Error(ctx, "creating second match record", tErr)
			return tErr
		}

		match = &muzz.Match{
			ID:            matchEnt.id,
			MatchedUserID: matchEnt.matchedUserID,
			Matched:       true,
		}

		return nil
		//	TODO set these ?
	}, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	return match, nil
}

func createMatchWithTx(ctx context.Context, w database.Writer, in *muzz.CreateMatchInput) (*matchEntity, error) {
	columns := []string{"id", "user_id", "matched_user_id"}
	row, err := w.InsertInto(matchTable).
		Columns(columns[1:]...).
		Values(in.UserID, in.MatchedUserID).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(matchEntity)
	err = row.Scan(&entity.id, &entity.userID, &entity.matchedUserID)
	if err != nil {
		return nil, err
	}

	return entity, nil
}
