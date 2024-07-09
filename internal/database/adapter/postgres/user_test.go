package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/database"
	mockdatabase "github.com/nickbadlose/muzz/internal/database/mocks"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
	"github.com/upper/db/v4"
)

func newTestDB(t *testing.T) (*database.Database, sqlmock.Sqlmock) {
	dbClient, mockSQL, err := mockdatabase.NewWrappedMock()

	dbase, err := database.New(
		context.Background(),
		&database.Config{},
		func(_ context.Context, _ *database.Config) (db.Session, error) {
			return dbClient, err
		},
	)

	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		mockSQL.ExpectClose()
		require.NoError(t, dbase.Close())
	})

	return dbase, mockSQL
}

func newTestUserAdapter(t *testing.T) (*UserAdapter, sqlmock.Sqlmock) {
	dbase, mock := newTestDB(t)
	return NewUserAdapter(dbase), mock
}

func TestUserAdapter_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		rows := mock.NewRows([]string{"id", "email", "password", "name", "gender", "age", "location"}).
			AddRow(1, "test@test.com", "Pa55w0rd!", "test", "male", 25, []byte("0101000020E61000002EFF21FDF63514C0355EBA490C224940"))

		mock.ExpectQuery(`INSERT INTO "user" \("email", "password", "name", "gender", "age", "location"\) 
VALUES \(\$1, \$2, \$3, \$4, \$5, ST_SetSRID\(ST_MakePoint\(\$6,\$7\),\$8\)\) 
RETURNING "id", "email", "password", "name", "gender", "age", "location"`).
			WillReturnRows(rows)

		got, err := sut.CreateUser(
			context.Background(),
			&muzz.CreateUserInput{
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
				Location: orb.Point{-5.0527, 50.266},
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
			Location: orb.Point{-5.0527, 50.266},
		}, got)
	})
}
