package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option is an interface which allows us to apply options to the config.
type Option interface{ apply(*zap.Config) }

// optionFunc is a helper function which implements Option
type optionFunc func(*zap.Config)

// apply implements Option interface.
func (o optionFunc) apply(cfg *zap.Config) { o(cfg) }

// WithLogLevel option to change the log level to use.
func WithLogLevel(l zapcore.Level) Option {
	return optionFunc(func(cfg *zap.Config) { cfg.Level = zap.NewAtomicLevelAt(l) })
}

// WithLogLevelString option to change the log level to use by a string.
func WithLogLevelString(s string) Option {
	return optionFunc(func(cfg *zap.Config) {
		l, ok := logLevelMappings[s]
		if !ok {
			return
		}
		cfg.Level = zap.NewAtomicLevelAt(l)
	})
}
