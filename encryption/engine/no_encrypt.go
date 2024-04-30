package engine

import (
	"context"

	"github.com/underbek/examples-go/encryption/domain"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
)

type NoEncrypt struct{}

// NewNoEncrypt returns an engine that returns raw data without any encryption/decryption
func NewNoEncrypt() *NoEncrypt {
	return &NoEncrypt{}
}

func (n NoEncrypt) Encrypt(ctx context.Context, _ domain.EncryptorData, value string) (encryptedValue string, err error) {
	_, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "no_encrypt", "engine.Encrypt")
	defer span.End()

	return value, nil
}

func (n NoEncrypt) Decrypt(ctx context.Context, _ domain.EncryptorData, encryptedValue string) (value string, err error) {
	_, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "no_encrypt", "engine.Decrypt")
	defer span.End()

	return encryptedValue, nil
}
