package pgtaskpool

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/underbek/examples-go/metrics"
)

const (
	namespace = "pgtaskpool"
	subsystem = "task_scheduler"

	//labels
	metricsTaskType   = "task_type"
	metricsTaskStatus = "task_status"
)

type counter interface {
	collectTasksMetrics() (uint32, uint32)
}

type taskSchedulerMetrics struct {
	tasksCount    *prometheus.CounterVec
	tasksDuration *prometheus.HistogramVec
}

type counterMetrics struct {
	counter counter

	processingTasksCount prometheus.Gauge
	overdueTasksCount    prometheus.Gauge
}

func newSchedulerMetrics(enable bool) taskSchedulerMetrics {
	m := taskSchedulerMetrics{
		tasksCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tasks_count",
				Help:      "Total number of running tasks",
			},
			[]string{metricsTaskType, metricsTaskStatus},
		),
		tasksDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tasks_duration_seconds",
				Help:      "Task completion time histogram (seconds)",
				Buckets:   metrics.DefBuckets,
			},
			[]string{metricsTaskType, metricsTaskStatus},
		),
	}

	if enable {
		prometheus.MustRegister(m.tasksCount)
		prometheus.MustRegister(m.tasksDuration)
	}

	return m
}

func (m taskSchedulerMetrics) incTask(task Task) taskSchedulerMetrics {
	m.tasksCount.With(
		m.labels(task),
	).Inc()
	return m
}

func (m taskSchedulerMetrics) registerTaskDuration(task Task, duration float64) taskSchedulerMetrics {
	m.tasksDuration.With(
		m.labels(task),
	).Observe(duration)
	return m
}

func (m taskSchedulerMetrics) labels(task Task) prometheus.Labels {
	status := statusProcess

	if task.Status != nil {
		status = *task.Status
	}

	labels := prometheus.Labels{
		metricsTaskType:   task.Type.String(),
		metricsTaskStatus: status,
	}

	return labels
}

func newCounterMetrics(counter counter, enable bool) {
	m := counterMetrics{
		counter: counter,
		processingTasksCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tasks_processing_count",
				Help:      "Number of processing tasks",
			},
		),
		overdueTasksCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tasks_overdue_count",
				Help:      "Number of overdue processing tasks",
			},
		),
	}

	if enable {
		prometheus.MustRegister(m)
	}
}

func (c counterMetrics) Describe(descs chan<- *prometheus.Desc) {
	c.processingTasksCount.Describe(descs)
	c.overdueTasksCount.Describe(descs)
}

func (c counterMetrics) Collect(metrics chan<- prometheus.Metric) {
	totalCount, overdueCount := c.counter.collectTasksMetrics()

	c.processingTasksCount.Set(float64(totalCount))
	c.overdueTasksCount.Set(float64(overdueCount))

	c.processingTasksCount.Collect(metrics)
	c.overdueTasksCount.Collect(metrics)
}
