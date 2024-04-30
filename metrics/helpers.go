package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/underbek/examples-go/logger"
)

func RegisterMetrics(logger *logger.Logger, cs ...prometheus.Collector) {
	for _, metric := range cs {
		if err := prometheus.Register(metric); err != nil {
			logger.WithError(err).Warn("failed to register the metric")
		}
	}
}
