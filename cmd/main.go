package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/api/router"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	internalcache "github.com/nickbadlose/muzz/internal/cache"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/database/adapter/postgres"
	"github.com/nickbadlose/muzz/internal/location"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/nickbadlose/muzz/internal/tracer"
)

const (
	// the timeout for the server to be idle before forcing a shutdown whilst attempting a graceful shutdown.
	idleTimeout = 30 * time.Second
)

// TODO
//  NewServer func which constructs server and abstracts it from here
//  integration tests will need current setup due to location mocking
//  Build cache, logger and db etc in here and pass into new server func which builds the rest from them

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	err := logger.New(logger.WithLogLevelString(cfg.LogLevel()))
	if err != nil {
		log.Fatalf("failed to initialise logger: %s", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	tp, err := tracer.New(ctx, cfg, "muzz")
	if err != nil {
		log.Fatalf("failed to initialise tracer: %s", err)
	}

	db, err := database.New(
		ctx,
		&database.Credentials{
			Username: cfg.DatabaseUser(),
			Password: cfg.DatabasePassword(),
			Name:     cfg.Database(),
			Host:     cfg.DatabaseHost(),
		},
		database.WithDebugMode(cfg.DebugEnabled()),
		database.WithTraceProvider(tp),
	)
	if err != nil {
		log.Fatalf("failed to initialise database: %s", err)
	}

	cache, err := internalcache.New(
		ctx,
		&internalcache.Credentials{
			Host:     cfg.CacheHost(),
			Password: cfg.CachePassword(),
		},
		internalcache.WithDebugMode(cfg.DebugEnabled()),
		internalcache.WithTraceProvider(tp),
	)
	if err != nil {
		log.Fatalf("failed to initialise cache: %s", err)
	}

	matchAdapter, err := postgres.NewMatchAdapter(db)
	if err != nil {
		log.Fatalf("failed to initialise match adapter: %s", err)
	}
	userAdapter, err := postgres.NewUserAdapter(db)
	if err != nil {
		log.Fatalf("failed to initialise user adapter: %s", err)
	}

	authorizer, err := auth.NewAuthoriser(cfg, userAdapter)
	if err != nil {
		log.Fatalf("failed to initialise authorizer: %s", err)
	}
	loc, err := location.New(cfg, cache)
	if err != nil {
		log.Fatalf("failed to initialise location: %s", err)
	}

	authService, err := service.NewAuthService(authorizer, userAdapter)
	if err != nil {
		log.Fatalf("failed to initialise auth service: %s", err)
	}
	matchService, err := service.NewMatchService(matchAdapter)
	if err != nil {
		log.Fatalf("failed to initialise match service: %s", err)
	}
	userService, err := service.NewUserService(userAdapter)
	if err != nil {
		log.Fatalf("failed to initialise user service: %s", err)
	}

	hlr, err := handlers.New(cfg, authorizer, loc, authService, userService, matchService)
	if err != nil {
		log.Fatalf("failed to initialise handlers: %s", err)
	}

	// TODO server configuration
	server := &http.Server{
		Handler: router.New(hlr, cfg, authorizer, tp),
		Addr:    cfg.Port(),
	}

	go func() {
		log.Printf("listening on port: %v\n", cfg.Port())
		sErr := server.ListenAndServe()
		if sErr != nil {
			log.Fatalf("starting server: %s", sErr)
		}
	}()

	<-sig
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, idleTimeout)
	defer timeoutCancel()
	// TODO close all service connections after graceful shutdown, logger last
	err = server.Shutdown(timeoutCtx)
	if err != nil {
		log.Fatalf("shutting down server: %s", err)
	}
	err = cache.Close()
	if err != nil {
		log.Fatalf("closing cache: %s", err)
	}
	err = db.Close()
	if err != nil {
		log.Fatalf("closing database: %s", err)
	}
	err = logger.Close()
	if err != nil {
		log.Fatalf("closing logger: %s", err)
	}
}
