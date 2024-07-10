package apperror

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestError_ToHTTP(t *testing.T) {
	cases := []struct {
		name           string
		status         Status
		expectedStatus int
	}{
		{
			name:           "internal server error",
			status:         StatusInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "bad request",
			status:         StatusBadInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			status:         StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			status:         StatusUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewErr(tc.status, errors.New("test error"))

			httpErr := err.ToHTTP()
			require.Equal(t, tc.expectedStatus, httpErr.Status)
		})
	}
}

func TestError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := NewErr(StatusInternal, errors.New("test error"))
		require.Equal(t, "test error", err.Error())
	})

	t.Run("Error: nil error", func(t *testing.T) {
		err := NewErr(StatusInternal, nil)
		require.Equal(t, "", err.Error())
	})

	t.Run("Status", func(t *testing.T) {
		err := NewErr(StatusInternal, errors.New("test error"))
		require.Equal(t, StatusInternal, err.Status())
	})
}
