package app

import (
	"context"

	"github.com/nickbadlose/muzz/internal/pkg/logger"
	"github.com/nickbadlose/muzz/internal/store"
	"go.uber.org/zap"
)

type Service interface {
	CreateUser(context.Context, *CreateUserRequest) (*User, error)
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service { return &service{store} }

// CreateUser takes a users details, validates them and creates a new user record in the database.
func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	logger.Debug(ctx, "request", zap.Any("request", req))

	err := req.validate()
	if err != nil {
		logger.Error(ctx, "validating request", err)
		return nil, err
	}

	// TODO encrypt password
	// TODO search by email index for login func

	storeUser, err := s.store.CreateUser(ctx, &store.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender.String(),
		Age:      req.Age,
	})
	if err != nil {
		logger.Error(ctx, "creating user in database", err)
		return nil, err
	}

	return &User{
		ID:       storeUser.ID,
		Email:    storeUser.Email,
		Password: storeUser.Password,
		Name:     storeUser.Name,
		Gender:   GenderValues[storeUser.Gender],
		Age:      storeUser.Age,
	}, nil
}
