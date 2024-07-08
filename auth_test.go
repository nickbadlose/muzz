package muzz

import (
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoginInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := &LoginInput{Email: "email", Password: "password", Location: orb.Point{1, 1}}
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
		{
			name:       "invalid latitude",
			input:      &LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{1, 1000}},
			errMessage: "location latitude is out of range",
		},
		{
			name:       "invalid longitude",
			input:      &LoginInput{Email: "test@test.com", Password: "Pa55w0rd!", Location: orb.Point{1000, 1}},
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
