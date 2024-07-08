package muzz

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/pkg/errors"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// TODO
//  - Document how we would break app into separate sections as it grows, user section, with user subrouter and handlers, then eventually it's own microservice
//  - tests at this level

// user field validations
const (
	minimumAge            = 18
	minimumPasswordLength = 8
	minimumLat            = -90
	maximumLat            = 90
	minimumLon            = -180
	maximumLon            = 180
)

// User contains all a users stored details.
type User struct {
	ID       int
	Email    string
	Password string
	Name     string
	Gender   Gender
	Age      int
	Location orb.Point
}

// UserDetails contains only public user details.
type UserDetails struct {
	ID             int
	Name           string
	Gender         Gender
	Age            int
	DistanceFromMe float64
}

// CreateUserInput is the accepted request to create a user.
type CreateUserInput struct {
	Email    string
	Password string
	Name     string
	Gender   string
	Age      int
	Location orb.Point
}

// Validate the CreateUserInput fields.
func (in *CreateUserInput) Validate() error {
	if in.Email == "" {
		return errors.New("email is a required field")
	}

	err := validateEmail(in.Email)
	if err != nil {
		return err
	}

	if in.Password == "" {
		return errors.New("password is a required field")
	}

	err = validatePassword(in.Password)
	if err != nil {
		return err
	}

	if in.Name == "" {
		return errors.New("name is a required field")
	}

	gender := GenderValues[in.Gender]
	err = gender.Validate()
	if err != nil {
		return err
	}

	if in.Age < minimumAge {
		return errors.New("the minimum age is 18")
	}

	err = validatePoint(in.Location)
	if err != nil {
		return err
	}

	return nil
}

// Gender is an enum for a persons gender.
type Gender uint8

const (
	// GenderUndefined for when no gender is provided, this is an invalid gender.
	GenderUndefined Gender = iota
	// GenderUnspecified is for when a user wishes not to state their.
	GenderUnspecified
	// GenderMale represents the "male" gender option.
	GenderMale
	// GenderFemale represents the "female" gender option.
	GenderFemale
)

var (
	// GenderValues maps the gender string values to the enum values.
	GenderValues = map[string]Gender{
		"unspecified": GenderUnspecified,
		"male":        GenderMale,
		"female":      GenderFemale,
	}
	// GenderNames maps the gender enum values to the string values.
	GenderNames = map[Gender]string{
		GenderUnspecified: "unspecified",
		GenderMale:        "male",
		GenderFemale:      "female",
	}
)

// String returns a lower-case representation of the Gender.
func (g Gender) String() string {
	gender, ok := GenderNames[g]
	if !ok {
		return fmt.Sprintf("Gender(%d)", g)
	}

	return gender
}

// Validate the Gender against the accepted values.
func (g Gender) Validate() error {
	_, ok := GenderNames[g]
	if !ok {
		genders := make([]string, 0, len(GenderValues))
		for name := range GenderValues {
			genders = append(genders, name)
		}
		return fmt.Errorf("invalid gender, valid values are: %s", strings.Join(genders, ", "))
	}

	return nil
}

func validatePassword(pass string) error {
	var number, upper, lower, special bool
	for _, c := range pass {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c):
			lower = true
		default:
			return fmt.Errorf("invalid character: %c", c)
		}
	}

	if !number {
		return errors.New("password must contain at least 1 number")
	}
	if !upper {
		return errors.New("password must contain at least 1 uppercase letter")
	}
	if !lower {
		return errors.New("password must contain at least 1 lowercase letter")
	}
	if !special {
		return errors.New("password must contain at least 1 special character")
	}
	if len(pass) < minimumPasswordLength {
		return fmt.Errorf("password must contain at least %d characters", minimumPasswordLength)
	}

	return nil
}

func validateEmail(email string) error {
	rgx, err := regexp.Compile(`\s`)
	if err != nil {
		return err
	}
	if rgx.MatchString(email) {
		return errors.New("email cannot contain spaces")
	}

	em, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}

	if len(strings.Split(em.Address, ".")) < 2 {
		return errors.New("invalid email address: missing '.' in email domain")
	}
	return nil
}

func validatePoint(p orb.Point) error {
	if p.Lat() < minimumLat || p.Lat() > maximumLat {
		return errors.New("location latitude is out of range")
	}

	if p.Lon() < minimumLon || p.Lon() > maximumLon {
		return errors.New("location longitude is out of range")
	}

	return nil
}

type GetUsersInput struct {
	UserID   int
	Location orb.Point
	Filters  *UserFilters
}

func (in *GetUsersInput) Validate() error {
	if in.UserID == 0 {
		return errors.New("user id is a required field")
	}

	err := validatePoint(in.Location)
	if err != nil {
		return err
	}

	// TODO accept nil filter or not?
	if in.Filters != nil {
		return in.Filters.Validate()
	}

	return nil
}

type UserFilters struct {
	MaxAge  int
	MinAge  int
	Genders []Gender
}

func (uf *UserFilters) Validate() error {
	if uf.MaxAge < minimumAge && uf.MaxAge != 0 {
		return errors.New("max age cannot be less than 18")
	}

	if uf.MinAge < minimumAge && uf.MinAge != 0 {
		return errors.New("min age cannot be less than 18")
	}

	if uf.MaxAge < uf.MinAge {
		return errors.New("max age cannot be less than min age")
	}

	for _, g := range uf.Genders {
		err := g.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func UserFiltersFromParams(maxAge, minAge, genderQueryParam string) (*UserFilters, error) {
	var (
		maxAgeInt, minAgeInt int
		genders              []Gender
		err                  error
	)

	if maxAge != "" {
		maxAgeInt, err = strconv.Atoi(maxAge)
		if err != nil {
			return nil, errors.Wrap(err, "max age must be an integer")
		}
	}

	if minAge != "" {
		minAgeInt, err = strconv.Atoi(minAge)
		if err != nil {
			return nil, errors.Wrap(err, "min age must be an integer")
		}
	}

	if genderQueryParam != "" {
		genderStrings := strings.Split(genderQueryParam, ",")
		genders = make([]Gender, len(genderStrings))
		for i, g := range genderStrings {
			genders[i] = GenderValues[g]
		}
	}

	return &UserFilters{
		MaxAge:  maxAgeInt,
		MinAge:  minAgeInt,
		Genders: genders,
	}, nil
}
