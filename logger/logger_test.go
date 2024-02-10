package logger

import (
	"context"
	"errors"
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
		With("slice_stringer", SliceStringer([]interface{}{12})).
		Info("test debug message with fields")
}

func TestLoggerWithContext(t *testing.T) {
	logger, err := New(false)
	require.NoError(t, err)

	data := ContextData{
		fields: make(map[string]any),
	}

	data.fields["duration"] = time.Second * 10
	data.fields["string"] = "string"
	data.fields["int"] = 12
	data.fields["float"] = 12.12
	data.fields["slice_stringer"] = SliceStringer([]interface{}{12})

	data.meta = map[string]string{"orderID": "some order ID"}

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

type tmpError struct {
	data map[string]any
}

func (t *tmpError) Error() string {
	return "tmpError"
}

func (t *tmpError) ErrorData() map[string]any {
	return t.data
}

func Test_parseErrorData(t *testing.T) {
	type testStruct struct {
		a string
		b int
	}

	tests := []struct {
		name string
		err  error
		want map[string]any
	}{
		{
			name: "Error is nil",
			err:  nil,
		},
		{
			name: "Nothing found",
			err:  errors.New("some error"),
		},
		{
			name: "Happy path",
			err: &tmpError{
				data: map[string]any{
					"test": 2,
					"kek":  []string{"1"},
					"lol":  testStruct{a: "3", b: 4},
				},
			},
			want: map[string]any{
				"test": 2,
				"kek":  []string{"1"},
				"lol":  testStruct{a: "3", b: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, parseErrorData(tt.err))
		})
	}
}
