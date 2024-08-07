package logger

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// the amount of function call frames to skip when determining the caller function.
	callerSkip = 4
	// traceIDKey for logging the trace ID..
	traceIDKey = "trace_id"
)

var (
	// lock allows for concurrent safe access to the global logger instance.
	lock = &sync.RWMutex{}

	// global instance of the internal logger.
	logger *zap.Logger

	// logLevelMappings sets the available log level mappings.
	logLevelMappings = logLevels()
)

type (
	// Field a type alias to the internal field type.
	Field = zap.Field
	// logFunc matches the signature for the logging functions available on the internal logger.
	logFunc func(msg string, fields ...Field)
)

// New initialises a new production logger.
func New(opts ...Option) error {
	lock.Lock()
	defer lock.Unlock()

	// update encode config to use RFC3339 for the time.
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encodeConfig.TimeKey = `time`

	config := zap.NewProductionConfig()
	config.EncoderConfig = encodeConfig
	config.DisableStacktrace = true

	// apply external options.
	for _, opt := range opts {
		opt.apply(&config)
	}

	zl, err := config.Build(zap.AddCallerSkip(callerSkip))
	if err != nil {
		return err
	}

	logger = zl

	return nil
}

func addTraceID(ctx context.Context) zap.Field {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	return zap.String(traceIDKey, traceID)
}

// decorateLogs with trace information.
func decorateLogs(ctx context.Context, fn logFunc, msg string, fields ...Field) {
	fn(msg, append(fields, addTraceID(ctx))...)
}

// wrapNoLogger helper func which wraps if the logger is not initialised.
func wrapNoLogger(fn func()) {
	lock.RLock()
	defer lock.RUnlock()

	if logger == nil {
		return
	}

	fn()
}

// Close closes a logger if its set-up, this flushes
// any remaining logs first.
func Close() error {
	lock.Lock()
	defer lock.Unlock()

	if logger == nil {
		return nil
	}

	err := logger.Sync()
	logger = nil
	return err
}

// Info logs with the info log level.
func Info(ctx context.Context, msg string, fields ...Field) {
	wrapNoLogger(func() { decorateLogs(ctx, logger.Info, msg, fields...) })
}

// Warn logs with the warn log level.
func Warn(ctx context.Context, msg string, fields ...Field) {
	wrapNoLogger(func() { decorateLogs(ctx, logger.Warn, msg, fields...) })
}

// Error logs with the error log level.
func Error(ctx context.Context, msg string, err error, fields ...Field) {
	wrapNoLogger(func() { decorateLogs(ctx, logger.Error, msg, append([]Field{zap.Error(err)}, fields...)...) })
}

// Debug logs with the debug log level.
func Debug(ctx context.Context, msg string, fields ...Field) {
	wrapNoLogger(func() { decorateLogs(ctx, logger.Debug, msg, fields...) })
}

// Fatal logs with the Fatal log level. calling this will also cause an os.Exit(1).
func Fatal(ctx context.Context, msg string, fields ...Field) {
	wrapNoLogger(func() { decorateLogs(ctx, logger.Fatal, msg, fields...) })
}

// MaybeError logs with the error log level if one exists.
func MaybeError(ctx context.Context, msg string, err error, fields ...Field) {
	if err == nil {
		return
	}
	wrapNoLogger(func() { decorateLogs(ctx, logger.Error, msg, append([]Field{zap.Error(err)}, fields...)...) })
}

// availableLogLevels the available zap log levels.
var availableLogLevels = []zapcore.Level{
	zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel,
	zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel,
	zapcore.FatalLevel,
}

// logLevels converts the available log levels into a map with the string
// representation of the level as the key and the level as the value.
func logLevels() map[string]zapcore.Level {
	m := make(map[string]zapcore.Level, len(availableLogLevels))
	for _, ll := range availableLogLevels {
		m[ll.String()] = ll
	}
	return m
}
