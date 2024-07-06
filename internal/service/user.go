package service

import (
	"context"
	"errors"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

// TODO
//  Check out github.com/deepmap/oapi-codegen /
//  https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/master/internal/trainer/ports/openapi_api.gen.go
//  for openai code gen docs

// TODO try t.Parallel in integration tests

// UserRepository is the interface to write and read data from the database.
type UserRepository interface {
	CreateUser(context.Context, *muzz.CreateUserInput) (*muzz.User, error)
	GetUsers(context.Context, int) ([]*muzz.UserDetails, error)
}

type UserService struct {
	repository UserRepository
}

func NewUserService(s UserRepository) *UserService { return &UserService{s} }

// Create takes a users details, validates them and creates a new user record in the database.
func (us *UserService) Create(ctx context.Context, in *muzz.CreateUserInput) (*muzz.User, *apperror.Error) {
	logger.Debug(ctx, "UserService Create", zap.Any("request", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating create user input", err)
		return nil, apperror.BadInput(err)
	}

	// TODO encrypt password
	// TODO search by email index for login func

	u, err := us.repository.CreateUser(ctx, in)
	if err != nil {
		logger.Error(ctx, "creating user in database", err)
		return nil, apperror.Internal(err)
	}

	return u, nil
}

// TODO
//  Pagination ? Check specs.
//  Exclude already swiped profiles.

func (us *UserService) Discover(ctx context.Context, userID int) ([]*muzz.UserDetails, *apperror.Error) {
	logger.Debug(ctx, "UserService Discover", zap.Int("userID", userID))

	if userID == 0 {
		err := apperror.BadInput(errors.New("user id is required"))
		logger.Error(ctx, "validating GetUsers request", err)
		return nil, err
	}

	users, err := us.repository.GetUsers(ctx, userID)
	if err != nil {
		if errors.Is(err, apperror.NoResults) {
			return nil, apperror.NotFound(err)
		}

		logger.Error(ctx, "getting users from database", err)
		return nil, apperror.Internal(err)
	}

	return users, nil
}
