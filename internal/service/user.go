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

// UserRepository is the interface to write and read user data from the repository.
type UserRepository interface {
	// CreateUser adds a user to the repository.
	CreateUser(context.Context, *muzz.CreateUserInput) (*muzz.User, error)
	// GetUsers from the repository.
	GetUsers(context.Context, *muzz.GetUsersInput) ([]*muzz.UserDetails, error)
	// UpdateUserLocation updates the user records location with the provided data.
	UpdateUserLocation(ctx context.Context, id int, location orb.Point) error
}

// UserService is the service which handles all user based requests.
type UserService struct {
	repository UserRepository
}

// NewUserService builds a new *UserService.
func NewUserService(ur UserRepository) (*UserService, error) {
	if ur == nil {
		return nil, errors.New("user repository can't be nil")
	}
	return &UserService{repository: ur}, nil
}

// Create takes a users details, validates them and creates a new user record in the database.
func (us *UserService) Create(ctx context.Context, in *muzz.CreateUserInput) (*muzz.User, *apperror.Error) {
	logger.Debug(ctx, "UserService Create", zap.Any("input", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating create user input", err)
		return nil, apperror.BadInput(err)
	}

	u, err := us.repository.CreateUser(ctx, in)
	if err != nil {
		logger.Error(ctx, "creating user in database", err)
		return nil, apperror.Internal(err)
	}

	return u, nil
}

// Discover attempts to discover new potential matches for a user, it will only return results that the user hasn't
// already performed a swipe on. Available filters and sort types for the records are on the input.
//
// Returned records will include a DistanceFromMe field, this is the calculated distance of each record from the
// provided location.
func (us *UserService) Discover(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, *apperror.Error) {
	logger.Debug(ctx, "UserService Discover", zap.Any("input", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating get users input", err)
		return nil, apperror.BadInput(err)
	}

	users, err := us.repository.GetUsers(ctx, in)
	if err != nil {
		if errors.Is(err, apperror.ErrNoResults) {
			return nil, apperror.NotFound(err)
		}

		logger.Error(ctx, "getting users from database", err)
		return nil, apperror.Internal(err)
	}

	return users, nil
}
