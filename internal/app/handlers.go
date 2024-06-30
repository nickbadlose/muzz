package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

type Service interface {
	CreateUser(context.Context, *CreateUserRequest) (*User, error)
}

type service struct {
	database database.Database
}

func NewService(database database.Database) Service { return &service{database} }

func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	logger.Debug(ctx, "request", zap.Any("request", req))

	// TODO see if we can abstract out database layer from leaking into application layer
	//  via an internal database/store package in the application package? This can implement a new interface and
	//  decouples us from any specific database type such as SQL or NoSQL
	w, err := s.database.WriteSessionContext(ctx)
	if err != nil {
		logger.Error(ctx, "failed to initialise write session", zap.Error(err))
		return nil, err
	}

	// TODO encrypt password

	columns := []string{"id", "email", "password", "name", "gender", "age"}
	row, err := w.InsertInto("User").
		Columns(columns[1:]...).
		Values(req.Email, req.Password, req.Name, req.Gender, req.Age).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		logger.Error(ctx, "failed to insert into user", zap.Error(err))
		return nil, err
	}

	res := new(User)
	err = row.Scan(&res.ID, &res.Email, &res.Password, &res.Name, &res.Gender, &res.Age)
	if err != nil {
		logger.Error(ctx, "failed to read user row", zap.Error(err))
		return nil, err
	}

	return res, nil
}

type Handlers interface {
	CreateUser(http.ResponseWriter, *http.Request)
}

type handlers struct {
	service Service
}

func NewHandlers(s Service) Handlers { return &handlers{s} }

func (h *handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := new(CreateUserRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		//	TODO handle error
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		// TODO handle error
	}

	w.WriteHeader(http.StatusCreated)

	render.JSON(w, r, &UserResponse{Result: user})
}
