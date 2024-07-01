package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-chi/render"
)

// TODO
//  - Document how we would break app into separate sections as it grows, user section, with user subrouter and handlers, then eventually it's own microservice

// user field validations
const (
	minimumAge            = 18
	minimumPasswordLength = 8
)

// Gender is an enum for a persons gender.
type Gender uint8

const (
	// GenderUnspecified for when no gender is provided. this is an invalid gender.
	GenderUnspecified Gender = iota
	// GenderMale represents the "male" gender option.
	GenderMale
	// GenderFemale represents the female gender option.
	GenderFemale
)

var (
	// GenderValues maps the gender string values to the enum values.
	GenderValues = map[string]Gender{
		"male":   GenderMale,
		"female": GenderFemale,
	}
	// GenderNames maps the gender enum values to the string values.
	GenderNames = map[Gender]string{
		GenderMale:   "male",
		GenderFemale: "female",
	}
)

// String returns a lower-case representation of the Gender.
func (g *Gender) String() string {
	switch *g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return fmt.Sprintf("Gender(%d)", g)
	}
}

// Valid validates the Gender against the accepted values.
func (g *Gender) Valid() bool {
	_, ok := GenderNames[*g]
	return ok
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (g *Gender) UnmarshalJSON(bytes []byte) error {
	var gString string
	err := json.Unmarshal(bytes, &gString)
	if err != nil {
		return err
	}

	gen, ok := GenderValues[gString]
	if !ok {
		return fmt.Errorf("unknown gender %s", gString)
	}

	*g = gen
	return nil
}

// MarshalJSON implements the json.Unmarshaler interface.
func (g *Gender) MarshalJSON() ([]byte, error) { return json.Marshal(g.String()) }

// User details.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   Gender `json:"gender"`
	Age      int    `json:"age"`
}

// UserResponse object to send to the client.
type UserResponse struct {
	Result *User `json:"result"`
}

// Render implements the render.Render interface.
func (u *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusCreated)
	return nil
}

// CreateUserRequest holds the accepted request format to create a user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   Gender `json:"gender"`
	Age      int    `json:"age"`
}

func (req *CreateUserRequest) validate() error {
	if req.Email == "" {
		return errors.New("email is a required field")
	}

	err := validateEmail(req.Email)
	if err != nil {
		return err
	}

	if req.Password == "" {
		return errors.New("password is a required field")
	}

	err = validatePassword(req.Password)
	if err != nil {
		return err
	}

	if req.Name == "" {
		return errors.New("name is a required field")
	}

	if !req.Gender.Valid() {
		genders := make([]string, 0, len(GenderValues))
		for name := range GenderValues {
			genders = append(genders, name)
		}
		return fmt.Errorf("please provide a valid gender from: %s", strings.Join(genders, ", "))
	}

	if req.Age < minimumAge {
		return errors.New("the minimum age is 18")
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
