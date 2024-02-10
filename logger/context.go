package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextDataKeyType string

const contextDataKey contextDataKeyType = "contextDataKey"

const (
	Meta = "domain_meta"
)

type ContextData struct {
	fields map[string]any
	meta   map[string]string
}

func (c ContextData) ToCtx(parent context.Context) context.Context {
	return context.WithValue(parent, contextDataKey, c)
}

// AddCtxValue returned context with specific value for further use in logger
func AddCtxValue(ctx context.Context, key string, value any) context.Context {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		ctxData := ContextData{
			fields: map[string]any{
				key: value,
			},
			meta: make(map[string]string),
		}
		return ctxData.ToCtx(ctx)
	}

	data.fields[key] = value
	return data.ToCtx(ctx)
}

func GetFieldsFromContext(ctx context.Context) []zap.Field {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		return nil
	}

	fields := make([]zap.Field, 0, len(data.fields))
	for key, value := range data.fields {
		fields = append(fields, zap.Any(key, value))
	}

	fields = append(fields, GetMetaFieldsFromContext(ctx)...)

	return fields
}

func GetMetaFieldsFromContext(ctx context.Context) []zap.Field {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		return nil
	}

	fields := make([]zap.Field, 0, len(data.fields))

	if len(data.meta) != 0 {
		fields = append(fields, zap.Any(Meta, data.meta))
	}

	return fields
}

func GetDataFromContext(ctx context.Context) map[string]any {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		return nil
	}

	fields := make(map[string]interface{}, len(data.fields))
	for key, value := range data.fields {
		fields[key] = value
	}

	if len(data.meta) != 0 {
		fields[Meta] = data.meta
	}

	return fields
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	if _, ok := ctx.Value(contextDataKey).(ContextData); !ok {
		return l
	}

	newLogger := &Logger{
		internal: l.internal,
		fields:   l.fields,
	}

	newFields := GetFieldsFromContext(ctx)

	newLogger.fields = append(newLogger.fields, newFields...)

	return newLogger
}

func AddCtxMetaValue(ctx context.Context, key, value string) context.Context {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		ctxData := ContextData{
			fields: map[string]any{},
			meta:   map[string]string{key: value},
		}
		return ctxData.ToCtx(ctx)
	}

	data.meta[key] = value

	return data.ToCtx(ctx)
}

func AddCtxMetaValues(ctx context.Context, values map[string]string) context.Context {
	if len(values) == 0 {
		return ctx
	}

	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		ctxData := ContextData{
			fields: map[string]any{},
			meta:   values,
		}
		return ctxData.ToCtx(ctx)
	}

	for k, v := range values {
		data.meta[k] = v
	}

	return data.ToCtx(ctx)
}

func ParseCtxMeta(ctx context.Context) map[string]string {
	data, ok := ctx.Value(contextDataKey).(ContextData)
	if !ok {
		return nil
	}

	return data.meta
}
