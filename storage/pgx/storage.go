package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExtContext interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	ExtContext
}

type Storage interface {
	Close()
	Ping(ctx context.Context) error
	Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error)
	ExtContext
}

type Option = func(pool *pgxpool.Pool, st Storage) Storage

type storage struct {
	*pgxpool.Pool
}

func New(ctx context.Context, cfg Config, opts ...Option) (Storage, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", cfg.Timeout/time.Millisecond)

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	var st Storage = &storage{
		Pool: db,
	}

	for _, opt := range opts {
		st = opt(db, st)
	}

	return st, nil
}

func (s *storage) Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error) {
	txOpts := pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}

	if opts != nil {
		txOpts.IsoLevel = opts.IsoLevel
	}

	tx, err := s.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *storage) Ping(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}
