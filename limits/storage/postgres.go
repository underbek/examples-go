package storage

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/underbek/examples-go/logger"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
)

type Storage struct {
	logger   *logger.Logger
	ext      goKitPgx.ExtContext
	maxLimit int
}

func New(logger *logger.Logger, ext goKitPgx.ExtContext, maxLimit int) *Storage {
	return &Storage{
		logger:   logger,
		ext:      ext,
		maxLimit: maxLimit,
	}
}
