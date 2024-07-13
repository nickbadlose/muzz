package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/apperror"
	mockauth "github.com/nickbadlose/muzz/internal/auth/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	viper.Set("DOMAIN_NAME", "https://test.com")
	viper.Set("JWT_DURATION", "12h")
	viper.Set("JWT_SECRET", "test")
}

func newTestAuthenticator(t *testing.T) (*Authoriser, *mockauth.Repository) {
	cfg, err := config.Load()
	require.NoError(t, err)
	m := mockauth.NewRepository(t)
	au, err := NewAuthoriser(cfg, m)
	require.NoError(t, err)
	return au, m
}

func newTestAuthorizer(t *testing.T) *Authoriser {
	cfg, err := config.Load()
	require.NoError(t, err)
	au, err := NewAuthoriser(cfg, mockauth.NewRepository(t))
	require.NoError(t, err)
	return au
}

func TestAuthorizer_Authenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		au, m := newTestAuthenticator(t)

		m.EXPECT().Authenticate(
			mock.Anything,
			"test@test.com",
			"Pa55w0rd!",
		).Return(&muzz.User{ID: 1}, nil)

		token, user, err := au.Authenticate(
			context.Background(),
			"test@test.com",
			"Pa55w0rd!",
		)

		require.Nil(t, err)
		require.NotNil(t, user)
		require.NotEmpty(t, token)
		require.Equal(t, 1, user.ID)

		claims := &Claims{}
		tkn, pErr := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) {
			return []byte("test"), nil
		}, jwt.WithLeeway(5*time.Second))
		require.NoError(t, pErr)
		require.Equal(t, 1, claims.UserID)
		require.Equal(t, "https://test.com", claims.Issuer)
		require.Equal(t, jwt.ClaimStrings{"https://test.com"}, claims.Audience)
		require.True(t, tkn.Valid)
	})

	t.Run("failure: incorrect credentials", func(t *testing.T) {
		au, m := newTestAuthenticator(t)

		m.EXPECT().Authenticate(
			mock.Anything,
			"test@test.com",
			"Pa55w0rd!",
		).Return(nil, apperror.ErrNoResults)

		token, user, err := au.Authenticate(
			context.Background(),
			"test@test.com",
			"Pa55w0rd!",
		)

		require.Nil(t, user)
		require.Empty(t, token)
		require.Error(t, err)
		require.Equal(t, "incorrect credentials", err.Error())
		require.Equal(t, apperror.StatusUnauthorized, err.Status())
		require.Empty(t, token)
	})
}

func TestAuthorizer_Authorize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		au, m := newTestAuthenticator(t)

		m.EXPECT().Authenticate(
			mock.Anything,
			"test@test.com",
			"Pa55w0rd!",
		).Return(&muzz.User{ID: 1}, nil)

		token, _, err := au.Authenticate(
			context.Background(),
			"test@test.com",
			"Pa55w0rd!",
		)
		require.Nil(t, err)
		require.NotEmpty(t, token)

		userID, err := au.Authorise(token)
		assert.Nil(t, err)
		assert.Equal(t, 1, userID)
	})

	tNow := time.Now().UTC()

	errCases := []struct {
		name       string
		claims     jwt.Claims
		secret     string
		errMessage string
	}{
		{
			name:       "failure: incorrect signature (secret)",
			claims:     Claims{},
			secret:     "incorrect secret",
			errMessage: "token signature is invalid: signature is invalid",
		},
		{
			name:       "failure invalid claims: no user",
			claims:     Claims{},
			secret:     "test",
			errMessage: "token has no user associated with it",
		},
		{
			name: "failure invalid claims: incorrect issuer",
			claims: &Claims{
				UserID: 1,
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(tNow),
					NotBefore: jwt.NewNumericDate(tNow),
					ExpiresAt: jwt.NewNumericDate(tNow.Add(5 * time.Minute)),
					Issuer:    "incorrect issuer",
					Audience:  []string{"https://test.com"},
				},
			},
			secret:     "test",
			errMessage: "token has invalid issuer",
		},
		{
			name: "failure invalid claims: incorrect audience",
			claims: &Claims{
				UserID: 1,
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(tNow),
					NotBefore: jwt.NewNumericDate(tNow),
					ExpiresAt: jwt.NewNumericDate(tNow.Add(5 * time.Minute)),
					Issuer:    "https://test.com",
					Audience:  []string{"incorrect audience"},
				},
			},
			secret:     "test",
			errMessage: "token has invalid audience",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, tc.claims)

			token, err := tkn.SignedString([]byte(tc.secret))
			require.NoError(t, err)

			au := newTestAuthorizer(t)
			userID, aErr := au.Authorise(token)
			require.Error(t, aErr)
			require.Equal(t, tc.errMessage, aErr.Error())
			require.Equal(t, apperror.StatusUnauthorized, aErr.Status())
			require.Empty(t, userID)
		})
	}

	t.Run("failure invalid claims: issued at plus JWTDuration has passed", func(t *testing.T) {
		viper.Set("JWT_DURATION", "1ns")
		claims := &Claims{
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(tNow),
				NotBefore: jwt.NewNumericDate(tNow),
				ExpiresAt: jwt.NewNumericDate(tNow.Add(5 * time.Minute)),
				Issuer:    "https://test.com",
				Audience:  []string{"https://test.com"},
			},
		}

		tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		token, err := tkn.SignedString([]byte("test"))
		require.NoError(t, err)

		au := newTestAuthorizer(t)
		userID, aErr := au.Authorise(token)
		require.Error(t, aErr)
		require.Equal(t, "token is expired", aErr.Error())
		require.Equal(t, apperror.StatusUnauthorized, aErr.Status())
		require.Empty(t, userID)

		viper.Set("JWT_DURATION", "12h")
	})
}

func TestAuthorizer_Context(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		au := newTestAuthorizer(t)
		ctxWithUser := au.UserOnContext(ctx, 1)

		userID, err := au.UserFromContext(ctxWithUser)
		require.NoError(t, err)
		require.Equal(t, 1, userID)
	})

	t.Run("failure: no user on context", func(t *testing.T) {
		ctx := context.Background()
		au := newTestAuthorizer(t)
		userID, err := au.UserFromContext(ctx)
		require.Error(t, err)
		require.Empty(t, userID)
		require.Equal(t, "authenticated user not on context", err.Error())
	})
}
