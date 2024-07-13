package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	mockservice "github.com/nickbadlose/muzz/internal/service/mocks"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mu := mockservice.NewUserRepository(t)
		ma := mockservice.NewAuthenticator(t)
		sut, err := NewAuthService(ma, mu)
		require.NoError(t, err)

		ma.EXPECT().
			Authenticate(mock.Anything, "test@test.com", "Pa55w0rd!").Once().Return(
			&muzz.User{
				ID:       1,
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   muzz.GenderMale,
				Age:      25,
			}, nil,
		).Once().Return("token", &muzz.User{ID: 1}, nil)

		mu.EXPECT().UpdateUserLocation(mock.Anything, 1, orb.Point{1, 1}).
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
		setupMockRepo func(*mockservice.Authenticator, *mockservice.UserRepository)
		errMessage    string
		errStatus     apperror.Status
	}{
		{
			name:          "invalid input",
			input:         &muzz.LoginInput{},
			setupMockRepo: func(ma *mockservice.Authenticator, mu *mockservice.UserRepository) {},
			errMessage:    "email is a required field",
			errStatus:     apperror.StatusBadInput,
		},
		{
			name:  "error from authenticator",
			input: &muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{0, 0}},
			setupMockRepo: func(ma *mockservice.Authenticator, mu *mockservice.UserRepository) {
				ma.EXPECT().
					Authenticate(mock.Anything, "test@test.com", "Pa55w0rd!").
					Once().
					Return(
						"",
						nil,
						apperror.NewErr(apperror.StatusInternal, errors.New("authenticator error")),
					)
			},
			errMessage: "authenticator error",
			errStatus:  apperror.StatusInternal,
		},
		{
			name:  "error from user repository - update user location",
			input: &muzz.LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{1, 1}},
			setupMockRepo: func(ma *mockservice.Authenticator, mu *mockservice.UserRepository) {
				ma.EXPECT().
					Authenticate(mock.Anything, "test@test.com", "Pa55w0rd!").
					Once().Return("token", &muzz.User{
					ID:       1,
					Email:    "test@test.com",
					Password: "Pa55w0rd!",
					Name:     "test",
					Gender:   muzz.GenderMale,
					Age:      25,
				}, nil)

				mu.EXPECT().UpdateUserLocation(mock.Anything, 1, orb.Point{1, 1}).
					Once().Return(errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			ma := mockservice.NewAuthenticator(t)
			mu := mockservice.NewUserRepository(t)
			tc.setupMockRepo(ma, mu)
			sut, err := NewAuthService(ma, mu)
			require.NoError(t, err)

			got, aErr := sut.Login(context.Background(), tc.input)
			require.Empty(t, got)
			require.Contains(t, aErr.Error(), tc.errMessage)
			require.Equal(t, aErr.Status(), tc.errStatus)
		})
	}
}
