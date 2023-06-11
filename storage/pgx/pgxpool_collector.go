package pgx

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// pgxPoolCollector is a Prometheus collector for pgx metrics.
// It implements the prometheus.Collector interface.
type pgxPoolCollector struct {
	db *pgxpool.Pool

	acquireConns            prometheus.Gauge
	canceledAcquireCount    prometheus.Gauge
	constructingConns       prometheus.Gauge
	emptyAcquireCount       prometheus.Gauge
	idleConns               prometheus.Gauge
	maxConns                prometheus.Gauge
	totalConns              prometheus.Gauge
	newConnsCount           prometheus.Gauge
	maxLifetimeDestroyCount prometheus.Gauge
	maxIdleDestroyCount     prometheus.Gauge

	queryExecutedCounter   *prometheus.CounterVec
	queryExecutedHistogram *prometheus.HistogramVec

	txExecutedCounter   *prometheus.CounterVec
	txExecutedHistogram *prometheus.HistogramVec
}

// newPgxPoolCollector returns a new pgxCollector.
// The dbName parameter is used to set the "db" label on the metrics.
// The db parameter is the pgxpool.Pool to collect metrics from.
// The db parameter must not be nil.
// The dbName parameter must not be empty.
func newPgxPoolCollector(db *pgxpool.Pool) *pgxPoolCollector {
	dbName := db.Config().ConnConfig.Database

	newGauge := func(name, help string) prometheus.Gauge {
		return prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "pgx",
				Subsystem:   "pool",
				Name:        name,
				Help:        help,
				ConstLabels: prometheus.Labels{"db": dbName},
			},
		)
	}

	return &pgxPoolCollector{
		db: db,
		acquireConns: newGauge(
			"acquire_connections",
			"Number of connections currently in the process of being acquired",
		),
		canceledAcquireCount: newGauge(
			"canceled_acquire_count",
			"Number of times a connection acquire was canceled",
		),
		constructingConns: newGauge(
			"constructing_connections",
			"Number of connections currently in the process of being constructed",
		),
		emptyAcquireCount: newGauge(
			"empty_acquire_count",
			"Number of times a connection acquire was canceled",
		),
		idleConns: newGauge(
			"idle_connections",
			"Number of idle connections in the pool",
		),
		maxConns: newGauge(
			"max_connections",
			"Maximum number of connections allowed in the pool",
		),
		totalConns: newGauge(
			"total_connections",
			"Total number of connections in the pool",
		),
		newConnsCount: newGauge(
			"new_connections_count",
			"Number of new connections created",
		),
		maxLifetimeDestroyCount: newGauge(
			"max_lifetime_destroy_count",
			"Number of connections destroyed due to MaxLifetime",
		),
		maxIdleDestroyCount: newGauge(
			"max_idle_destroy_count",
			"Number of connections destroyed due to MaxIdleTime",
		),
		queryExecutedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "pgx",
				Subsystem:   "pool",
				Name:        "query_executed_total",
				Help:        "Total number of query called, regardless of success (code=0) or failure.",
				ConstLabels: prometheus.Labels{"db": dbName},
			},
			[]string{"method", "code"},
		),
		queryExecutedHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   "pgx",
				Subsystem:   "pool",
				Name:        "query_executed_seconds",
				Help:        "Histogram of query latency (seconds) of db that had been application-level handled by the pgx connection.",
				ConstLabels: prometheus.Labels{"db": dbName},
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		txExecutedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "pgx",
				Subsystem:   "pool",
				Name:        "tx_executed_total",
				Help:        "Total number of tx called, regardless of success (code=0) or failure.",
				ConstLabels: prometheus.Labels{"db": dbName},
			},
			[]string{"code"},
		),
		txExecutedHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   "pgx",
				Subsystem:   "pool",
				Name:        "tx_executed_seconds",
				Help:        "Histogram of transaction latency (seconds) of db that had been application-level handled by the pgx connection.",
				ConstLabels: prometheus.Labels{"db": dbName},
				Buckets:     prometheus.DefBuckets,
			},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (p pgxPoolCollector) Describe(descs chan<- *prometheus.Desc) {
	p.acquireConns.Describe(descs)
	p.canceledAcquireCount.Describe(descs)
	p.constructingConns.Describe(descs)
	p.emptyAcquireCount.Describe(descs)
	p.idleConns.Describe(descs)
	p.maxConns.Describe(descs)
	p.totalConns.Describe(descs)
	p.newConnsCount.Describe(descs)
	p.maxLifetimeDestroyCount.Describe(descs)
	p.maxIdleDestroyCount.Describe(descs)

	p.queryExecutedCounter.Describe(descs)
	p.queryExecutedHistogram.Describe(descs)

	p.txExecutedCounter.Describe(descs)
	p.txExecutedHistogram.Describe(descs)
}

// Collect implements the prometheus.Collector interface.
func (p pgxPoolCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := p.db.Stat()

	p.acquireConns.Set(float64(stats.AcquiredConns()))
	p.canceledAcquireCount.Set(float64(stats.CanceledAcquireCount()))
	p.constructingConns.Set(float64(stats.ConstructingConns()))
	p.emptyAcquireCount.Set(float64(stats.EmptyAcquireCount()))
	p.idleConns.Set(float64(stats.IdleConns()))
	p.maxConns.Set(float64(stats.MaxConns()))
	p.totalConns.Set(float64(stats.TotalConns()))
	p.newConnsCount.Set(float64(stats.NewConnsCount()))
	p.maxLifetimeDestroyCount.Set(float64(stats.MaxLifetimeDestroyCount()))
	p.maxIdleDestroyCount.Set(float64(stats.MaxIdleDestroyCount()))

	p.acquireConns.Collect(metrics)
	p.canceledAcquireCount.Collect(metrics)
	p.constructingConns.Collect(metrics)
	p.emptyAcquireCount.Collect(metrics)
	p.idleConns.Collect(metrics)
	p.maxConns.Collect(metrics)
	p.totalConns.Collect(metrics)
	p.newConnsCount.Collect(metrics)
	p.maxLifetimeDestroyCount.Collect(metrics)
	p.maxIdleDestroyCount.Collect(metrics)

	p.queryExecutedCounter.Collect(metrics)
	p.queryExecutedHistogram.Collect(metrics)

	p.txExecutedCounter.Collect(metrics)
	p.txExecutedHistogram.Collect(metrics)
}
