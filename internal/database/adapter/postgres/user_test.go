package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
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
	require.NoError(t, err)

	dbase, err := database.New(
		context.Background(),
		&database.Credentials{},
		database.WithClientFunc(func(_ context.Context, _ *database.Config) (db.Session, error) {
			return dbClient, nil
		}),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		mockSQL.ExpectClose()
		require.NoError(t, dbase.Close())
	})

	return dbase, mockSQL
}

func newTestUserAdapter(t *testing.T) (*UserAdapter, sqlmock.Sqlmock) {
	dbase, mock := newTestDB(t)
	ua, err := NewUserAdapter(dbase)
	require.NoError(t, err)
	return ua, mock
}

func TestUserAdapter_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		rows := mock.NewRows([]string{"id", "email", "password", "name", "gender", "age", "location"}).
			AddRow(1, "test@test.com", "Pa55w0rd!", "test", "male", 25, []byte("0101000020E61000002EFF21FDF63514C0355EBA490C224940"))

		mock.ExpectQuery(`INSERT INTO "user" \("email", "password", "name", "gender", "age", "location"\) 
VALUES \(\$1, crypt\(\$2, gen_salt\('bf'\)\), \$3, \$4, \$5, ST_SetSRID\(ST_MakePoint\(\$6,\$7\),\$8\)\) 
RETURNING "id", "email", "password", "name", "gender", "age", "location"`).
			WithArgs("test@test.com", "Pa55w0rd!", "test", "male", 25, -5.0527, 50.266, 4326).
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

	t.Run("query error", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		mock.ExpectQuery(`INSERT INTO "user" \("email", "password", "name", "gender", "age", "location"\) 
VALUES \(\$1, crypt\(\$2, gen_salt\('bf'\)\), \$3, \$4, \$5, ST_SetSRID\(ST_MakePoint\(\$6,\$7\),\$8\)\) 
RETURNING "id", "email", "password", "name", "gender", "age", "location"`).
			WithArgs("test@test.com", "Pa55w0rd!", "test", "male", 25, -5.0527, 50.266, 4326).
			WillReturnError(errors.New("database error"))

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

		require.Error(t, err)
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, "database error", err.Error())
	})
}

func TestUserAdapter_Authenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		rows := mock.NewRows([]string{"id", "email", "password", "name", "gender", "age"}).
			AddRow(1, "test@test.com", "Pa55w0rd!", "test", "male", 25)

		mock.ExpectQuery(`SELECT "id", "email", "password", "name", "gender", "age" FROM "user" 
                                WHERE \(email = \$1 AND password = crypt\(\$2, password\)`).
			WithArgs("test@test.com", "Pa55w0rd!").
			WillReturnRows(rows)

		got, err := sut.Authenticate(context.Background(), "test@test.com", "Pa55w0rd!")

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

	t.Run("success: unsupported gender", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		rows := mock.NewRows([]string{"id", "email", "password", "name", "gender", "age"}).
			AddRow(1, "test@test.com", "Pa55w0rd!", "test", "unsupported", 25)

		mock.ExpectQuery(`SELECT "id", "email", "password", "name", "gender", "age" FROM "user" 
                                WHERE \(email = \$1 AND password = crypt\(\$2, password\)`).
			WithArgs("test@test.com", "Pa55w0rd!").
			WillReturnRows(rows)

		got, err := sut.Authenticate(context.Background(), "test@test.com", "Pa55w0rd!")

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		require.Equal(t, &muzz.User{
			ID:       1,
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   muzz.GenderUndefined,
			Age:      25,
		}, got)
	})

	errCases := []struct {
		name            string
		err             error
		expectedMessage string
	}{
		{
			name:            "query error",
			err:             errors.New("database error"),
			expectedMessage: "database error",
		},
		{
			name:            "error no rows",
			err:             sql.ErrNoRows,
			expectedMessage: "no results found in database",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, mock := newTestUserAdapter(t)

			mock.ExpectQuery(`SELECT "id", "email", "password", "name", "gender", "age" FROM "user" 
                                WHERE \(email = \$1 AND password = crypt\(\$2, password\)`).
				WithArgs("test@test.com", "Pa55w0rd!").
				WillReturnError(tc.err)

			got, err := sut.Authenticate(
				context.Background(),
				"test@test.com",
				"Pa55w0rd!",
			)

			require.Error(t, err)
			require.Nil(t, got)
			require.NoError(t, mock.ExpectationsWereMet())
			require.Equal(t, tc.expectedMessage, err.Error())
		})
	}
}

func TestUserAdapter_GetUsers(t *testing.T) {
	successCases := []struct {
		name      string
		input     *muzz.GetUsersInput
		expected  []*muzz.UserDetails
		setupMock func(m sqlmock.Sqlmock)
	}{
		{
			name: "success: no filters",
			input: &muzz.GetUsersInput{
				UserID:   1,
				Location: orb.Point{-5.0527, 50.266},
				SortType: muzz.SortTypeDistance,
			},
			expected: []*muzz.UserDetails{
				{
					ID:             2,
					Name:           "test2",
					Gender:         muzz.GenderMale,
					Age:            25,
					DistanceFromMe: 3.123,
				},
				{
					ID:             6,
					Name:           "test6",
					Gender:         muzz.GenderUndefined,
					Age:            34,
					DistanceFromMe: 5.356,
				},
			},
			setupMock: func(m sqlmock.Sqlmock) {
				rows := m.NewRows([]string{"id", "name", "gender", "age", "distance"}).
					AddRows(
						[][]driver.Value{
							{2, "test2", "male", 25, 3.123},
							{6, "test6", "unsupportedGender", 34, 5.356},
						}...,
					)

				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
		WHERE \(u.id != \$4 AND u.id NOT IN \(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\)\) 
		ORDER BY "distance" ASC LIMIT 50`).
					WithArgs(-5.0527, 50.266, 4326, 1, 1).
					WillReturnRows(rows)
			},
		},
		{
			name: "success: with filters",
			input: &muzz.GetUsersInput{
				UserID:   1,
				Location: orb.Point{-5.0527, 50.266},
				SortType: muzz.SortTypeDistance,
				Filters: &muzz.UserFilters{
					MaxAge:  30,
					MinAge:  20,
					Genders: []muzz.Gender{muzz.GenderMale, muzz.GenderFemale},
				},
			},
			expected: []*muzz.UserDetails{
				{
					ID:             2,
					Name:           "test2",
					Gender:         muzz.GenderMale,
					Age:            25,
					DistanceFromMe: 3.123,
				},
			},
			setupMock: func(m sqlmock.Sqlmock) {
				rows := m.NewRows([]string{"id", "name", "gender", "age", "distance"}).
					AddRows([][]driver.Value{{2, "test2", "male", 25, 3.123}}...)

				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
		WHERE \(u.id != \$4 AND u.id NOT IN \(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\) 
		AND u.age >= \$6 AND u.age <= \$7 AND "u"."gender" IN \(\$8, \$9\)\) ORDER BY "distance" ASC LIMIT 50`).
					WithArgs(-5.0527, 50.266, 4326, 1, 1, 20, 30, "male", "female").
					WillReturnRows(rows)
			},
		},
		{
			name: "success: sort attractiveness",
			input: &muzz.GetUsersInput{
				UserID:   1,
				Location: orb.Point{-5.0527, 50.266},
				SortType: muzz.SortTypeAttractiveness,
			},
			expected: []*muzz.UserDetails{
				{
					ID:             6,
					Name:           "test6",
					Gender:         muzz.GenderMale,
					Age:            34,
					DistanceFromMe: 5.356,
				},
				{
					ID:             2,
					Name:           "test2",
					Gender:         muzz.GenderMale,
					Age:            25,
					DistanceFromMe: 3.123,
				},
			},
			setupMock: func(m sqlmock.Sqlmock) {
				rows := m.NewRows([]string{"id", "name", "gender", "age", "distance"}).
					AddRows(
						[][]driver.Value{
							{6, "test6", "male", 34, 5.356},
							{2, "test2", "male", 25, 3.123},
						}...,
					)

				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
        LEFT JOIN "swipe" AS "s" ON \(u.id = s.swiped_user_id\) WHERE \(u.id != \$4 AND u.id NOT IN 
		\(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\)\) GROUP BY "u"."id" 
		ORDER BY \(NULLIF\(sum\(case when s.preference then 1 else 0 end\),0\)::float / COUNT\(u.id\)::float\) 
		DESC LIMIT 50`).
					WithArgs(-5.0527, 50.266, 4326, 1, 1).
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range successCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, mock := newTestUserAdapter(t)

			tc.setupMock(mock)

			got, err := sut.GetUsers(context.Background(), tc.input)

			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
			require.Equal(t, tc.expected, got)
		})
	}

	errCases := []struct {
		name            string
		setupMock       func(m sqlmock.Sqlmock)
		expectedMessage string
	}{
		{
			name: "query error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
		WHERE \(u.id != \$4 AND u.id NOT IN \(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\)\) 
		ORDER BY "distance" ASC LIMIT 50`).
					WithArgs(-5.0527, 50.266, 4326, 1, 1).
					WillReturnError(errors.New("query error"))
			},
			expectedMessage: "query error",
		},
		{
			name: "row error",
			setupMock: func(m sqlmock.Sqlmock) {
				rows := m.NewRows([]string{"id", "name", "gender", "age", "distance"}).
					AddRow(6, "test6", "male", 34, 5.356)

				rows.RowError(0, errors.New("row error"))

				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
		WHERE \(u.id != \$4 AND u.id NOT IN \(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\)\) 
		ORDER BY "distance" ASC LIMIT 50`).
					WillReturnRows(rows).RowsWillBeClosed()
			},
			expectedMessage: "row error",
		},
		{
			name: "error no rows",
			setupMock: func(m sqlmock.Sqlmock) {
				emptyRows := m.NewRows([]string{"id", "name", "gender", "age", "distance"})

				m.ExpectQuery(`SELECT "u"."id", "u"."name", "u"."gender", "u"."age", 
       \(u.location <-> ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\)\) / 1000 AS distance FROM "user" AS "u" 
		WHERE \(u.id != \$4 AND u.id NOT IN \(SELECT swiped_user_id FROM swipe WHERE user_id = \$5\)\) 
		ORDER BY "distance" ASC LIMIT 50`).
					WillReturnRows(emptyRows)
			},
			expectedMessage: "no results found in database",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, mock := newTestUserAdapter(t)

			tc.setupMock(mock)

			got, err := sut.GetUsers(
				context.Background(),
				&muzz.GetUsersInput{
					UserID:   1,
					Location: orb.Point{-5.0527, 50.266},
					SortType: muzz.SortTypeDistance,
				},
			)

			require.Error(t, err)
			require.Nil(t, got)
			require.NoError(t, mock.ExpectationsWereMet())
			require.Equal(t, tc.expectedMessage, err.Error())
		})
	}
}

func TestUserAdapter_UpdateUserLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)
		mock.ExpectExec(`UPDATE "user" SET location = ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\) WHERE \(id = \$4\)`).
			WithArgs(-5.0527, 50.266, 4326, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := sut.UpdateUserLocation(context.Background(), 1, orb.Point{-5.0527, 50.266})

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		sut, mock := newTestUserAdapter(t)

		mock.ExpectExec(`UPDATE "user" SET location = ST_SetSRID\(ST_MakePoint\(\$1,\$2\),\$3\) WHERE \(id = \$4\)`).
			WithArgs(-5.0527, 50.266, 4326, 1).
			WillReturnError(errors.New("exec error"))

		err := sut.UpdateUserLocation(context.Background(), 1, orb.Point{-5.0527, 50.266})

		require.Error(t, err)
		require.Equal(t, "exec error", err.Error())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
