package app

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/pkg/auth"
	"github.com/nickbadlose/muzz/internal/store"
	mockstore "github.com/nickbadlose/muzz/internal/store/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestService(m *mockstore.Store) Service {
	cfg := config.Load()
	au := auth.NewAuthoriser(cfg)
	return NewService(m, au)
}

func TestNewService(t *testing.T) {
	svc := newTestService(&mockstore.Store{})
	require.NotNil(t, svc)
}

func TestService_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockstore.NewStore(t)
		sut := newTestService(m)

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
			Gender:   "male",
			Age:      25,
		})
		require.NoError(t, err)

		require.NotNil(t, got)
		require.Equal(t, got, &User{
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
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "not a valid gender"},
			errMessage: "please provide a valid gender from",
		},
		{
			name:       "missing age",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "male", Age: 0},
			errMessage: "the minimum age is 18",
		},
		{
			name:       "invalid age: too low",
			req:        &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "male", Age: 17},
			errMessage: "the minimum age is 18",
		},
	}

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			sut := newTestService(&mockstore.Store{})

			got, err := sut.CreateUser(context.Background(), tc.req)
			require.Nil(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), ErrorStatusBadRequest)
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
			req:  &CreateUserRequest{Email: "test@test.com", Password: "Pa55w0rd!", Name: "test", Gender: "female", Age: 25},
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
			sut := newTestService(m)

			got, err := sut.CreateUser(context.Background(), tc.req)
			require.Nil(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), ErrorStatusInternal)
		})
	}
}

func init() {
	viper.Set("DOMAIN_NAME", "https://test.com")
	viper.Set("JWT_DURATION", "12h")
	viper.Set("JWT_SECRET", "test")
}

func TestService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockstore.NewStore(t)
		sut := newTestService(m)

		m.EXPECT().
			GetUserByEmail(mock.Anything, "test@test.com").Once().Return(
			&store.User{
				ID:       1,
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "male",
				Age:      25,
			}, nil,
		)

		// TODO test claims in auth package, noot here.

		got, err := sut.Login(context.Background(), &LoginRequest{Email: "test@test.com", Password: "Pa55w0rd!"})
		require.NoError(t, err)
		require.NotEmpty(t, got)
		claims := &auth.Claims{}
		tkn, pErr := jwt.ParseWithClaims(got, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test"), nil
		})
		require.NoError(t, pErr)
		require.Equal(t, 1, claims.UserID)
		require.Equal(t, "https://test.com", claims.Issuer)
		require.Equal(t, jwt.ClaimStrings{"https://test.com"}, claims.Audience)
		require.True(t, tkn.Valid)
	})

	validationCases := []struct {
		name       string
		req        *LoginRequest
		errMessage string
	}{
		{
			name:       "empty email",
			req:        &LoginRequest{},
			errMessage: "email is a required field",
		},
		{
			name:       "empty password",
			req:        &LoginRequest{Email: "test@test.com"},
			errMessage: "password is a required field",
		},
	}

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			sut := newTestService(&mockstore.Store{})

			got, err := sut.Login(context.Background(), tc.req)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), ErrorStatusBadRequest)
		})
	}

	errCases := []struct {
		name           string
		req            *LoginRequest
		setupMockStore func(*mockstore.Store)
		errMessage     string
		errStatus      ErrorStatus
	}{
		{
			name: "error from store",
			req:  &LoginRequest{Email: "test@test.com", Password: "Pa55w0rd!"},
			setupMockStore: func(m *mockstore.Store) {
				m.EXPECT().GetUserByEmail(mock.Anything, "test@test.com").
					Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  ErrorStatusInternal,
		},
		{
			name: "incorrect email",
			req:  &LoginRequest{Email: "wrong@email.com", Password: "Pa55w0rd!"},
			setupMockStore: func(m *mockstore.Store) {
				m.EXPECT().GetUserByEmail(mock.Anything, "wrong@email.com").
					Once().Return(nil, store.ErrorNotFound)
			},
			errMessage: "incorrect email or password",
			errStatus:  ErrorStatusUnauthorised,
		},
		{
			name: "incorrect password",
			req:  &LoginRequest{Email: "test@test.com", Password: "Pa55w0rd!"},
			setupMockStore: func(m *mockstore.Store) {
				m.EXPECT().GetUserByEmail(mock.Anything, "test@test.com").
					Once().Return(&store.User{
					ID:       0,
					Email:    "test@test.com",
					Password: "SomeOtherPa55w0rd!",
					Name:     "",
					Gender:   "",
					Age:      0,
				}, nil)
			},
			errMessage: "incorrect email or password",
			errStatus:  ErrorStatusUnauthorised,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockstore.NewStore(t)
			tc.setupMockStore(m)
			sut := newTestService(m)

			got, err := sut.Login(context.Background(), tc.req)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), tc.errStatus)
		})
	}
}

func TestService_Discover(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockstore.NewStore(t)

		m.EXPECT().
			GetUsers(mock.Anything, 1).Once().Return(
			[]*store.UserDetails{
				{
					ID:     2,
					Name:   "test",
					Gender: "female",
					Age:    25,
				},
				{
					ID:     3,
					Name:   "test",
					Gender: "unknown gender",
					Age:    25,
				},
			},
			nil,
		)

		sut := newTestService(m)

		got, err := sut.GetUsers(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, []*UserDetails{
			{
				ID:     2,
				Name:   "test",
				Gender: GenderFemale,
				Age:    25,
			},
			{
				ID:     3,
				Name:   "test",
				Gender: GenderUndefined,
				Age:    25,
			},
		}, got)
	})

	validationCases := []struct {
		name       string
		req        int
		errMessage string
	}{
		{
			name:       "empty id",
			req:        0,
			errMessage: "user id is required",
		},
	}

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			sut := newTestService(&mockstore.Store{})

			got, err := sut.GetUsers(context.Background(), tc.req)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), ErrorStatusBadRequest)
		})
	}

	errCases := []struct {
		name           string
		setupMockStore func(*mockstore.Store)
		errMessage     string
		errStatus      ErrorStatus
	}{
		{
			name: "error from store",
			setupMockStore: func(m *mockstore.Store) {
				m.EXPECT().GetUsers(mock.Anything, 1).
					Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  ErrorStatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockstore.NewStore(t)
			tc.setupMockStore(m)
			sut := newTestService(m)

			got, err := sut.GetUsers(context.Background(), 1)
			require.Empty(t, got)
			require.Contains(t, err.Error(), tc.errMessage)
			require.Equal(t, err.Status(), tc.errStatus)
		})
	}
}
