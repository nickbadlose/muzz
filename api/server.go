package api

import (
	"errors"
	"net/http"
	"time"

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

const (
	// timeouts.
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 10 * time.Second
	readTimeout       = readHeaderTimeout + writeTimeout
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

	r, err := router.New(cfg, handler, authoriser, tp)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.Port(),
		Handler:           r,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}, nil
}
