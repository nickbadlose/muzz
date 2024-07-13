package service

import (
	"context"
	"errors"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

// MatchRepository is the interface to write and read match/swipe data from the repository.
type MatchRepository interface {
	// CreateSwipe adds a swipe record to the repository and if appropriate, adds corresponding match records too.
	CreateSwipe(context.Context, *muzz.CreateSwipeInput) (*muzz.Match, error)
}

// MatchService is the service which handles all match and swipe based requests.
type MatchService struct {
	repository MatchRepository
}

// NewMatchService builds a new *MatchService.
func NewMatchService(mr MatchRepository) (*MatchService, error) {
	if mr == nil {
		return nil, errors.New("match repository cannot be nil")
	}
	return &MatchService{repository: mr}, nil
}

// Swipe performs a swipe action against a given user. There are 4 possible outcomes from this action:
//   - The authenticated user swipes with a preference of false, in which case, no match record is ever created.
//   - The authenticated user swipes with a preference of true, the swiped user has already swiped with a preference of
//     false, no match record will ever be created.
//   - The authenticated user swipes with a preference of true, the swiped user has not yet performed a swipe action
//     against the authenticated user, no match record will be created, however one may be in the future.
//   - The authenticated user swipes with a preference of true, the swiped user has already swiped with a preference of
//     true, two match records will be created, one for each user, and one will be returned from the request.
//
// If there was no match, the returned muzz.Match will have Matched = false and ID = 0.
func (ms *MatchService) Swipe(ctx context.Context, in *muzz.CreateSwipeInput) (*muzz.Match, *apperror.Error) {
	logger.Debug(ctx, "MatchService Swipe", zap.Any("request", in))

	err := in.Validate()
	if err != nil {
		logger.Error(ctx, "validating create swipe input request", err)
		return nil, apperror.BadInput(err)
	}

	match, err := ms.repository.CreateSwipe(ctx, in)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return match, nil
}
