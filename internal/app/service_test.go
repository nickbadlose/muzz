package app

import (
	"context"
	"errors"
	"testing"

	"github.com/nickbadlose/muzz/internal/store"
	mockstore "github.com/nickbadlose/muzz/internal/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	svc := NewService(mockstore.NewStore(t))
	assert.NotNil(t, svc)
}

func TestService_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockstore.NewStore(t)
		sut := NewService(m)

		m.EXPECT().
			CreateUser(mock.Anything, &store.CreateUserInput{
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
			}).Once().Return(
			&store.User{
				ID:       1,
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
			}, nil,
		)

		got, err := sut.CreateUser(context.Background(), &CreateUserRequest{
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   GenderMale,
			Age:      25,
		})
		assert.NoError(t, err)

		assert.NotNil(t, got)
		assert.Equal(t, got, &User{
			ID:       1,
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   GenderMale,
			Age:      25,
		})
	})

	validationCases := []struct {
		name       string
		req        *CreateUserRequest
		errMessage string
	}{
		{
			name:       "empty emails",
			req:        &CreateUserRequest{},
			errMessage: "email is a required field",
		},
		{
			name:       "invalid email: contains spaces",
			req:        &CreateUserRequest{Email: "te st@test.com"},
			errMessage: "email cannot contain spaces",
		},
		{
			name:       "invalid email: no @",
			req:        &CreateUserRequest{Email: "invalidEmail"},
			errMessage: "mail: missing '@' or angle-addr",
		},
		{
			name:       "invalid email: no .",
			req:        &CreateUserRequest{Email: "test@test"},
			errMessage: "invalid email address: missing '.' in email domain",
		},
		{
			name:       "missing password",
			req:        &CreateUserRequest{Email: "test@test.com"},
			errMessage: "password is a required field",
		},
		{
			name:       "invalid password: no upper case",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "passw0rd!"},
			errMessage: "password must contain at least 1 uppercase letter",
		},
		{
			name:       "invalid password: no lower case",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "PASSW0RD!"},
			errMessage: "password must contain at least 1 lowercase letter",
		},
		{
			name:       "invalid password: no special character",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Passw0rd"},
			errMessage: "password must contain at least 1 special character",
		},
		{
			name:       "invalid password: no numbers",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Password!"},
			errMessage: "password must contain at least 1 number",
		},
		{
			name:       "invalid password: not 8 characters",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55!"},
			errMessage: "password must contain at least 8 characters",
		},
		{
			name:       "missing name",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!"},
			errMessage: "name is a required field",
		},
		{
			name:       "missing name",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!"},
			errMessage: "name is a required field",
		},
		{
			name:       "missing gender",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test"},
			errMessage: "please provide a valid gender from",
		},
		{
			name:       "invalid gender: out of range int",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: Gender(100)},
			errMessage: "please provide a valid gender from",
		},
		{
			name:       "missing age",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: GenderMale, Age: 0},
			errMessage: "the minimum age is 18",
		},
		{
			name:       "invalid age: too low",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: GenderMale, Age: 17},
			errMessage: "the minimum age is 18",
		},
	}

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			sut := NewService(mockstore.NewStore(t))

			got, err := sut.CreateUser(context.Background(), tc.req)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tc.errMessage)
		})
	}

	errCases := []struct {
		name           string
		req            *CreateUserRequest
		setupMockStore func(*mockstore.Store)
		errMessage     string
	}{
		{
			name: "error from store",
			req:  &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "test", Gender: GenderFemale, Age: 25},
			setupMockStore: func(m *mockstore.Store) {
				m.EXPECT().CreateUser(mock.Anything, &store.CreateUserInput{
					Email:    "test@test.com",
					Password: "Pa55w0rd!",
					Name:     "test",
					Gender:   "female",
					Age:      25,
				}).Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockstore.NewStore(t)
			tc.setupMockStore(m)
			sut := NewService(m)

			got, err := sut.CreateUser(context.Background(), tc.req)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tc.errMessage)
		})
	}
}
