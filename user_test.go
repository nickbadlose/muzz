package muzz

import (
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
