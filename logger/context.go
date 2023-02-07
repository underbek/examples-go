package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextDataKeyType string

const contextDataKey contextDataKeyType = "contextDataKey"

type ContextData map[string]any

func (c ContextData) ToCtx(parent context.Context) context.Context {
	return context.WithValue(parent, contextDataKey, c)
}

// AddCtxValue returned context with specific value for further use in logger
func AddCtxValue(ctx context.Context, key string, value any) context.Context {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		ctxData := ContextData{key: value}
		return ctxData.ToCtx(ctx)
	}

	data[key] = value
	return data.ToCtx(ctx)
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		return l
	}

	newLogger := &Logger{
		internal: l.internal,
		fields:   l.fields,
	}

	for key, value := range data {
		newLogger.fields = append(newLogger.fields, zap.Any(key, value))
	}

	return newLogger
}
