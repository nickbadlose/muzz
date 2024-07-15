package main

import (
	"context"
	"github.com/nickbadlose/muzz/api"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nickbadlose/muzz/config"
	internalcache "github.com/nickbadlose/muzz/internal/cache"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/nickbadlose/muzz/internal/tracer"
)

const (
	// the timeout for the server to be idle before forcing a shutdown whilst attempting a graceful shutdown.
	idleTimeout     = 30 * time.Second
	applicationName = "muzz"
)

func main() {
	ctx := context.Background()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	cfg := config.MustLoad()

	err := logger.New(logger.WithLogLevelString(cfg.LogLevel()))
	if err != nil {
		log.Fatalf("failed to initialise logger: %s", err)
	}

	tp, err := tracer.New(cfg, applicationName)
	if err != nil {
		logger.Fatal(ctx, "failed to initialise tracer", zap.Error(err))
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
		logger.Fatal(ctx, "failed to initialise database", zap.Error(err))
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
		logger.Fatal(ctx, "failed to initialise cache", zap.Error(err))
	}

	srv, err := api.NewServer(cfg, db, cache, tp)
	if err != nil {
		logger.Fatal(ctx, "failed to initialise server", zap.Error(err))
	}

	go func() {
		log.Printf("listening on port: %v\n", cfg.Port())
		sErr := srv.ListenAndServe()
		if sErr != nil {
			logger.Error(ctx, "starting server", sErr)
		}
	}()

	<-sig

	ctx, cancel := context.WithTimeout(ctx, idleTimeout)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error(ctx, "shutting down server", err)
	}
	err = cache.Close()
	if err != nil {
		logger.Error(ctx, "closing cache", err)
	}
	err = db.Close()
	if err != nil {
		logger.Error(ctx, "closing database", err)
	}
	err = logger.Close()
	if err != nil {
		log.Printf("closing logger: %s\n", err)
	}
}
