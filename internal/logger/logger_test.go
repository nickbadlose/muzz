package logger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New()
	require.NoError(t, err)
}

// captureOutput helper function which sets up the logger in a way
// in which we can capture the output and return it as a string to test against.
func captureOutput(t *testing.T, fn func(), opts ...Option) []byte {
	// generate a temp file.
	f, err := os.CreateTemp("", "logger-test-*")
	if err != nil {
		panic(err)
	}

	// generate a new logger pointing to the temp file.
	err = New(
		append(
			[]Option{
				ReplaceErrorOutputPaths(f.Name()),
				ReplaceOutputPaths(f.Name()),
				WithLogLevelString("debug"),
			},
			opts...,
		)...,
	)
	require.NoError(t, err)

	// run any logging output.
	fn()

	// close the temp fir to flush output
	require.NoError(t, f.Close())

	// defer that we remove the temp file and close the logger.
	t.Cleanup(func() {
		require.NoError(t, Close())
		require.NoError(t, os.Remove(f.Name()))
	})

	// read the output from the temp file and return it as a string.
	b, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	return b
}

type logOutput struct {
	Level   string    `json:"level"`
	Time    time.Time `json:"time"`
	Caller  string    `json:"caller"`
	Message string    `json:"msg"`
	TraceID string    `json:"traceID"`
	Stack   string    `json:"stacktrace"`
	Query   string    `json:"query"`
	Error   string    `json:"error"`
}

func TestInfo(t *testing.T) {
	output := captureOutput(t, func() {
		Info(context.TODO(), `test`)
	})

	var lo logOutput
	require.NoError(t, json.Unmarshal(output, &lo))
	require.Equal(t, `info`, lo.Level)
	require.Equal(t, `test`, lo.Message)
}

func TestDebug(t *testing.T) {
	output := captureOutput(t, func() {
		Debug(context.TODO(), `test`)
	})

	var lo logOutput
	require.NoError(t, json.Unmarshal(output, &lo))
	require.Equal(t, `debug`, lo.Level)
	require.Equal(t, `test`, lo.Message)
}

func TestWarn(t *testing.T) {
	output := captureOutput(t, func() {
		Warn(context.TODO(), `test`)
	})

	var lo logOutput
	require.NoError(t, json.Unmarshal(output, &lo))
	require.Equal(t, `warn`, lo.Level)
}

func TestError(t *testing.T) {
	output := captureOutput(t, func() {
		Error(context.TODO(), `test`, errors.New("test error"))
	})

	var lo logOutput
	require.NoError(t, json.Unmarshal(output, &lo))
	require.Equal(t, `error`, lo.Level)
	require.Equal(t, "test error", lo.Error)
	require.Equal(t, `test`, lo.Message)
}

func TestLogger_MaybeError(t *testing.T) {
	t.Run("error is not nil", func(t *testing.T) {
		output := captureOutput(t, func() {
			MaybeError(context.TODO(), `test`, errors.New("test error"))
		})

		var lo logOutput
		require.NoError(t, json.Unmarshal(output, &lo))
		require.Equal(t, `error`, lo.Level)
		require.Equal(t, "test error", lo.Error)
		require.Equal(t, `test`, lo.Message)
	})

	t.Run("error is nil", func(t *testing.T) {
		output := captureOutput(t, func() {
			MaybeError(context.TODO(), `test`, nil)
		})

		require.Equal(t, 0, len(output))
	})
}
