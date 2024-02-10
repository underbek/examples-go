package kafka

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
)

const unknownLabelValue = "unknown"

type consumerMetrics struct {
	latency       *prometheus.HistogramVec
	messagesCount *prometheus.CounterVec
	errorsCount   *prometheus.CounterVec
}

func newConsumerMetrics(cfg ConsumerConfig) consumerMetrics {
	m := consumerMetrics{
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "kafka",
				Subsystem: "consumer",
				Name:      "messages_latency",
				ConstLabels: map[string]string{
					"group_id": cfg.GroupID,
				},
			},
			[]string{"topic", "partition"},
		),
		messagesCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "kafka",
				Subsystem: "consumer",
				Name:      "messages_count",
				ConstLabels: map[string]string{
					"group_id": cfg.GroupID,
				},
			},
			[]string{"topic", "partition"},
		),
		errorsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "kafka",
				Subsystem: "consumer",
				Name:      "errors_count",
				ConstLabels: map[string]string{
					"group_id": cfg.GroupID,
				},
			},
			[]string{"topic", "partition", "source"},
		),
	}

	if cfg.EnableMetrics {
		prometheus.MustRegister(m.latency)
		prometheus.MustRegister(m.messagesCount)
		prometheus.MustRegister(m.errorsCount)
	}

	return m
}

func (m consumerMetrics) observeLatency(msg kafka.Message) consumerMetrics {
	if !msg.Time.IsZero() {
		m.latency.WithLabelValues(m.getMsgLabels(msg)...).Observe(time.Since(msg.Time).Seconds())
	}
	return m
}

func (m consumerMetrics) incMessage(msg kafka.Message) consumerMetrics {
	m.messagesCount.WithLabelValues(m.getMsgLabels(msg)...).Inc()
	return m
}

func (m consumerMetrics) incError(sourceLabel string, msg kafka.Message) consumerMetrics {
	m.errorsCount.WithLabelValues(append(m.getMsgLabels(msg), sourceLabel)...).Inc()
	return m
}

func (m consumerMetrics) getMsgLabels(msg kafka.Message) []string {
	topic := msg.Topic
	if topic == "" {
		topic = unknownLabelValue
	}

	partition := unknownLabelValue
	if msg.Partition != 0 {
		partition = strconv.FormatInt(int64(msg.Partition), 10)
	}

	return []string{topic, partition}
}
