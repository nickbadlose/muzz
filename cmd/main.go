package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/app"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/router"
)

const (
	// the timeout for the server to be idle before forcing a shutdown whilst attempting a graceful shutdown.
	idleTimeout = 30 * time.Second
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("loading .env file: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	err = logger.New(logger.WithLogLevelString(cfg.LogLevel))
	if err != nil {
		log.Fatalf("failed to initialize logger: %s", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// TODO get from env
	db, err := database.New(ctx, &database.Config{
		Username:     cfg.DatabaseUser,
		Password:     cfg.DatabasePassword,
		Name:         cfg.Database,
		Host:         cfg.DatabaseHost,
		DebugEnabled: false,
	})
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	service := app.NewService(db)
	handlers := app.NewHandlers(service)

	// TODO server configuration
	server := &http.Server{
		Handler: router.New(handlers),
		Addr:    cfg.Port,
	}

	go func() {
		log.Printf("listening on port: %v\n", cfg.Port)
		sErr := server.ListenAndServe()
		if sErr != nil {
			log.Fatalf("starting server: %s", sErr)
		}
	}()

	<-sig
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, idleTimeout)
	defer timeoutCancel()
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
