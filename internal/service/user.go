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

// TODO try t.Parallel in integration tests

// UserRepository is the interface to write and read data from the database.
type UserRepository interface {
	CreateUser(context.Context, *muzz.CreateUserInput) (*muzz.User, error)
	GetUsers(context.Context, *muzz.GetUsersInput) ([]*muzz.UserDetails, error)
	UpdateUserLocation(ctx context.Context, id int, location orb.Point) error
}

type UserService struct {
	repository UserRepository
}

// TODO
//  log.fatal in new funcs or return an error from them.
//  Should we check for nil in pointer receiver methods?

func NewUserService(s UserRepository) *UserService { return &UserService{s} }

// Create takes a users details, validates them and creates a new user record in the database.
func (us *UserService) Create(ctx context.Context, in *muzz.CreateUserInput) (*muzz.User, *apperror.Error) {
	logger.Debug(ctx, "UserService Create", zap.Any("input", in))

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

func (us *UserService) Discover(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, *apperror.Error) {
	logger.Debug(ctx, "UserService Discover", zap.Any("input", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating get users input", err)
		return nil, apperror.BadInput(err)
	}

	users, err := us.repository.GetUsers(ctx, in)
	if err != nil {
		if errors.Is(err, apperror.NoResults) {
			return nil, apperror.NotFound(err)
		}

		logger.Error(ctx, "getting users from database", err)
		return nil, apperror.Internal(err)
	}

	return users, nil
}
