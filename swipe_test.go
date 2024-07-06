package muzz

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateSwipeInput_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := &CreateSwipeInput{UserID: 1, SwipedUserID: 2, Preference: true}
		require.NoError(t, in.Validate())
	})

	errCases := []struct {
		name       string
		input      *CreateSwipeInput
		errMessage string
	}{
		{
			name:       "missing user id",
			input:      &CreateSwipeInput{},
			errMessage: "user id is a required field",
		},
		{
			name:       "missing swiped user id",
			input:      &CreateSwipeInput{UserID: 1},
			errMessage: "swiped user id is a required field",
		},
		{
			name:       "user id and swiped user id are the same",
			input:      &CreateSwipeInput{UserID: 1, SwipedUserID: 1},
			errMessage: "user id and swiped user id cannot be the same value",
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
