package main

import (
	"context"
	"github.com/nickbadlose/muzz/internal/http/handlers"
	"github.com/nickbadlose/muzz/internal/http/router"
	"github.com/nickbadlose/muzz/internal/pkg/auth"
	"github.com/nickbadlose/muzz/internal/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/app"
	"github.com/nickbadlose/muzz/internal/pkg/database"
	"github.com/nickbadlose/muzz/internal/pkg/logger"
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

	str := store.New(db)
	au := auth.NewAuthoriser(cfg)
	svc := app.NewService(str, au)
	hlr := handlers.NewHandlers(svc, au)

	// TODO server configuration
	server := &http.Server{
		Handler: router.New(hlr, au),
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
