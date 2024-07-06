package muzz

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoginInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := &LoginInput{"email", "password"}
		require.NoError(t, in.Validate())
	})

	errCases := []struct {
		name       string
		input      *LoginInput
		errMessage string
	}{
		{
			name:       "missing email",
			input:      &LoginInput{},
			errMessage: "email is a required field",
		},
		{
			name:       "missing password",
			input:      &LoginInput{Email: "test@test.com"},
			errMessage: "password is a required field",
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
