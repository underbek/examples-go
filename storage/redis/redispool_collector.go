package redis

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// pgxPoolCollector is a Prometheus collector for pgx metrics.
// It implements the prometheus.Collector interface.
type redisPoolCollector struct {
	rCli *redis.Client

	totalConns prometheus.Gauge
	idleConns  prometheus.Gauge

	callExecutedCounter   *prometheus.CounterVec
	callExecutedHistogram *prometheus.HistogramVec
}

func newRedisPoolCollector(cli *redis.Client) *redisPoolCollector {
	dbName := cli.String()
	buckets := make([]float64, 0)
	buckets = append(buckets, prometheus.DefBuckets...)
	buckets = append(buckets, 20, 40, 60, 80, 100, 120)

	newGauge := func(name, help string) prometheus.Gauge {
		return prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "redis",
				Subsystem:   "pool",
				Name:        name,
				Help:        help,
				ConstLabels: prometheus.Labels{"db": dbName},
			},
		)
	}

	return &redisPoolCollector{
		rCli: cli,
		totalConns: newGauge(
			"total_connections",
			"Number of connections currently in the process of being acquired",
		),
		idleConns: newGauge(
			"idle_connections",
			"Number of connections currently in the process of being acquired",
		),
		callExecutedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "redis",
				Subsystem:   "pool",
				Name:        "call_executed_total",
				Help:        "Total number of calls called, regardless of success or error.",
				ConstLabels: prometheus.Labels{"db": dbName},
			},
			[]string{"method", "error"},
		),
		callExecutedHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   "redis",
				Subsystem:   "pool",
				Name:        "call_executed_seconds",
				Help:        "Histogram of transaction latency (seconds) of db that had been application-level handled by the redis connection.",
				ConstLabels: prometheus.Labels{"db": dbName},
				Buckets:     buckets,
			},
			[]string{"method"},
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (p redisPoolCollector) Describe(descs chan<- *prometheus.Desc) {
	p.totalConns.Describe(descs)
	p.idleConns.Describe(descs)

	p.callExecutedCounter.Describe(descs)
	p.callExecutedHistogram.Describe(descs)
}

// Collect implements the prometheus.Collector interface.
func (p redisPoolCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := p.rCli.PoolStats()

	p.totalConns.Set(float64(stats.TotalConns))
	p.idleConns.Set(float64(stats.IdleConns))

	p.totalConns.Collect(metrics)
	p.idleConns.Collect(metrics)

	p.callExecutedCounter.Collect(metrics)
	p.callExecutedHistogram.Collect(metrics)
}
