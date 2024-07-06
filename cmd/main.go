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
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/database/adapter/postgres"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/internal/service"
)

const (
	// the timeout for the server to be idle before forcing a shutdown whilst attempting a graceful shutdown.
	idleTimeout = 30 * time.Second
)

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

	db, err := database.New(ctx, &database.Config{
		Username:     cfg.DatabaseUser(),
		Password:     cfg.DatabasePassword(),
		Name:         cfg.Database(),
		Host:         cfg.DatabaseHost(),
		DebugEnabled: false,
	})
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	matchAdapter := postgres.NewMatchAdapter(db)
	userAdapter := postgres.NewUserAdapter(db)

	authorizer := auth.NewAuthorizer(cfg)

	authService := service.NewAuthService(userAdapter, authorizer)
	matchService := service.NewMatchService(matchAdapter)
	userService := service.NewUserService(userAdapter)

	hlr := handlers.New(authorizer, authService, userService, matchService)

	// TODO server configuration
	server := &http.Server{
		Handler: router.New(hlr, authorizer),
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
	err = db.Close()
	if err != nil {
		log.Fatalf("closing database: %s", err)
	}
	err = logger.Close()
	if err != nil {
		log.Fatalf("closing logger: %s", err)
	}
}
