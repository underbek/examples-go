package logger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		debug bool
	}{
		{debug: true},
		{debug: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Logger with debug=%v", tt.debug), func(t *testing.T) {
			logger, err := New(tt.debug)
			require.NoError(t, err)

			logger.Debug("test debug message")
			logger.Info("test info message")
			logger.Warn("test warn message")
			logger.Error("test error message")

			require.Panics(t, func() {
				logger.Panic("test panic message")
			})
		})
	}
}

func TestLoggerFields(t *testing.T) {
	logger, err := New(false)
	require.NoError(t, err)

	logger.
		With("duration", time.Second*10).
		With("string", "string").
		With("int", 12).
		With("float", 12.12).
		Info("test debug message with fields")
}

func TestLoggerWithContext(t *testing.T) {
	logger, err := New(false)
	require.NoError(t, err)

	data := make(ContextData)
	data["duration"] = time.Second * 10
	data["string"] = "string"
	data["int"] = 12
	data["float"] = 12.12

	ctx := data.ToCtx(context.Background())

	logger.
		WithCtx(ctx).
		Info("context")
}

func checkFunc(logger *Logger) {
	logger.Info("checkFunc")
}

func TestWithOptions(t *testing.T) {
	logger, err := New(false)
	require.NoError(t, err)

	loggerWithSkip := logger.WithOptions(AddCallerSkip(1))

	checkFunc(logger)
	checkFunc(loggerWithSkip)
}
