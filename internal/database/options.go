package database

import (
	"context"
	"time"

	"github.com/upper/db/v4"
	"go.opentelemetry.io/otel/trace"
)

// Config the set of configurations required for a database connection.
type Config struct {
	// clientFunc connects to the database and returns a client interface.
	clientFunc func(context.Context, *Config) (db.Session, error)
	// credentials to use to connect to the database.
	credentials *Credentials

	// MaxIdleConns sets the default maximum number of connections in the idle connection pool.
	MaxIdleConns int
	// MaxOpenConns sets the default maximum number of open connections to the database.
	MaxOpenConns int
	// ConnMaxLifetime sets the default maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration
	// DebugEnabled whether debug settings should be configured. This value should be false in production, to avoid
	// tracing sensitive information.
	DebugEnabled bool
	// TracerProvider for tracing queries
	TracerProvider trace.TracerProvider
}

// Option is an interface which allows us to apply options to the config.
type Option interface{ apply(*Config) }

// optionFunc is a helper function which implements Option.
type optionFunc func(*Config)

// apply implements Option interface.
func (o optionFunc) apply(cfg *Config) { o(cfg) }

// WithClientFunc sets the client func to connect to the database with.
func WithClientFunc(fn func(context.Context, *Config) (db.Session, error)) Option {
	return optionFunc(func(cfg *Config) { cfg.clientFunc = fn })
}

// WithDebugMode sets the DebugEnabled configuration to true. This value should be false in production, to avoid
// tracing sensitive information.
func WithDebugMode(debugEnabled bool) Option {
	return optionFunc(func(cfg *Config) { cfg.DebugEnabled = debugEnabled })
}

// WithTraceProvider sets the trace provider for tracing.
func WithTraceProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(cfg *Config) { cfg.TracerProvider = tp })
}

// WithMaxIdleConns sets the MaxIdleConns configuration.
func WithMaxIdleConns(n int) Option {
	return optionFunc(func(cfg *Config) { cfg.MaxIdleConns = n })
}

// WithMaxOpenConns sets the MaxOpenConns configuration.
func WithMaxOpenConns(n int) Option { return optionFunc(func(cfg *Config) { cfg.MaxOpenConns = n }) }

// WithConnMaxLifetime sets the ConnMaxLifetime configuration.
func WithConnMaxLifetime(t time.Duration) Option {
	return optionFunc(func(cfg *Config) { cfg.ConnMaxLifetime = t })
}
