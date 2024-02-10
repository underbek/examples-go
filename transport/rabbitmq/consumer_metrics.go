package rabbitmq

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rabbitmq/amqp091-go"
)

type metrics struct {
	latency *prometheus.HistogramVec
}

func newConsumerMetrics(queue string, enabled bool) metrics {
	m := metrics{
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "rabbitmq",
				Subsystem: "consumer",
				Name:      "messages_latency",
				ConstLabels: map[string]string{
					"queue": queue,
				},
			},
			[]string{"consumer_tag"},
		),
	}

	if enabled {
		prometheus.MustRegister(m.latency)
	}

	return m
}

func (m metrics) observeLatency(msg amqp091.Delivery) {
	if !msg.Timestamp.IsZero() {
		m.latency.WithLabelValues(msg.ConsumerTag).Observe(time.Since(msg.Timestamp).Seconds())
	}
}
