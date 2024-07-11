package service

import (
	"context"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

// TODO
//  Check out github.com/deepmap/oapi-codegen /
//  https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/master/internal/trainer/ports/openapi_api.gen.go
//  for openai code gen docs
//  Do some sort of docs, if README or swagger or something else

type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (token string, user *muzz.User, err *apperror.Error)
}

type AuthService struct {
	repository    UserRepository
	authenticator Authenticator
}

func NewAuthService(auth Authenticator, ur UserRepository) *AuthService {
	return &AuthService{ur, auth}
}

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
