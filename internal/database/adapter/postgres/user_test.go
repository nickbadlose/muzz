package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/database"
	mockdatabase "github.com/nickbadlose/muzz/internal/database/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func newTestDB(t *testing.T) (*database.Database, sqlmock.Sqlmock) {
	dbClient, mockSQL, err := mockdatabase.NewWrappedMock()

	db, err := database.New(
		context.Background(),
		&database.Config{},
		func(_ context.Context, _ *database.Config) (database.Client, error) {
			return dbClient, err
		},
	)

	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		mockSQL.ExpectClose()
		require.NoError(t, db.Close())
	})

	return db, mockSQL
}

func newTestUserAdapter(t *testing.T) (*UserAdapter, sqlmock.Sqlmock) {
	db, mock := newTestDB(t)
	return NewUserAdapter(db), mock
}

func TestUserAdapter_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		rows := mock.NewRows([]string{"id", "email", "password", "name", "gender", "age"}).
			AddRow(1, "test@test.com", "Pa55w0rd!", "test", "male", 25)

		mock.
			ExpectQuery(`INSERT INTO "user" \("email", "password", "name", "gender", "age"\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING "id", "email", "password", "name", "gender", "age"`).
			WillReturnRows(rows)

		got, err := sut.CreateUser(
			context.Background(),
			&muzz.CreateUserInput{
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
			},
		)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, &muzz.User{
			ID:       1,
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   muzz.GenderMale,
			Age:      25,
		}, got)
	})
}
