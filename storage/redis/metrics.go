package redis

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type DBCollector struct {
	Storage
	collector *redisPoolCollector
}

func WithMetrics() Option {
	return func(cli *redis.Client, st Storage) Storage {
		collector := newRedisPoolCollector(cli)

		prometheus.MustRegister(collector)

		return &DBCollector{
			collector: collector,
			Storage:   st,
		}
	}
}

func collectMetrics(c *redisPoolCollector, start time.Time, method string, err error) {
	c.callExecutedHistogram.WithLabelValues(method).Observe(time.Since(start).Seconds())

	var errLabel = "success"
	if err != nil {
		errLabel = err.Error()
	}

	c.callExecutedCounter.WithLabelValues(method, errLabel).Inc()
}
