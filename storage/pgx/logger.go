package pgx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/underbek/examples-go/logger"
)

type DBLogger struct {
	conn   Storage
	logger *logger.Logger
}

type TxLogger struct {
	tx     Transaction
	logger *logger.Logger
}

func WithLogger(logger *logger.Logger) Option {
	return func(_ *pgxpool.Pool, st Storage) Storage {
		return &DBLogger{
			conn:   st,
			logger: logger,
		}
	}
}

func (l *DBLogger) Exec(ctx context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	commandTag, err = l.conn.Exec(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).WithError(err).Debug("Exec")

	return commandTag, err
}

func (l *DBLogger) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := l.conn.Query(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).WithError(err).Debug("Query")

	return rows, err
}

func (l *DBLogger) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	row := l.conn.QueryRow(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).Debug("QueryRow")

	return row
}

func (l *DBLogger) Close() {
	l.conn.Close()
	l.logger.Debug("Close")
}

func (l *DBLogger) Ping(ctx context.Context) error {
	return l.conn.Ping(ctx)
}

// Begin returned transaction wrapper
func (l *DBLogger) Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error) {
	tx, err := l.conn.Begin(ctx, opts)
	l.logger.WithCtx(ctx).WithError(err).With("options", opts).Debug("Begin")
	if err != nil {
		return nil, err
	}

	return &TxLogger{
		tx:     tx,
		logger: l.logger,
	}, nil
}

//
// Transaction methods below
//

func (l *TxLogger) Exec(ctx context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	commandTag, err = l.tx.Exec(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).WithError(err).Debug("ExecTx")

	return commandTag, err
}

func (l *TxLogger) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := l.tx.Query(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).WithError(err).Debug("QueryTx")

	return rows, err
}

func (l *TxLogger) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	row := l.tx.QueryRow(ctx, sql, args...)
	l.logger.WithCtx(ctx).With("sql", sql).With("arguments", logger.SliceStringer(args)).Debug("QueryRowTx")

	return row
}

func (l *TxLogger) Commit(ctx context.Context) error {
	err := l.tx.Commit(ctx)
	l.logger.WithCtx(ctx).WithError(err).Debug("Commit")

	return err
}

func (l *TxLogger) Rollback(ctx context.Context) error {
	err := l.tx.Rollback(ctx)

	//to remove "tx is closed from" logs
	if errors.Is(err, pgx.ErrTxClosed) {
		err = nil
	}

	l.logger.WithCtx(ctx).WithError(err).Debug("Rollback")

	return err
}
