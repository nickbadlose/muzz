package app

import (
	"context"
	"errors"

	"github.com/nickbadlose/muzz/internal/pkg/auth"
	"github.com/nickbadlose/muzz/internal/pkg/logger"
	"github.com/nickbadlose/muzz/internal/store"
	"go.uber.org/zap"
)

type Service interface {
	CreateUser(context.Context, *CreateUserRequest) (*User, Error)
	Login(context.Context, *LoginRequest) (string, Error)
	GetUsers(context.Context, int) ([]*UserDetails, Error)
}

type service struct {
	store store.Store
	auth  auth.Generator
}

func NewService(s store.Store, gen auth.Generator) Service { return &service{s, gen} }

// CreateUser takes a users details, validates them and creates a new user record in the database.
func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, Error) {
	logger.Debug(ctx, "CreateUser", zap.Any("request", req))

	err := req.Validate()
	if err != nil {
		logger.Error(ctx, "validating request", err)
		return nil, errBadRequest(err)
	}

	// TODO encrypt password
	// TODO search by email index for login func

	storeUser, err := s.store.CreateUser(ctx, &store.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender,
		Age:      req.Age,
	})
	if err != nil {
		logger.Error(ctx, "creating user in database", err)
		return nil, errInternal(err)
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

// TODO use custom inputs in integration tests. Un-export Gender type
//  Keep requests and responses exported as the fields will be due to json

func (s *service) Login(ctx context.Context, req *LoginRequest) (string, Error) {
	logger.Debug(ctx, "Login", zap.Any("request", req))

	err := req.Validate()
	if err != nil {
		logger.Error(ctx, "validating request", err)
		return "", errBadRequest(err)
	}

	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, store.ErrorNotFound) {
		return "", errUnauthorised(errors.New("incorrect email or password"))
	}
	if err != nil {
		logger.Error(ctx, "getting user from database", err)
		return "", errInternal(err)
	}

	// TODO when checking if token is valid, check issuedAt + cfg.Expriry

	// TODO compare with encryption lib
	if user.Password != req.Password {
		return "", errUnauthorised(errors.New("incorrect email or password"))
	}

	token, err := s.auth.GenerateJWT(user.ID)
	if err != nil {
		logger.Error(ctx, "generating token", err)
		return "", errInternal(err)
	}

	return token, nil
}

// TODO pagination ? Check specs

func (s *service) GetUsers(ctx context.Context, userID int) ([]*UserDetails, Error) {
	if userID == 0 {
		err := errBadRequest(errors.New("user id is required"))
		logger.Error(ctx, "no user id supplied", err)
		return nil, err
	}

	storeUsers, err := s.store.GetUsers(ctx, userID)
	if err != nil {
		logger.Error(ctx, "getting users from database", err)
		return nil, errInternal(err)
	}

	users := make([]*UserDetails, 0, len(storeUsers))
	for _, user := range storeUsers {
		gender, ok := GenderValues[user.Gender]
		if !ok {
			logger.Error(
				ctx,
				"invalid gender returned from database",
				errors.New("invalid gender"),
				zap.String("gender", gender.String()),
			)
		}
		users = append(users, &UserDetails{ID: user.ID, Name: user.Name, Gender: gender, Age: user.Age})
	}

	return users, nil
}
