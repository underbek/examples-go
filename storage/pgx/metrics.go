package pgx

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type DBCollector struct {
	conn      Storage
	collector *pgxPoolCollector
}

type TxCollector struct {
	tx        Transaction
	collector *pgxPoolCollector
	start     time.Time
}

const (
	successErrorCode         = "00000"
	undefinedErrorCode       = "XX011"
	contextCanceledErrorCode = "XX012"

	selectMethod    = "SELECT"
	insertMethod    = "INSERT"
	updateMethod    = "UPDATE"
	deleteMethod    = "DELETE"
	undefinedMethod = "UNDEFINED"
)

func WithMetrics() Option {
	return func(pool *pgxpool.Pool, st Storage) Storage {
		collector := newPgxPoolCollector(pool)

		prometheus.MustRegister(collector)

		return &DBCollector{
			collector: collector,
			conn:      st,
		}
	}
}

func parseErrorCode(err error) string {
	if err == nil {
		return successErrorCode
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	if err == context.Canceled || err == context.DeadlineExceeded {
		return contextCanceledErrorCode
	}

	return undefinedErrorCode
}

func parseSQLMethod(sql string) string {
	sql = strings.TrimSpace(sql)

	sl := strings.Split(sql, " ")
	if len(sl) == 0 {
		return undefinedMethod
	}

	switch strings.ToUpper(sl[0]) {
	case selectMethod:
		return selectMethod
	case insertMethod:
		return insertMethod
	case updateMethod:
		return updateMethod
	case deleteMethod:
		return deleteMethod
	}

	return undefinedMethod
}

func collectMetrics(c *pgxPoolCollector, start time.Time, sql string, err error) {
	method := parseSQLMethod(sql)

	c.queryExecutedHistogram.WithLabelValues(method).Observe(time.Since(start).Seconds())
	c.queryExecutedCounter.WithLabelValues(method, parseErrorCode(err)).Inc()
}

func (c *DBCollector) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	commandTag, err := c.conn.Exec(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, err)

	return commandTag, err
}

func (c *DBCollector) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := c.conn.Query(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, err)

	return rows, err
}

func (c *DBCollector) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := c.conn.QueryRow(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, nil)

	return row
}

func (c *DBCollector) Close() {
	prometheus.Unregister(c.collector)
	c.conn.Close()
}

// Begin returned transaction wrapper
func (c *DBCollector) Begin(ctx context.Context, opts *pgx.TxOptions) (Transaction, error) {
	tx, err := c.conn.Begin(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &TxCollector{
		tx:        tx,
		collector: c.collector,
		start:     time.Now(),
	}, nil
}

//
// Transaction methods below
//

func (c *TxCollector) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	commandTag, err := c.tx.Exec(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, err)

	return commandTag, err
}

func (c *TxCollector) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := c.tx.Query(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, err)

	return rows, err
}

func (c *TxCollector) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := c.tx.QueryRow(ctx, sql, args...)

	collectMetrics(c.collector, start, sql, nil)

	return row
}

func (c *TxCollector) collectTxMetrics(err error) error {
	if c.start.IsZero() {
		return err
	}

	start := c.start
	c.start = time.Time{}

	c.collector.txExecutedHistogram.WithLabelValues().Observe(time.Since(start).Seconds())
	c.collector.txExecutedCounter.WithLabelValues(parseErrorCode(err)).Inc()

	return err
}

func (c *TxCollector) Commit(ctx context.Context) error {
	return c.collectTxMetrics(c.tx.Commit(ctx))
}

func (c *TxCollector) Rollback(ctx context.Context) error {
	return c.collectTxMetrics(c.tx.Rollback(ctx))
}
