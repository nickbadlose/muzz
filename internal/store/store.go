package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nickbadlose/muzz/internal/pkg/database"
)

const (
	userTable = "User"
)

var (
	// ErrorNotFound represents when a requested record does not exist in the database.
	ErrorNotFound = errors.New("record not found")
)

// TODO see if we can abstract out database layer from leaking into application layer
//  via an internal database/store package in the application package? This can implement a new interface and
//  decouples us from any specific database type such as SQL or NoSQL

// Store is the interface to write and read data from the database.
type Store interface {
	CreateUser(context.Context, *CreateUserInput) (*User, error)
	GetUserByEmail(context.Context, string) (*User, error)
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
	row, err := w.InsertInto(userTable).
		Columns(columns[1:]...).
		Values(req.Email, req.Password, req.Name, req.Gender, req.Age).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	user := new(User)
	err = row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Gender, &user.Age)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	r, err := s.database.ReadSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO index searching by email
	columns := []interface{}{"id", "email", "password", "name", "gender", "age"}
	row, err := r.SelectFrom(userTable).Columns(columns...).Where("email = ?", email).QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO check not found?

	user := new(User)
	err = row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Gender, &user.Age)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrorNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
