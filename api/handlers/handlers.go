package handlers

import (
	"context"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/paulmach/orb"
)

const (
	renderingErrorMessage = "rendering error response"
)

// TODO move handlers to handlers sub package and service to service subpackage. Both can uses types from here, or do
//  a domain package for types
//  Have a check for nil items in constructors and log.FatalF if they are nil.

// TODO omitempty MatchID in SwipeResponse

// TODO restrict handlers that need authorizer? Pass in using handler adapter pattern.
//  No need for service interfaces in here. If the service needs editing,
//  so does this probably as it will be a business decision

type Authorizer interface {
	// UserFromContext gets the authenticated user from context.
	UserFromContext(ctx context.Context) (userID int, err error)
}

type Locationer interface {
	ByIP(ctx context.Context, sourceIP string) (orb.Point, error)
}

type Handlers struct {
	authorizer   Authorizer
	location     Locationer
	authService  *service.AuthService
	userService  *service.UserService
	matchService *service.MatchService
}

func New(auth Authorizer, l Locationer, as *service.AuthService, us *service.UserService, ms *service.MatchService) *Handlers {
	return &Handlers{auth, l, as, us, ms}
}
