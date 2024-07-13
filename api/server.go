package api

import (
	"errors"
	"net/http"

	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/api/router"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	"github.com/nickbadlose/muzz/internal/cache"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/database/adapter/postgres"
	"github.com/nickbadlose/muzz/internal/location"
	"github.com/nickbadlose/muzz/internal/service"
	"go.opentelemetry.io/otel/trace"
)

// NewServer builds a new *http.Server, configured with the provided database, cache and tracer.
func NewServer(cfg *config.Config, db *database.Database, c *cache.Cache, tp trace.TracerProvider) (*http.Server, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if db == nil {
		return nil, errors.New("database cannot be nil")
	}
	if c == nil {
		return nil, errors.New("cache cannot be nil")
	}
	if tp == nil {
		return nil, errors.New("tracer provider cannot be nil")
	}

	matchAdapter, err := postgres.NewSwipeAdapter(db)
	if err != nil {
		return nil, err
	}
	userAdapter, err := postgres.NewUserAdapter(db)
	if err != nil {
		return nil, err
	}

	authoriser, err := auth.NewAuthoriser(cfg, userAdapter)
	if err != nil {
		return nil, err
	}
	loc, err := location.New(cfg, c)
	if err != nil {
		return nil, err
	}

	authService, err := service.NewAuthService(authoriser, userAdapter)
	if err != nil {
		return nil, err
	}
	matchService, err := service.NewSwipeService(matchAdapter)
	if err != nil {
		return nil, err
	}
	userService, err := service.NewUserService(userAdapter)
	if err != nil {
		return nil, err
	}

	handler, err := handlers.New(cfg, authoriser, loc, authService, userService, matchService)
	if err != nil {
		return nil, err
	}

	routes, err := router.New(cfg, handler, authoriser, tp)
	if err != nil {
		return nil, err
	}

	// TODO server configuration
	return &http.Server{
		Handler: routes,
		Addr:    cfg.Port(),
	}, nil
}
