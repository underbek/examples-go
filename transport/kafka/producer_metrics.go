package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/metrics"
)

type producerMetrics struct {
	topic         string
	messagesCount *prometheus.CounterVec
	errorsCount   *prometheus.CounterVec
}

func newProducerMetrics(logger *logger.Logger, cfg ProducerConfig) producerMetrics {
	m := producerMetrics{
		topic: cfg.Topic,
		messagesCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "kafka",
				Subsystem: "producer",
				Name:      "messages_count",
			},
			[]string{"topic"},
		),
		errorsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "kafka",
				Subsystem: "producer",
				Name:      "errors_count",
			},
			[]string{"topic"},
		),
	}

	if cfg.EnableMetrics {
		metrics.RegisterMetrics(logger, m.messagesCount, m.errorsCount)
	}

	return m
}

func (m producerMetrics) incMessage(msg kafka.Message) producerMetrics {
	m.messagesCount.WithLabelValues(m.getMsgLabel(msg)).Inc()
	return m
}

func (m producerMetrics) incError(msg kafka.Message) producerMetrics {
	m.errorsCount.WithLabelValues(m.getMsgLabel(msg)).Inc()
	return m
}

func (m producerMetrics) getMsgLabel(msg kafka.Message) string {
	topic := m.topic
	if topic == "" {
		topic = msg.Topic
	}
	if topic == "" {
		topic = unknownLabelValue
	}

	return topic
}
