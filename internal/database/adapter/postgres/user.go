package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
	"github.com/upper/db/v4"
	"go.uber.org/zap"
)

// TODO
//  Once we have repository interface in service, we don't need to abstract from upper/io? So much,
//  we can use its full types in here worry free and the database package will be a lot easier to use

const (
	userTable = "user"

	srid = 4326
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
	location orb.Point
	distance float64
}

func (ua *UserAdapter) CreateUser(ctx context.Context, in *muzz.CreateUserInput) (*muzz.User, error) {
	w, err := ua.database.WriteSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []string{"id", "email", "password", "name", "gender", "age", "location"}
	row, err := w.InsertInto(userTable).
		Columns(columns[1:]...).
		Values(
			in.Email,
			in.Password,
			in.Name,
			in.Gender,
			in.Age,
			pointValue(in.Location),
		).
		Returning(columns...).
		QueryRowContext(ctx)
	if err != nil {
		return nil, err
	}

	entity := new(userEntity)
	err = row.Scan(
		&entity.id,
		&entity.email,
		&entity.password,
		&entity.name,
		&entity.gender,
		&entity.age,
		ewkb.Scanner(&entity.location),
	)
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
		Location: entity.location,
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
//  Exclude already swiped users from results.
//  Have tie breaker order by column? ID check if that's default anyway
//  index for swiped user_id (SELECT swiped_user_id FROM swipe WHERE user_id = ?) seed data and analyze before and after indexes for distance, swiped user etc.

func (ua *UserAdapter) GetUsers(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, error) {
	r, err := ua.database.ReadSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []interface{}{
		"id",
		"name",
		"gender",
		"age",
		// TODO we are officially directly dependant on the upper lib here. So do we remove abstraction interface and just use lib?
		db.Raw(`(location::geography <-> ST_SetSRID(ST_MakePoint(?,?),?)::geography) / 1000 AS distance`, in.Location.Lon(), in.Location.Lat(), srid),
	}
	selector := r.Select(columns...).
		From(userTable).
		Where("id != ?", in.UserID).
		And("id NOT IN ?", db.Raw("(SELECT swiped_user_id FROM swipe WHERE user_id = ?)", in.UserID)).
		OrderBy("distance")

	selector = applyUserFilters(in.Filters, selector)

	rows, err := selector.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]*userEntity, 0, 1)
	for rows.Next() {
		entity := new(userEntity)
		err = rows.Scan(&entity.id, &entity.name, &entity.gender, &entity.age, &entity.distance)
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

	users := make([]*muzz.UserDetails, len(entities))
	for i, entity := range entities {
		gender, ok := muzz.GenderValues[entity.gender]
		if !ok {
			logger.Warn(ctx, "unsupported gender retrieved from database", zap.String("gender", gender.String()))
		}

		users[i] = &muzz.UserDetails{
			ID:             entity.id,
			Name:           entity.name,
			Gender:         gender,
			Age:            entity.age,
			DistanceFromMe: entity.distance,
		}
	}

	return users, nil
}

func (ua *UserAdapter) UpdateUserLocation(ctx context.Context, id int, location orb.Point) error {
	w, err := ua.database.WriteSessionContext(ctx)
	if err != nil {
		return err
	}

	_, err = w.Update(userTable).
		Set(db.Raw("location = ?", pointValue(location))).
		Where("id = ?", id).
		ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
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

func pointValue(p orb.Point) *db.RawExpr {
	return db.Raw("ST_SetSRID(ST_MakePoint(?,?),?)", p.Lon(), p.Lat(), srid)
}
