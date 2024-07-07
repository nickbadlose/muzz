package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

// TODO
//  Once we have repository interface in service, we don't need to abstract from upper/io? So much,
//  we can use its full types in here worry free and the database package will be a lot easier to use

const (
	userTable = "user"
)

// UserAdapter adapts a *database.Database to the service.UserRepository interface.
type UserAdapter struct {
	database *database.Database
}

func NewUserAdapter(d *database.Database) *UserAdapter {
	return &UserAdapter{database: d}
}

type userEntity struct {
	id       int
	email    string
	password string
	name     string
	gender   string
	age      int
}

func (ua *UserAdapter) CreateUser(ctx context.Context, in *muzz.CreateUserInput) (*muzz.User, error) {
	w, err := ua.database.WriteSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []string{"id", "email", "password", "name", "gender", "age"}
	row, err := w.InsertInto(userTable).
		Columns(columns[1:]...).
		Values(in.Email, in.Password, in.Name, in.Gender, in.Age).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(userEntity)
	err = row.Scan(&entity.id, &entity.email, &entity.password, &entity.name, &entity.gender, &entity.age)
	if err != nil {
		return nil, err
	}

	return &muzz.User{
		ID:       entity.id,
		Email:    entity.email,
		Password: entity.password,
		Name:     entity.name,
		Gender:   muzz.GenderValues[entity.gender],
		Age:      entity.age,
	}, nil
}

func (ua *UserAdapter) UserByEmail(ctx context.Context, email string) (*muzz.User, error) {
	r, err := ua.database.ReadSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO index searching by email
	columns := []interface{}{"id", "email", "password", "name", "gender", "age"}
	row, err := r.Select(columns...).From(userTable).Where("email = ?", email).QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(userEntity)
	err = row.Scan(&entity.id, &entity.email, &entity.password, &entity.name, &entity.gender, &entity.age)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NoResults
	}
	if err != nil {
		return nil, err
	}

	gender, ok := muzz.GenderValues[entity.gender]
	if !ok {
		logger.Warn(ctx, "unsupported gender retrieved from database", zap.String("gender", gender.String()))
	}

	return &muzz.User{
		ID:       entity.id,
		Email:    entity.email,
		Password: entity.password,
		Name:     entity.name,
		Gender:   gender,
		Age:      entity.age,
	}, nil
}

// TODO
//  Not required for this project, but in the future you can use this pattern
//  then if you need to use some methods in transactions and isolation.
//  You can pass in database.Reader/Writer into sub fn getUsers(ctx, tx, userID) ([]*userEntity, error)
//  where tx can be either a transaction or read/write session based on the needs of the caller.
//  Parent fn can just create ReadSessionContext or WriteSessionContext or tx and pass in.

func (ua *UserAdapter) GetUsers(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, error) {
	r, err := ua.database.ReadSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO apply filters

	columns := []interface{}{"id", "name", "gender", "age"}
	selector := r.Select(columns...).From(userTable).Where("id != ?", in.UserID)

	selector = applyUserFilters(in.Filters, selector)

	rows, err := selector.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]*userEntity, 0, 1)
	for rows.Next() {
		entity := new(userEntity)
		err = rows.Scan(&entity.id, &entity.name, &entity.gender, &entity.age)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if errors.Is(rows.Err(), sql.ErrNoRows) {
		return nil, apperror.NoResults
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	users := make([]*muzz.UserDetails, 0, len(entities))
	for _, entity := range entities {
		gender, ok := muzz.GenderValues[entity.gender]
		if !ok {
			logger.Warn(ctx, "unsupported gender retrieved from database", zap.String("gender", gender.String()))
		}

		users = append(users, &muzz.UserDetails{
			ID:     entity.id,
			Name:   entity.name,
			Gender: gender,
			Age:    entity.age,
		})
	}

	return users, nil
}

func applyUserFilters(in *muzz.UserFilters, selector database.Selector) database.Selector {
	if in == nil {
		return selector
	}

	if in.MinAge != 0 {
		selector = selector.And("age >= ?", in.MinAge)
	}

	if in.MaxAge != 0 {
		selector = selector.And("age <= ?", in.MaxAge)
	}

	if len(in.Genders) != 0 {
		genderStrings := make([]string, len(in.Genders))
		for i, gen := range in.Genders {
			genderStrings[i] = gen.String()
		}
		selector = selector.And("gender IN ", genderStrings)
	}

	return selector
}
