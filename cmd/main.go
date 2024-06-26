package main

import (
	"context"
	"fmt"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// cfgDefaultHost is the default host/port for the web app.
	cfgDefaultHost = "3002"
	// cfgHost is the configuration key for the host/port to bind to.
	cfgHost = "host"
	// the timeout for the server to be idle before forcing a shutdown whilst attempting a graceful shutdown.
	idleTimeout = 30 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	server := &http.Server{
		Handler: router.New(),
		Addr:    cfg.Host(),
	}

	go func() {
		fmt.Printf("listening on port: %v\n", cfg.Host())
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("starting server: %s", err)
		}
	}()

	<-sig
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, idleTimeout)
	defer timeoutCancel()
	err := server.Shutdown(timeoutCtx)
	if err != nil {
		log.Fatalf("shutting down server: %s", err)
	}
}
