package pgx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type DBTracer struct {
	conn Storage
}

type TxTracer struct {
	tx Transaction
}

func WithTrace() Option {
	return func(_ *pgxpool.Pool, st Storage) Storage {
		return &DBTracer{
			conn: st,
		}
	}
}

func (t *DBTracer) Exec(ctx context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "Exec",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.conn.Exec(ctx, sql, args...)
}

func (t *DBTracer) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "Query",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.conn.Query(ctx, sql, args...)
}

func (t *DBTracer) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "QueryRow",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.conn.QueryRow(ctx, sql, args...)
}

func (t *DBTracer) Close() {
	t.conn.Close()
}

func (t *DBTracer) Ping(ctx context.Context) error {
	return t.conn.Ping(ctx)
}

// Begin returned transaction wrapper
func (t *DBTracer) Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "Begin")
	defer span.End()

	tx, err := t.conn.Begin(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &TxTracer{
		tx: tx,
	}, nil
}

//
// Transaction methods below
//

func (t *TxTracer) Exec(ctx context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "ExecTx",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.tx.Exec(ctx, sql, args...)
}

func (t *TxTracer) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "QueryTx",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.tx.Query(ctx, sql, args...)
}

func (t *TxTracer) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "QueryRowTx",
		trace.WithAttributes(attribute.String("sql", sql)))
	defer span.End()

	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *TxTracer) Commit(ctx context.Context) error {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "Commit")
	defer span.End()

	return t.tx.Commit(ctx)
}

func (t *TxTracer) Rollback(ctx context.Context) error {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindInternal, "pgx", "Rollback")
	defer span.End()

	return t.tx.Rollback(ctx)
}
