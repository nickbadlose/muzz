package postgres

import (
	"context"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/upper/db/v4"
)

const (
	matchTable = "match"
)

// matchEntity represents a row in the match table.
//
// upper db tags attached for batch inserting match rows.
type matchEntity struct {
	ID            int `db:"id"`
	UserID        int `db:"user_id"`
	MatchedUserID int `db:"matched_user_id"`
}

// createMatchWithSession batch inserts a list of matches into the match table using the provided session.
func createMatchWithSession(ctx context.Context, s db.SQL, in []*muzz.CreateMatchInput) ([]*matchEntity, error) {
	columns := []string{"id", "user_id", "matched_user_id"}

	bi := s.InsertInto(matchTable).
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
