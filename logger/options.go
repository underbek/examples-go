package logger

import "go.uber.org/zap"

type Option interface {
	apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) apply(logger *Logger) {
	f(logger)
}

func (l *Logger) WithOptions(opts ...Option) *Logger {
	c := &Logger{
		internal: l.internal,
		fields:   l.fields,
	}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func AddCallerSkip(skip int) Option {
	return optionFunc(func(l *Logger) {
		l.internal = l.internal.WithOptions(zap.AddCallerSkip(skip))
	})
}
