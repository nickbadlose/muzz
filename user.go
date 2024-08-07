package muzz

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/paulmach/orb"
	"github.com/pkg/errors"
)

const (
	// user field validations.
	minimumAge            = 18
	minimumPasswordLength = 8

	// location validations.
	minimumLat = -90
	maximumLat = 90
	minimumLon = -180
	maximumLon = 180
)

// User contains all a user records stored details.
type User struct {
	// ID is the unique identifier of the user record.
	ID int
	// Email of the user record.
	Email string
	// Password of the user record.
	Password string
	// Name of the user record.
	Name string
	// Gender of the user record.
	Gender Gender
	// Age of the user record.
	Age int
	// Location of the user in longitude and latitude on their last login.
	Location orb.Point
}

// UserDetails contains only public user details of a user record.
type UserDetails struct {
	// ID is the unique identifier of the user record.
	ID int
	// Name of the user record.
	Name string
	// Gender of the user record.
	Gender Gender
	// Age of the user record.
	Age int
	// DistanceFromMe is the calculated distance from the authenticated user making
	// the request at the time of the query.
	DistanceFromMe float64
}

// CreateUserInput to create a user record.
type CreateUserInput struct {
	// Email address of the user.
	Email string
	// Password of the user to authenticate with.
	Password string
	// Name of the user.
	Name string
	// Gender of the user.
	Gender string
	// Age of the user,
	Age int
	// Location of the user in longitude and latitude.
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

// validatePassword returns an error if the given password does not match the following criteria:
//   - Contains at least 1 lowercase letter.
//   - Contains at least 1 uppercase letter.
//   - Contains at least 1 number.
//   - Contains at least 1 special character.
//   - Contains at least minimumPasswordLength characters.
//   - Contains no spaces.
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
	rgx := regexp.MustCompile(`\s`)
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

// SortType is an enum for a sort type.
type SortType uint8

const (
	// SortTypeDistance sorts user records by the distance from the authenticated user.
	SortTypeDistance SortType = iota
	// SortTypeAttractiveness sorts user records by how attractive they are according to swipes.
	SortTypeAttractiveness
)

var (
	// SortValues maps the sort type string values to the enum values.
	SortValues = map[string]SortType{
		"distance":       SortTypeDistance,
		"attractiveness": SortTypeAttractiveness,
	}
	// SortNames maps the sort type enum values to the string values.
	SortNames = map[SortType]string{
		SortTypeDistance:       "distance",
		SortTypeAttractiveness: "attractiveness",
	}
)

// String returns a lower-case representation of the SortType.
func (s SortType) String() string {
	gender, ok := SortNames[s]
	if !ok {
		return fmt.Sprintf("SortType(%d)", s)
	}

	return gender
}

// Validate the SortType against the accepted values.
func (s SortType) Validate() error {
	_, ok := SortNames[s]
	if !ok {
		sortTypes := make([]string, 0, len(SortValues))
		for name := range SortValues {
			sortTypes = append(sortTypes, name)
		}
		return fmt.Errorf("invalid sort type, valid values are: %s", strings.Join(sortTypes, ", "))
	}

	return nil
}

// GetUsersInput to get a list of user records.
type GetUsersInput struct {
	// UserID to filter already swiped users from the returned records.
	UserID int
	// Location of the authenticated user in longitude and latitude,
	// used to calculate the user records distance from the authenticated user.
	Location orb.Point
	// SortType to sort the user records by.
	SortType SortType
	// Filters to filter the user records by.
	Filters *UserFilters
}

// Validate the GetUsersInput fields.
func (in *GetUsersInput) Validate() error {
	if in.UserID == 0 {
		return errors.New("user id is a required field")
	}

	err := validatePoint(in.Location)
	if err != nil {
		return err
	}

	if in.Filters != nil {
		return in.Filters.Validate()
	}

	return nil
}

// UserFilters provided optional fields to filter user records by.
// Zero values mean no filter should be performed for the respective field.
type UserFilters struct {
	// MaxAge of the user records to return, inclusive.
	MaxAge int
	// MinAge of the user records to return, inclusive.
	MinAge int
	// Genders of the user records to return, if empty, all genders are returned.
	Genders []Gender
}

// Validate the UserFilters fields.
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

// UserFiltersFromParams takes generic parameter strings from the request URL and attempts to convert
// them into the correct type for their corresponding UserFilters fields.
//
// It DOES NOT validate the values, if you wish to validate them, a call to UserFilters.Validate() should be made.
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
		genders = make([]Gender, 0, len(genderStrings))
		for _, g := range genderStrings {
			genders = append(genders, GenderValues[g])
		}
	}

	return &UserFilters{
		MaxAge:  maxAgeInt,
		MinAge:  minAgeInt,
		Genders: genders,
	}, nil
}
