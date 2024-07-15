package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
	"github.com/upper/db/v4"
	"go.uber.org/zap"
)

const (
	userTable = "user"
	// nolint:revive // this is a URL.
	// spatial reference identifier, https://postgis.net/docs/manual-3.4/using_postgis_dbmanagement.html#spatial_ref_sys:~:text=The%20columns%20are%3A-,srid,-An%20integer%20code
	srid         = 4326
	defaultLimit = 50
)

// UserAdapter adapts a *database.Database to the service.UserRepository interface.
type UserAdapter struct {
	database *database.Database
}

// NewUserAdapter builds a new *UserAdapter.
func NewUserAdapter(d *database.Database) (*UserAdapter, error) {
	if d == nil {
		return nil, errors.New("database cannot be nil")
	}
	return &UserAdapter{database: d}, nil
}

// userEntity represents a row in the user table.
type userEntity struct {
	id       int
	email    string
	password string
	name     string
	gender   string
	age      int
	location orb.Point
}

// CreateUser adds a user record to the user table.
//
// It encrypts the password and manipulates the email to remove whitespace and force lowercase.
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

// Authenticate attempts to retrieve the a user record by matching with the provided email and password.
//
// It uses pgcrypto to compare the provided password with the stored encrypted one.
func (ua *UserAdapter) Authenticate(ctx context.Context, email, password string) (*muzz.User, error) {
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []any{"id", "email", "password", "name", "gender", "age"}
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
		return nil, apperror.ErrNoResults
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

// GetUsers retrieves a list of filtered user records from the table sorted by the provided sort type. A limit of 50
// users are returned per request.
func (ua *UserAdapter) GetUsers(ctx context.Context, in *muzz.GetUsersInput) ([]*muzz.UserDetails, error) {
	s, err := ua.database.SQLSessionContext(ctx)
	if err != nil {
		return nil, err
	}

	columns := []any{
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
		return nil, apperror.ErrNoResults
	}

	return users, nil
}

// UpdateUserLocation updates a user records location field with the provided data.
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
		genderStrings := make([]string, 0, len(in.Genders))
		for _, gen := range in.Genders {
			genderStrings = append(genderStrings, gen.String())
		}
		selector = selector.And("u.gender IN ", genderStrings)
	}

	return selector
}

// pointValue creates a postgres Point type from the given longitude and latitude points using srid = 4326.
func pointValue(p orb.Point) *db.RawExpr {
	return db.Raw("ST_SetSRID(ST_MakePoint(?,?),?)", p.Lon(), p.Lat(), srid)
}
