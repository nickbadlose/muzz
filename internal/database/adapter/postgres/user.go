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
	"strings"
)

// TODO
//  Strings.ToLower emails wherever set and read

const (
	userTable    = "user"
	srid         = 4326
	defaultLimit = 50
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
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []string{"id", "email", "password", "name", "gender", "age", "location"}
	row, err := s.InsertInto(userTable).
		Columns(columns[1:]...).
		Values(
			strings.TrimSpace(strings.ToLower(in.Email)),
			db.Raw(`crypt(?, gen_salt('bf'))`, in.Password),
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

func (ua *UserAdapter) Authenticate(ctx context.Context, email, password string) (*muzz.User, error) {
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO index searching by email
	columns := []interface{}{"id", "email", "password", "name", "gender", "age"}
	row, err := s.Select(columns...).
		From(userTable).
		Where("email = ?", email).
		And("password = crypt(?, password)", password).
		QueryRowContext(ctx)
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
//  Parent fn can just create SQLSessionContext or WriteSessionContext or tx and pass in.
//  Exclude already swiped users from results.
//  Have tie breaker order by column? ID check if that's default anyway
//  index for swiped user_id (SELECT swiped_user_id FROM swipe WHERE user_id = ?) seed data and analyze before and after indexes for distance, swiped user etc.
//  Check all make functions and use correct methods ie len or cap with append or [i]
//  Do we need entities, we are just converting to another object immediately

func (ua *UserAdapter) GetUsers(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, error) {
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []interface{}{
		"u.id",
		"u.name",
		"u.gender",
		"u.age",
		// "/ 1000" converts from metres to km.
		db.Raw(
			`(u.location <-> ST_SetSRID(ST_MakePoint(?,?),?)) / 1000 AS distance`,
			in.Location.Lon(),
			in.Location.Lat(),
			srid,
		),
	}

	selector := s.Select(columns...).
		From(fmt.Sprintf("%s %s", userTable, "u")).
		// exclude the current user.
		Where("u.id != ?", in.UserID).
		// exclude any users they have already swiped.
		And("u.id NOT IN ?", db.Raw(`(SELECT swiped_user_id FROM swipe WHERE user_id = ?)`, in.UserID))

	var order any = in.SortType.String()
	if in.SortType == muzz.SortTypeAttractiveness {
		// attractiveness sorting algorithm: total_preferred_swipes / total_swipes
		// total_preferred_swipes - swipes on the user where preferred = true
		// total_swipes - total swipes on the user of either preference
		// this gives us an attractiveness percentage, between 0 and 1.
		//
		// for the algorithm we need to count both values, so we join the user and swipe tables
		// on swiped_user_id, so we can count the occurrences of the user.id column as the value
		// for 'total_swipes' and by excluding cases where preference = false from the same count,
		// we get 'total_preferred_swipes'.
		selector = selector.LeftJoin(fmt.Sprintf("%s %s", swipeTable, "s")).
			On("u.id = s.swiped_user_id").
			// group the user rows into one for counting and removing duplicate users.
			GroupBy("u.id")

		order = db.Raw(`(NULLIF(sum(case when s.preference then 1 else 0 end),0)::float / COUNT(u.id)::float) DESC`)
	}

	rows, err := applyUserFilters(in.Filters, selector).
		OrderBy(order).
		Limit(defaultLimit).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		logger.MaybeError(ctx, "closing db rows", rows.Close())
	}()

	users := make([]*muzz.UserDetails, 0, 2)
	for rows.Next() {
		user := new(muzz.UserDetails)
		var genderStr string
		dest := []any{&user.ID, &user.Name, &genderStr, &user.Age, &user.DistanceFromMe}

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

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if len(users) == 0 {
		return nil, apperror.NoResults
	}

	return users, nil
}

func (ua *UserAdapter) UpdateUserLocation(ctx context.Context, id int, location orb.Point) error {
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return err
	}

	_, err = s.Update(userTable).
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
