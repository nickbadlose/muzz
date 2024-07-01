package store

import (
	"context"

	"github.com/nickbadlose/muzz/internal/pkg/database"
)

// TODO see if we can abstract out database layer from leaking into application layer
//  via an internal database/store package in the application package? This can implement a new interface and
//  decouples us from any specific database type such as SQL or NoSQL

// Store is the interface to write and read data from the database.
type Store interface {
	CreateUser(context.Context, *CreateUserInput) (*User, error)
}

type store struct {
	database database.Database
}

// New builds and returns a Store interface to use.
func New(database database.Database) Store { return &store{database: database} }

// CreateUserInput represents the details required to create a user in the "User" table.
type CreateUserInput struct {
	Email    string
	Password string
	Name     string
	Gender   string
	Age      int
}

// User represents a row from the "User" table.
type User struct {
	ID       int
	Email    string
	Password string
	Name     string
	Gender   string
	Age      int
}

// CreateUser takes a user input and creates a new record in the database.
func (s *store) CreateUser(ctx context.Context, req *CreateUserInput) (*User, error) {
	w, err := s.database.WriteSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []string{"id", "email", "password", "name", "gender", "age"}
	row, err := w.InsertInto("User").
		Columns(columns[1:]...).
		Values(req.Email, req.Password, req.Name, req.Gender, req.Age).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	res := new(User)
	err = row.Scan(&res.ID, &res.Email, &res.Password, &res.Name, &res.Gender, &res.Age)
	if err != nil {
		return nil, err
	}

	return res, nil
}
