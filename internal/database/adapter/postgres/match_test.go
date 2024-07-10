package postgres

import (
	"context"
	"errors"
	"testing"

	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nickbadlose/muzz"
	"github.com/stretchr/testify/require"
)

func newTestMatchAdapter(t *testing.T) (*MatchAdapter, sqlmock.Sqlmock) {
	dbase, mock := newTestDB(t)
	return NewMatchAdapter(dbase), mock
}

func TestMatchAdapter_CreateSwipe(t *testing.T) {
	t.Run("success: swiped false", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		rows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(1, 1, 2, false)

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, false).
			WillReturnRows(rows)

		mock.ExpectCommit()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   false,
			},
		)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, &muzz.Match{
			ID:            0,
			MatchedUserID: 0,
			Matched:       false,
		}, got)
	})

	t.Run("success: swiped true but not swiped back", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		insertSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(2, 1, 2, true)

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, true).
			WillReturnRows(insertSwipeRows)

		getSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(1, 2, 1, false)

		mock.ExpectQuery(`SELECT "id", "user_id", "swiped_user_id", "preference" 
FROM "swipe" WHERE \(user_id = \$1 AND swiped_user_id = \$2\)`).
			WithArgs(2, 1).
			WillReturnRows(getSwipeRows)

		mock.ExpectCommit()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   true,
			},
		)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, &muzz.Match{
			ID:            0,
			MatchedUserID: 0,
			Matched:       false,
		}, got)
	})

	t.Run("success: match", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		insertSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(2, 1, 2, true)

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, true).
			WillReturnRows(insertSwipeRows)

		getSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(1, 2, 1, true)

		mock.ExpectQuery(`SELECT "id", "user_id", "swiped_user_id", "preference" 
FROM "swipe" WHERE \(user_id = \$1 AND swiped_user_id = \$2\)`).
			WithArgs(2, 1).
			WillReturnRows(getSwipeRows)

		insertMatchRows := mock.NewRows([]string{"id", "user_id", "matched_user_id"}).
			AddRows(
				[]driver.Value{1, 1, 2},
				[]driver.Value{2, 2, 1},
			)

		mock.ExpectQuery(`INSERT INTO "match" \("user_id", "matched_user_id"\) VALUES \(\$1, \$2\), \(\$3, \$4\) 
                                                   RETURNING "id", "user_id", "matched_user_id"`).
			WithArgs(1, 2, 2, 1).
			WillReturnRows(insertMatchRows)

		mock.ExpectCommit()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   true,
			},
		)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, &muzz.Match{
			ID:            1,
			MatchedUserID: 2,
			Matched:       true,
		}, got)
	})

	t.Run("error inserting swipe", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, false).
			WillReturnError(errors.New("query error"))

		mock.ExpectRollback()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   false,
			},
		)

		require.Error(t, err)
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, "query error", err.Error())
	})

	t.Run("error getting swiped user", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		insertSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(2, 1, 2, true)

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, true).
			WillReturnRows(insertSwipeRows)

		mock.ExpectQuery(`SELECT "id", "user_id", "swiped_user_id", "preference" 
FROM "swipe" WHERE \(user_id = \$1 AND swiped_user_id = \$2\)`).
			WithArgs(2, 1).
			WillReturnError(errors.New("query error"))

		mock.ExpectRollback()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   true,
			},
		)

		require.Error(t, err)
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, "query error", err.Error())
	})

	t.Run("error inserting matches: batch wait", func(t *testing.T) {
		sut, mock := newTestMatchAdapter(t)

		mock.ExpectBegin()

		insertSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(2, 1, 2, true)

		mock.ExpectQuery(`INSERT INTO "swipe" \("user_id", "swiped_user_id", "preference"\) 
VALUES \(\$1, \$2, \$3\) RETURNING "id", "user_id", "swiped_user_id", "preference"`).
			WithArgs(1, 2, true).
			WillReturnRows(insertSwipeRows)

		getSwipeRows := mock.NewRows([]string{"id", "user_id", "swiped_user_id", "preference"}).
			AddRow(1, 2, 1, true)

		mock.ExpectQuery(`SELECT "id", "user_id", "swiped_user_id", "preference" 
FROM "swipe" WHERE \(user_id = \$1 AND swiped_user_id = \$2\)`).
			WithArgs(2, 1).
			WillReturnRows(getSwipeRows)

		mock.ExpectQuery(`INSERT INTO "match" \("user_id", "matched_user_id"\) VALUES \(\$1, \$2\), \(\$3, \$4\) 
                                                   RETURNING "id", "user_id", "matched_user_id"`).
			WithArgs(1, 2, 2, 1).
			WillReturnError(errors.New("query error"))

		mock.ExpectRollback()

		got, err := sut.CreateSwipe(
			context.Background(),
			&muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   true,
			},
		)

		require.Error(t, err)
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, "query error", err.Error())
	})
}
