package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	internal *zap.Logger
	fields   []zap.Field
}

// New makes new Logger by debug level.
func New(debug bool) (*Logger, error) {
	zapLogger, err := newZapLogger(debug)
	if err != nil {
		return nil, err
	}

	return &Logger{
		internal: zapLogger,
	}, nil
}

func (l *Logger) Sync() error {
	return l.internal.Sync()
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{
		internal: l.internal.Named(name),
	}
}

func (l *Logger) Internal() any {
	return l.internal
}

func (l *Logger) Debug(msg string) {
	l.internal.Debug(msg, l.fields...)
}

func (l *Logger) Info(msg string) {
	l.internal.Info(msg, l.fields...)
}

func (l *Logger) Warn(msg string) {
	l.internal.Warn(msg, l.fields...)
}

func (l *Logger) Error(msg string) {
	l.internal.Error(msg, l.fields...)
}

func (l *Logger) Panic(msg string) {
	l.internal.Panic(msg, l.fields...)
}

func (l *Logger) Fatal(msg string) {
	l.internal.Fatal(msg, l.fields...)
}

func (l *Logger) With(key string, value any) *Logger {
	return &Logger{
		internal: l.internal,
		fields:   append(l.fields, zap.Any(key, value)),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		internal: l.internal,
		fields:   append(l.fields, zap.Error(err)),
	}
}

func newZapLogger(debug bool) (*zap.Logger, error) {
	cfg := zap.NewProductionEncoderConfig()
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(os.Stdout), level)
	return zap.New(core).
			WithOptions(zap.WithCaller(true)).
			WithOptions(zap.AddCallerSkip(1)),
		nil
}
