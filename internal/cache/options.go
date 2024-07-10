package cache

import (
	"go.opentelemetry.io/otel/trace"
)

// config for our cache configuration.
type config struct {
	connectionPoolSize int
	debugEnabled       bool
	tracerProvider     trace.TracerProvider
}

// Option is an interface which allows us to apply options to the config.
type Option interface{ apply(*config) }

// optionFunc is a helper function which implements Option
type optionFunc func(*config)

// apply implements Option interface.
func (o optionFunc) apply(cfg *config) { o(cfg) }

func WithDebugMode(debugEnabled bool) Option {
	return optionFunc(func(cfg *config) { cfg.debugEnabled = debugEnabled })
}

func WithTraceProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(cfg *config) { cfg.tracerProvider = tp })
}

func WithConnectionPoolSize(n int) Option {
	return optionFunc(func(cfg *config) { cfg.connectionPoolSize = n })
}
