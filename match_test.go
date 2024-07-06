package muzz

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateMatchInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := &CreateMatchInput{1, 2}
		require.NoError(t, in.Validate())
	})

	errCases := []struct {
		name       string
		input      *CreateMatchInput
		errMessage string
	}{
		{
			name:       "missing user id",
			input:      &CreateMatchInput{},
			errMessage: "user id is a required field",
		},
		{
			name:       "missing matched user id",
			input:      &CreateMatchInput{UserID: 1},
			errMessage: "matched user id is a required field",
		},
		{
			name:       "user id and matched user id are the same",
			input:      &CreateMatchInput{UserID: 1, MatchedUserID: 1},
			errMessage: "user id and matched user id cannot be the same value",
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
