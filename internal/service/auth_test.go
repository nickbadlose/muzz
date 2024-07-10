package service

import (
	"context"
	"errors"
	"github.com/paulmach/orb"
	"testing"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/auth"
	mockservice "github.com/nickbadlose/muzz/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(t *testing.T, m *mockservice.AuthRepository) *AuthService {
	cfg, err := config.Load()
	require.NoError(t, err)
	a := auth.NewAuthorizer(cfg)
	return NewAuthService(m, a)
}

func TestAuthService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockservice.NewAuthRepository(t)
		sut := newTestAuthService(t, m)

		m.EXPECT().
			UserByEmail(mock.Anything, "test@test.com").Once().Return(
			&muzz.User{
				ID:       1,
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   muzz.GenderMale,
				Age:      25,
			}, nil,
		)

		m.EXPECT().UpdateUserLocation(mock.Anything, 1, orb.Point{1, 1}).
			Once().Return(nil)

		got, err := sut.Login(
			context.Background(),
			&muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{1, 1}},
		)
		require.Nil(t, err)
		require.NotEmpty(t, got)
	})

	errCases := []struct {
		name          string
		input         *muzz.LoginInput
		setupMockRepo func(*mockservice.AuthRepository)
		errMessage    string
		errStatus     apperror.Status
	}{
		{
			name:          "invalid input",
			input:         &muzz.LoginInput{},
			setupMockRepo: func(m *mockservice.AuthRepository) {},
			errMessage:    "email is a required field",
			errStatus:     apperror.StatusBadInput,
		},
		{
			name:  "error from repository - user by email",
			input: &muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{0, 0}},
			setupMockRepo: func(m *mockservice.AuthRepository) {
				m.EXPECT().UserByEmail(mock.Anything, "test@test.com").
					Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
		{
			name:  "incorrect email",
			input: &muzz.LoginInput{Email: "wrong@email.com", Password: "Pa55w0rd!"},
			setupMockRepo: func(m *mockservice.AuthRepository) {
				m.EXPECT().UserByEmail(mock.Anything, "wrong@email.com").
					Once().Return(nil, apperror.NoResults)
			},
			errMessage: "incorrect credentials",
			errStatus:  apperror.StatusUnauthorized,
		},
		{
			name:  "authentication failed",
			input: &muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!"},
			setupMockRepo: func(m *mockservice.AuthRepository) {
				m.EXPECT().UserByEmail(mock.Anything, "test@test.com").
					Once().Return(&muzz.User{
					ID:       0,
					Email:    "test@test.com",
					Password: "SomeOtherPa55w0rd!",
					Name:     "",
					Gender:   muzz.GenderUndefined,
					Age:      0,
				}, nil)
			},
			errMessage: "incorrect credentials",
			errStatus:  apperror.StatusUnauthorized,
		},
		{
			name:  "error from repository - update user location",
			input: &muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{1, 1}},
			setupMockRepo: func(m *mockservice.AuthRepository) {
				m.EXPECT().UserByEmail(mock.Anything, "test@test.com").
					Once().Return(&muzz.User{
					ID:       1,
					Email:    "test@test.com",
					Password: "Pa55w0rd!",
					Name:     "test",
					Gender:   muzz.GenderMale,
					Age:      25,
				}, nil)

				m.EXPECT().UpdateUserLocation(mock.Anything, 1, orb.Point{1, 1}).
					Once().Return(errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockservice.NewAuthRepository(t)
			tc.setupMockRepo(m)
			sut := newTestAuthService(t, m)

			got, err := sut.Login(context.Background(), tc.input)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), tc.errStatus)
		})
	}
}
