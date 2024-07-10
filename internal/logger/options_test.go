package logger

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func Test_WithLogLevelString(t *testing.T) {
	tt := []struct {
		Name     string
		Input    string
		Expected zapcore.Level
	}{
		{"debug", "debug", zapcore.DebugLevel},
		{"info", "info", zapcore.InfoLevel},
		{"warn", "warn", zapcore.WarnLevel},
		{"error", "error", zapcore.ErrorLevel},
		{"dpanic", "dpanic", zapcore.DPanicLevel},
		{"panic", "panic", zapcore.PanicLevel},
		{"fatal", "fatal", zapcore.FatalLevel},

		{"invalid", "invalid", zapcore.InfoLevel}, // the default
		{"", "", zapcore.InfoLevel},               // the default
	}

	for _, tc := range tt {
		opt := WithLogLevelString(tc.Input)
		cfg := zap.Config{Level: zap.NewAtomicLevelAt(zapcore.InfoLevel)}
		opt.apply(&cfg)
		require.Equal(t, zap.NewAtomicLevelAt(tc.Expected), cfg.Level)
	}
}
