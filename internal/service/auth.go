package service

import (
	"context"
	"errors"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

// Authenticator is the interface to authenticate a user.
type Authenticator interface {
	// Authenticate the provided user credentials.
	Authenticate(ctx context.Context, email, password string) (token string, user *muzz.User, err *apperror.Error)
}

// AuthService is the service which handles all authentication and authorization based requests.
type AuthService struct {
	authenticator Authenticator
	repository    UserRepository
}

// NewAuthService builds a new *AuthService.
func NewAuthService(auth Authenticator, ur UserRepository) (*AuthService, error) {
	if auth == nil {
		return nil, errors.New("user repository cannot be nil")
	}
	if ur == nil {
		return nil, errors.New("authenticator cannot be nil")
	}
	return &AuthService{repository: ur, authenticator: auth}, nil
}

// Login takes a users credentials, authenticates them and returns a token on success.
//
// If the user is successfully authenticated, the users location data is updated in the database.
func (as *AuthService) Login(ctx context.Context, in *muzz.LoginInput) (string, *apperror.Error) {
	logger.Debug(ctx, "AuthService Login", zap.Any("request", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating Login request", err)
		return "", apperror.BadInput(err)
	}

	token, user, aErr := as.authenticator.Authenticate(ctx, in.Email, in.Password)
	if aErr != nil {
		logger.Error(ctx, "authenticating user", aErr)
		return "", aErr
	}

	err = as.repository.UpdateUserLocation(ctx, user.ID, in.Location)
	if err != nil {
		logger.Error(ctx, "updating user location", err)
		return "", apperror.Internal(err)
	}

	return token, nil
}
