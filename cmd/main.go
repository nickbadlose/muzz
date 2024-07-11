package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	err := logger.New(logger.WithLogLevelString(cfg.LogLevel()))
	if err != nil {
		log.Fatalf("failed to initialize logger: %s", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	tp, err := tracer.New(ctx, cfg, "muzz")
	if err != nil {
		log.Fatalf("failed to initialize tracer: %s", err)
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
		log.Fatalf("failed to initialize database: %s", err)
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
		log.Fatalf("failed to initialize cache: %s", err)
	}

	matchAdapter := postgres.NewMatchAdapter(db)
	userAdapter := postgres.NewUserAdapter(db)

	authorizer := auth.NewAuthorizer(cfg)
	loc := location.New(cfg, cache)

	authService := service.NewAuthService(userAdapter, authorizer)
	matchService := service.NewMatchService(matchAdapter)
	userService := service.NewUserService(userAdapter)

	hlr := handlers.New(cfg, authorizer, loc, authService, userService, matchService)

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
