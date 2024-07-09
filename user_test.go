package muzz

import (
	"github.com/paulmach/orb"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUserInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := &CreateUserInput{
			Email:    "test@test.com",
			Password: "Pa55w0rd!",
			Name:     "test",
			Gender:   "male",
			Age:      25,
			Location: orb.Point{1, 1},
		}

		require.NoError(t, input.Validate())
	})

	errCases := []struct {
		name       string
		input      *CreateUserInput
		errMessage string
	}{
		{
			name:       "missing email",
			input:      &CreateUserInput{},
			errMessage: "email is a required field",
		},
		{
			name:       "invalid email: contains spaces",
			input:      &CreateUserInput{Email: "te st@test.com"},
			errMessage: "email cannot contain spaces",
		},
		{
			name:       "invalid email: no @",
			input:      &CreateUserInput{Email: "invalidEmail"},
			errMessage: "mail: missing '@' or angle-addr",
		},
		{
			name:       "invalid email: no .",
			input:      &CreateUserInput{Email: "test@test"},
			errMessage: "invalid email address: missing '.' in email domain",
		},
		{
			name:       "missing password",
			input:      &CreateUserInput{Email: "test@test.com"},
			errMessage: "password is a required field",
		},
		{
			name:       "invalid password: no upper case",
			input:      &CreateUserInput{Email: "test@test.com", Password: "passw0rd!"},
			errMessage: "password must contain at least 1 uppercase letter",
		},
		{
			name:       "invalid password: no lower case",
			input:      &CreateUserInput{Email: "test@test.com", Password: "PASSW0RD!"},
			errMessage: "password must contain at least 1 lowercase letter",
		},
		{
			name:       "invalid password: no special character",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Passw0rd"},
			errMessage: "password must contain at least 1 special character",
		},
		{
			name:       "invalid password: no numbers",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Password!"},
			errMessage: "password must contain at least 1 number",
		},
		{
			name:       "invalid password: not 8 characters",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55!"},
			errMessage: "password must contain at least 8 characters",
		},
		{
			name:       "missing name",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!"},
			errMessage: "name is a required field",
		},
		{
			name:       "missing name",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!"},
			errMessage: "name is a required field",
		},
		{
			name:       "missing gender",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test"},
			errMessage: "invalid gender, valid values are:",
		},
		{
			name:       "invalid gender: out of range int",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "not a valid gender"},
			errMessage: "invalid gender, valid values are:",
		},
		{
			name:       "missing age",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "male", Age: 0},
			errMessage: "the minimum age is 18",
		},
		{
			name:       "invalid age: too low",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "male", Age: 17},
			errMessage: "the minimum age is 18",
		},
		{
			name:       "invalid longitude",
			input:      &CreateUserInput{Email: "test@test.com", Password: "Pa55w0rd!", Name: "Test", Gender: "male", Age: 18, Location: orb.Point{1000, 1}},
			errMessage: "location longitude is out of range",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errMessage)
		})
	}
}

func TestGender_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, GenderUnspecified.Validate())
		require.NoError(t, GenderMale.Validate())
		require.NoError(t, GenderFemale.Validate())
	})

	t.Run("invalid", func(t *testing.T) {
		gender := Gender(0)
		require.Error(t, gender.Validate())
		require.Contains(t, gender.Validate().Error(), "invalid gender, valid values are:")
		require.Error(t, Gender(100).Validate())
	})
}

func TestGender_String(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.Equal(t, "unspecified", GenderUnspecified.String())
		require.Equal(t, "male", GenderMale.String())
		require.Equal(t, "female", GenderFemale.String())
	})

	t.Run("invalid", func(t *testing.T) {
		require.Equal(t, "Gender(0)", GenderUndefined.String())
		require.Equal(t, "Gender(100)", Gender(100).String())
	})
}

func TestUserFiltersFromParams(t *testing.T) {
	validCases := []struct {
		name                    string
		maxAge, minAge, genders string
		out                     *UserFilters
	}{
		{
			name:    "valid no params provided",
			maxAge:  "",
			minAge:  "",
			genders: "",
			out:     &UserFilters{},
		},
		{
			name:    "valid all params provided",
			maxAge:  "30",
			minAge:  "20",
			genders: "male,female,unspecified",
			out:     &UserFilters{MaxAge: 30, MinAge: 20, Genders: []Gender{GenderMale, GenderFemale, GenderUnspecified}},
		},
		{
			name:    "valid with invalid genders",
			genders: "invalidGender,another invalid gender",
			out:     &UserFilters{MaxAge: 0, MinAge: 0, Genders: []Gender{GenderUndefined, GenderUndefined}},
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := UserFiltersFromParams(tc.maxAge, tc.minAge, tc.genders)
			require.NoError(t, err)
			require.Equal(t, got, tc.out)
		})
	}

	errCases := []struct {
		name                    string
		maxAge, minAge, genders string
		errMessage              string
	}{
		{
			name:       "invalid max age",
			maxAge:     "not an int",
			errMessage: "max age must be an integer: strconv.Atoi: parsing \"not an int\": invalid syntax",
		},
		{
			name:       "invalid min age",
			minAge:     "not an int",
			errMessage: "min age must be an integer: strconv.Atoi: parsing \"not an int\": invalid syntax",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := UserFiltersFromParams(tc.maxAge, tc.minAge, tc.genders)
			require.Nil(t, got)
			require.Error(t, err)
			require.Equal(t, tc.errMessage, err.Error())
		})
	}
}

func TestGetUsersInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := &GetUsersInput{
			UserID:  1,
			Filters: nil,
		}

		require.NoError(t, input.Validate())
	})

	validCases := []struct {
		name  string
		input *GetUsersInput
	}{
		{
			name: "valid with filters",
			input: &GetUsersInput{
				UserID: 1,
				Filters: &UserFilters{
					MaxAge:  30,
					MinAge:  20,
					Genders: []Gender{GenderMale, GenderFemale, GenderUnspecified},
				},
			},
		},
		{
			name: "valid without filters",
			input: &GetUsersInput{
				UserID:  1,
				Filters: nil,
			},
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.input.Validate())
		})
	}

	errCases := []struct {
		name       string
		input      *GetUsersInput
		errMessage string
	}{
		{
			name: "invalid no user id",
			input: &GetUsersInput{
				UserID:  0,
				Filters: nil,
			},
			errMessage: "user id is a required field",
		},
		{
			name: "invalid filters, max age less than 18",
			input: &GetUsersInput{
				UserID: 1,
				Filters: &UserFilters{
					MaxAge:  10,
					MinAge:  0,
					Genders: nil,
				},
			},
			errMessage: "max age cannot be less than 18",
		},
		{
			name: "invalid filters, min age less than 18",
			input: &GetUsersInput{
				UserID: 1,
				Filters: &UserFilters{
					MaxAge:  30,
					MinAge:  10,
					Genders: nil,
				},
			},
			errMessage: "min age cannot be less than 18",
		},
		{
			name: "invalid filters, max age less than min age",
			input: &GetUsersInput{
				UserID: 1,
				Filters: &UserFilters{
					MaxAge:  30,
					MinAge:  40,
					Genders: nil,
				},
			},
			errMessage: "max age cannot be less than min age",
		},
		{
			name: "invalid filters, invalid genders",
			input: &GetUsersInput{
				UserID: 1,
				Filters: &UserFilters{
					MaxAge:  30,
					MinAge:  20,
					Genders: []Gender{GenderUndefined},
				},
			},
			errMessage: "invalid gender, valid values are:",
		},
		{
			name: "invalid latitude",
			input: &GetUsersInput{
				UserID:   1,
				Location: orb.Point{1, 1000},
			},
			errMessage: "location latitude is out of range",
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errMessage)
		})
	}
}

func TestSortType_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, SortTypeDistance.Validate())
		require.NoError(t, SortTypeAttractiveness.Validate())
	})

	t.Run("invalid", func(t *testing.T) {
		gender := SortType(100)
		require.Error(t, gender.Validate())
		require.Contains(t, gender.Validate().Error(), "invalid sort type, valid values are:")
	})
}

func TestSortType_String(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.Equal(t, "distance", SortTypeDistance.String())
		require.Equal(t, "attractiveness", SortTypeAttractiveness.String())
	})

	t.Run("invalid", func(t *testing.T) {
		require.Equal(t, "SortType(100)", SortType(100).String())
	})
}
