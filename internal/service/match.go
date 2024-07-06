package service

import (
	"context"

	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.uber.org/zap"
)

type MatchRepository interface {
	CreateSwipe(context.Context, *muzz.CreateSwipeInput) (*muzz.Match, error)
}

type MatchService struct {
	repository MatchRepository
}

func NewMatchService(mr MatchRepository) *MatchService {
	return &MatchService{mr}
}

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
