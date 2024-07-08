package service

import (
	"context"
	"errors"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/paulmach/orb"
	"go.uber.org/zap"
)

// TODO
//  Check out github.com/deepmap/oapi-codegen /
//  https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/master/internal/trainer/ports/openapi_api.gen.go
//  for openai code gen docs
//  Do some sort of docs, if README or swagger or something else

type Authenticator interface {
	Authenticate(user *muzz.User, password string) (string, *apperror.Error)
}

type AuthRepository interface {
	UserByEmail(ctx context.Context, email string) (*muzz.User, error)
	UpdateUserLocation(ctx context.Context, id int, location orb.Point) error
}

type AuthService struct {
	repository    AuthRepository
	authenticator Authenticator
}

func NewAuthService(ar AuthRepository, auth Authenticator) *AuthService {
	return &AuthService{ar, auth}
}

func (as *AuthService) Login(ctx context.Context, in *muzz.LoginInput) (string, *apperror.Error) {
	logger.Debug(ctx, "AuthService Login", zap.Any("request", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating Login request", err)
		return "", apperror.BadInput(err)
	}

	user, err := as.repository.UserByEmail(ctx, in.Email)
	if errors.Is(err, apperror.NoResults) {
		return "", apperror.IncorrectCredentials()
	}
	if err != nil {
		logger.Error(ctx, "getting user from database", err)
		return "", apperror.Internal(err)
	}

	token, aErr := as.authenticator.Authenticate(user, in.Password)
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
