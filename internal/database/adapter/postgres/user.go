package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	id             int
	email          string
	password       string
	name           string
	gender         string
	age            int
	location       orb.Point
	distance       float64
	attractiveness float64
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
//  Check all make functions and use correct methods ie len or cap with append or [i]
//  Do we need entities, we are just converting to another object immediately

func (ua *UserAdapter) GetUsers(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, error) {
	r, err := ua.database.ReadSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []interface{}{
		"u.id",
		"u.name",
		"u.gender",
		"u.age",
		db.Raw(
			`(u.location::geography <-> ST_SetSRID(ST_MakePoint(?,?),?)::geography) / 1000 AS distance`,
			in.Location.Lon(),
			in.Location.Lat(),
			srid,
		),
	}

	order := muzz.SortTypeDistance.String()
	if in.SortType == muzz.SortTypeAttractiveness {
		order = fmt.Sprintf("-%s", muzz.SortTypeAttractiveness.String())
		columns = append(
			columns,
			db.Raw(
				`NULLIF((SELECT COUNT(swiped_user_id) FROM swipe WHERE 
swiped_user_id = u.id AND preference = true),0)::float / 
(SELECT COUNT(swiped_user_id) FROM swipe WHERE swiped_user_id = u.id)::float AS attractiveness`,
			),
		)
	}

	selector := r.Select(columns...).
		From(fmt.Sprintf("%s %s", userTable, "u")).
		Where("u.id != ?", in.UserID).
		And("u.id NOT IN ?", db.Raw(`(SELECT swiped_user_id FROM swipe WHERE user_id = ?)`, in.UserID)).
		OrderBy(order)

	selector = applyUserFilters(in.Filters, selector)

	query := selector.String()
	fmt.Println(query)

	rows, err := selector.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]*muzz.UserDetails, 0, 2)
	for rows.Next() {
		user := new(muzz.UserDetails)
		var genderStr string
		dest := []any{&user.ID, &user.Name, &genderStr, &user.Age, &user.DistanceFromMe}
		if in.SortType == muzz.SortTypeAttractiveness {
			// We need to return this column as we use it to sort, but we have no business case for it
			pseudoFloat := sql.NullFloat64{}
			dest = append(dest, &pseudoFloat)
		}
		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		gender, ok := muzz.GenderValues[genderStr]
		if !ok {
			logger.Warn(ctx, "unsupported gender retrieved from database", zap.String("gender", gender.String()))
		}
		user.Gender = gender

		users = append(users, user)
	}

	if errors.Is(rows.Err(), sql.ErrNoRows) {
		return nil, apperror.NoResults
	}
	if rows.Err() != nil {
		return nil, rows.Err()
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

func applyUserFilters(in *muzz.UserFilters, selector db.Selector) db.Selector {
	if in == nil {
		return selector
	}

	if in.MinAge != 0 {
		selector = selector.And("u.age >= ?", in.MinAge)
	}

	if in.MaxAge != 0 {
		selector = selector.And("u.age <= ?", in.MaxAge)
	}

	if len(in.Genders) != 0 {
		genderStrings := make([]string, len(in.Genders))
		for i, gen := range in.Genders {
			genderStrings[i] = gen.String()
		}
		selector = selector.And("u.gender IN ", genderStrings)
	}

	return selector
}

func pointValue(p orb.Point) *db.RawExpr {
	return db.Raw("ST_SetSRID(ST_MakePoint(?,?),?)", p.Lon(), p.Lat(), srid)
}
