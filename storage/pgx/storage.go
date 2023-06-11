package pgx

import (
	"context"

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

type Config struct {
	DSN string `env:"POSTGRES_DSN" valid:"required"`
}

type Storage interface {
	Close()
	Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error)
	ExtContext
}

type Option = func(pool *pgxpool.Pool, st Storage) Storage

type storage struct {
	*pgxpool.Pool
}

func New(ctx context.Context, dataSource string, opts ...Option) (Storage, error) {
	db, err := pgxpool.New(ctx, dataSource)
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
		IsoLevel: pgx.Serializable,
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
