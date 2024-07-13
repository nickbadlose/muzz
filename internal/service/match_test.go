package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	mockservice "github.com/nickbadlose/muzz/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMatchService_Swipe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mockservice.NewMatchRepository(t)
		sut, err := NewSwipeService(m)
		require.NoError(t, err)

		m.EXPECT().
			CreateSwipe(mock.Anything, &muzz.CreateSwipeInput{
				UserID:       1,
				SwipedUserID: 2,
				Preference:   true,
			}).Once().Return(
			&muzz.Match{
				ID:            1,
				MatchedUserID: 2,
				Matched:       true,
			}, nil,
		)

		got, err := sut.Swipe(context.Background(), &muzz.CreateSwipeInput{
			UserID:       1,
			SwipedUserID: 2,
			Preference:   true,
		})
		require.Nil(t, err)
		require.Equal(t, &muzz.Match{
			ID:            1,
			MatchedUserID: 2,
			Matched:       true,
		}, got)
	})

	errCases := []struct {
		name          string
		input         *muzz.CreateSwipeInput
		setupMockRepo func(*mockservice.MatchRepository)
		errMessage    string
		errStatus     apperror.Status
	}{
		{
			name:          "invalid input",
			input:         &muzz.CreateSwipeInput{},
			setupMockRepo: func(m *mockservice.MatchRepository) {},
			errMessage:    "user id is a required field",
			errStatus:     apperror.StatusBadInput,
		},
		{
			name:  "error from repository",
			input: &muzz.CreateSwipeInput{UserID: 1, SwipedUserID: 2, Preference: true},
			setupMockRepo: func(m *mockservice.MatchRepository) {
				m.EXPECT().CreateSwipe(mock.Anything, &muzz.CreateSwipeInput{
					UserID:       1,
					SwipedUserID: 2,
					Preference:   true,
				}).
					Once().Return(nil, errors.New("database error"))
			},
			errMessage: "database error",
			errStatus:  apperror.StatusInternal,
		},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockservice.NewMatchRepository(t)
			tc.setupMockRepo(m)
			sut, err := NewSwipeService(m)
			require.NoError(t, err)

			got, aErr := sut.Swipe(context.Background(), tc.input)
			require.Empty(t, got)
			require.Contains(t, aErr.Error(), tc.errMessage)
			require.Equal(t, aErr.Status(), tc.errStatus)
		})
	}
}
