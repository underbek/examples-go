package service

import (
	"context"
	"errors"
	"github.com/underbek/examples-go/encryption/config"
	domain2 "github.com/underbek/examples-go/encryption/domain"
	"github.com/underbek/examples-go/logger"
	"strconv"
	"time"

	"github.com/underbek/examples-go/encryption/engine"
	goKitErrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/storage/pgx"
)

const vault = "vault"

var (
	errTimeExceeded = errors.New("max execution time exceeded")
)

type Postgres interface {
	GetEncryptorData(ctx context.Context, encryptorType string, id string) (domain2.EncryptorData, error)
}

type Engine interface {
	Encrypt(ctx context.Context, encryptorData domain2.EncryptorData, value string) (encryptedValue string, err error)
	Decrypt(ctx context.Context, encryptorData domain2.EncryptorData, encryptedValue string) (value string, err error)
}

type EncryptionService interface {
	Encrypt(ctx context.Context, req domain2.EncryptRequest) (domain2.EncryptResponse, error)
	Decrypt(ctx context.Context, req domain2.DecryptRequest) (domain2.DecryptResponse, error)
}

type Service struct {
	logger        *logger.Logger
	storage       pgx.Storage
	createStorage CreateStorage
	config        config.Config
}

type CreateStorage = func(ext pgx.ExtContext) Postgres

func New(logger *logger.Logger, storage pgx.Storage, createStorage CreateStorage, config config.Config) *Service {
	return &Service{
		logger:        logger,
		storage:       storage,
		createStorage: createStorage,
		config:        config,
	}
}

func (s Service) Encrypt(ctx context.Context, req domain2.EncryptRequest) (domain2.EncryptResponse, error) {
	storage := s.createStorage(s.storage)
	data, err := storage.GetEncryptorData(ctx, req.Type.String(), "")
	if err != nil {
		return domain2.EncryptResponse{}, goKitErrors.Wrap(err, goKitErrors.TypeDatabase, "repository.GetEncryptorData")
	}

	maxAttempts := s.getAttempts(data.Engine)

	var encryptedValue string
	for attempt := uint(1); attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return domain2.EncryptResponse{}, goKitErrors.Wrap(errTimeExceeded, goKitErrors.TypeInternal, "failed to get decrypt data")
		default:
			s.logger.WithCtx(ctx).With("attempt", attempt).Info("attempt to encrypt")

			var eng Engine
			eng, err = s.getEngine(data.Engine, data.Additional)
			if err != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("failed to get engine to encrypt data")
				continue
			}

			encryptedValue, err = eng.Encrypt(ctx, data, req.Value)
			if err != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("failed to make encryption")
			}
		}

		if err == nil || maxAttempts == attempt { //to avoid sleep at the last attempt
			break
		}

		time.Sleep(s.config.Pool.RetryDuration)
	}

	if err != nil {
		return domain2.EncryptResponse{}, goKitErrors.Wrap(err, goKitErrors.TypeInternal, "failed to get encrypt data")
	}

	encryptorID := strconv.FormatInt(data.ID, 10)
	return domain2.EncryptResponse{
		EncryptedValue: encryptedValue,
		EncryptorID:    encryptorID,
	}, nil
}

func (s Service) Decrypt(ctx context.Context, req domain2.DecryptRequest) (domain2.DecryptResponse, error) {
	var logTag, logEngine string
	defer func() {
		s.logger.
			WithCtx(ctx).
			With("tag", logTag).
			With("encryptor_id", req.EncryptorID).
			With("engine", logEngine).
			Info("Decryption")
	}()

	storage := s.createStorage(s.storage)
	data, err := storage.GetEncryptorData(ctx, "", req.EncryptorID)
	if err != nil {
		return domain2.DecryptResponse{}, goKitErrors.Wrap(err, goKitErrors.TypeDatabase, "repository.GetEncryptorDataByID")
	}
	logTag, logEngine = data.EncryptorType.String(), data.Engine

	maxAttempts := s.getAttempts(data.Engine)

	var decryptedValue string
	for attempt := uint(1); attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return domain2.DecryptResponse{}, goKitErrors.Wrap(errTimeExceeded, goKitErrors.TypeInternal, "failed to get decrypt data")
		default:
			s.logger.WithCtx(ctx).With("attempt", attempt).Info("attempt to decrypt")

			var eng Engine
			eng, err = s.getEngine(data.Engine, data.Additional)
			if err != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("failed to get engine to decrypt data")
				continue
			}

			decryptedValue, err = eng.Decrypt(ctx, data, req.EncryptedValue)
			if err != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("failed to make decryption")
			}
		}

		if err == nil || maxAttempts == attempt { //to avoid sleep at the last attempt
			break
		}

		time.Sleep(s.config.Pool.RetryDuration)
	}

	if err != nil {
		return domain2.DecryptResponse{}, goKitErrors.Wrap(err, goKitErrors.TypeInternal, "failed to get decrypt data")
	}

	return domain2.DecryptResponse{
		Value: decryptedValue,
		Type:  data.EncryptorType,
	}, nil
}

func (s *Service) getAttempts(engine string) uint {
	if s.config.Pool.RetryAttempts > 0 && engine == vault {
		return s.config.Pool.RetryAttempts
	}

	return 1
}

func (s *Service) getEngine(e string, attributes domain2.Attributes) (Engine, error) {
	switch e {
	case vault:
		return engine.NewVault(
			s.config.Vault,
			s.config.Pool.PoolMode,
			attributes["path"].(string),
			attributes["key"].(string),
		)

	default:
		return engine.NewNoEncrypt(), nil
	}

}
