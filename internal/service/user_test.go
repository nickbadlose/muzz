package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	mockservice "github.com/nickbadlose/muzz/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockservice.NewUserRepository(t)
		sut := NewUserService(m)

		m.EXPECT().
			CreateUser(mock.Anything, &muzz.CreateUserInput{
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
			}).Once().Return(
			&muzz.User{
				ID:       1,
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   muzz.GenderMale,
				Age:      25,
			}, nil,
		)

		got, err := sut.Create(context.Background(), &muzz.CreateUserInput{
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   "male",
			Age:      25,
		})
		require.Nil(t, err)

		require.NotNil(t, got)
		require.Equal(t, got, &muzz.User{
			ID:       1,
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   muzz.GenderMale,
			Age:      25,
		})
	})

	errCases := []struct {
		name          string
		input         *muzz.CreateUserInput
		setupMockRepo func(m *mockservice.UserRepository)
		errMessage    string
		errStatus     apperror.Status
	}{
		{
			name:          "invalid input",
			input:         &muzz.CreateUserInput{},
			setupMockRepo: func(m *mockservice.UserRepository) {},
			errMessage:    "email is a required field",
			errStatus:     apperror.StatusBadInput,
		},
		{
			name:  "error from repository",
			input: &muzz.CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "test", Gender: "female", Age: 25},
			setupMockRepo: func(m *mockservice.UserRepository) {
				m.EXPECT().CreateUser(mock.Anything, &muzz.CreateUserInput{
					Email:    "test@test.com",
					Password: "Pa55w0rd!",
					Name:     "test",
					Gender:   "female",
					Age:      25,
				}).Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockservice.NewUserRepository(t)
			tc.setupMockRepo(m)
			sut := NewUserService(m)

			got, err := sut.Create(context.Background(), tc.input)
			require.Nil(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), tc.errStatus)
		})
	}
}

func TestUserService_Discover(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockservice.NewUserRepository(t)

		m.EXPECT().
			GetUsers(mock.Anything, &muzz.GetUsersInput{
				UserID:  1,
				Filters: nil,
			}).Once().Return(
			[]*muzz.UserDetails{
				{
					ID:     2,
					Name:   "test",
					Gender: muzz.GenderFemale,
					Age:    25,
				},
				{
					ID:     3,
					Name:   "test",
					Gender: muzz.GenderUndefined,
					Age:    25,
				},
			},
			nil,
		)

		sut := NewUserService(m)

		got, err := sut.Discover(context.Background(), &muzz.GetUsersInput{
			UserID:  1,
			Filters: nil,
		})
		require.Nil(t, err)
		require.Equal(t, []*muzz.UserDetails{
			{
				ID:     2,
				Name:   "test",
				Gender: muzz.GenderFemale,
				Age:    25,
			},
			{
				ID:     3,
				Name:   "test",
				Gender: muzz.GenderUndefined,
				Age:    25,
			},
		}, got)
	})

	errCases := []struct {
		name          string
		input         *muzz.GetUsersInput
		setupMockRepo func(*mockservice.UserRepository)
		errMessage    string
		errStatus     apperror.Status
	}{
		{
			name:          "invalid input",
			input:         &muzz.GetUsersInput{},
			setupMockRepo: func(m *mockservice.UserRepository) {},
			errMessage:    "user id is a required field",
			errStatus:     apperror.StatusBadInput,
		},
		{
			name: "error from repository",
			input: &muzz.GetUsersInput{
				UserID:  1,
				Filters: nil,
			},
			setupMockRepo: func(m *mockservice.UserRepository) {
				m.EXPECT().GetUsers(mock.Anything, &muzz.GetUsersInput{
					UserID:  1,
					Filters: nil,
				}).
					Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockservice.NewUserRepository(t)
			tc.setupMockRepo(m)
			sut := NewUserService(m)

			got, err := sut.Discover(context.Background(), tc.input)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), tc.errStatus)
		})
	}
}
