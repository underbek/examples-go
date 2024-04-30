package engine

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/underbek/examples-go/encryption/config"
	"github.com/underbek/examples-go/encryption/domain"
	gokitErrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
)

type uniqueKey struct {
	path string
	key  string
}

type connPoolWithToken struct {
	pool  ConnectionPool
	token string
}

var (
	mtx               = &sync.Mutex{}
	activeConnections map[uniqueKey]connPoolWithToken
)

func init() {
	activeConnections = make(map[uniqueKey]connPoolWithToken)
}

type Vault struct {
	conn ConnectionPool
}

// NewVault returns an engine that encrypts/decrypts data via hasicorp vault
func NewVault(cfg config.VaultENV, poolMode bool, path, key string) (*Vault, error) {
	token, err := getVaultToken(cfg.TokenPath)
	if err != nil {
		return nil, gokitErrors.Wrap(err, gokitErrors.TypeInternal, "get token")
	}

	uk := uniqueKey{
		path: path,
		key:  key,
	}

	mtx.Lock()
	defer mtx.Unlock()

	if connWithToken, ok := activeConnections[uk]; !ok || connWithToken.token != token {
		if ok {
			connWithToken.pool.Close()
		}

		var conn ConnectionPool
		if poolMode {
			conn, err = NewConnectionPool(cfg, token, key, path)
		} else {
			conn, err = NewSingleConnection(cfg, token, key, path)
		}
		if err != nil {
			return nil, gokitErrors.Wrap(err, gokitErrors.TypeInternal, "transit wrapper")
		}

		activeConnections[uk] = connPoolWithToken{
			pool:  conn,
			token: token,
		}
	}

	return &Vault{conn: activeConnections[uk].pool}, nil
}

func (v *Vault) Encrypt(ctx context.Context, _ domain.EncryptorData, value string) (encryptedValue string, err error) {
	_, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "vault", "engine.Encrypt")
	defer span.End()

	conn, err := v.conn.Acquire()
	if err != nil {
		return "", gokitErrors.Wrap(err, gokitErrors.TypeExternal, "failed to acquire connection")
	}
	defer v.conn.Release(conn)

	_, wrapperSpan := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "vault", "wrapper.Encrypt")
	defer wrapperSpan.End()

	data, err := conn.wrapper.GetClient().Encrypt([]byte(value))
	if err != nil {
		return "", gokitErrors.Wrap(err, gokitErrors.TypeExternal, "v.wrapper.GetClient().Encrypt")
	}

	encryptedValue = string(data)
	return encryptedValue, nil
}

func (v *Vault) Decrypt(ctx context.Context, _ domain.EncryptorData, encryptedValue string) (value string, err error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "vault", "engine.Decrypt")
	defer span.End()

	conn, err := v.conn.Acquire()
	if err != nil {
		return "", gokitErrors.Wrap(err, gokitErrors.TypeExternal, "failed to acquire connection")
	}
	defer v.conn.Release(conn)

	_, wrapperSpan := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "vault", "wrapper.Decrypt")
	defer wrapperSpan.End()

	data, err := conn.wrapper.GetClient().Decrypt([]byte(encryptedValue))
	if err != nil {
		return "", gokitErrors.Wrap(err, gokitErrors.TypeExternal, "v.wrapper.GetClient().Decrypt")
	}

	value = string(data)
	return value, nil
}

func getVaultToken(path string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(path))

	if err != nil {
		return "", err
	}

	return string(content), nil
}
