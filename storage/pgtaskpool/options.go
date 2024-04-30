package pgtaskpool

import "time"

type options struct {
	MaxAttempts           int
	ProcessStatus         string
	FailStatus            string
	SuccessStatus         string
	CanceledStatus        string
	Table                 string
	EnableMetrics         bool
	MetricsCollectTimeout time.Duration
	BlockTimeout          time.Duration
}

func defaultWorkerOptions() options {
	return options{
		MaxAttempts:           10,
		ProcessStatus:         statusProcess,
		FailStatus:            statusFailed,
		SuccessStatus:         statusSuccess,
		CanceledStatus:        statusCancelled,
		Table:                 defaultTaskTable,
		EnableMetrics:         false,
		MetricsCollectTimeout: time.Second * 5,
		BlockTimeout:          time.Minute * 5,
	}
}

type WorkerOption interface {
	apply(*options)
}

type funcWorkerOption struct {
	f func(*options)
}

func (fwo *funcWorkerOption) apply(do *options) {
	fwo.f(do)
}

func newFuncWorkerOption(f func(*options)) *funcWorkerOption {
	return &funcWorkerOption{
		f: f,
	}
}

// WithMaxAttempts Specifies the maximum number of attempts to process a task
func WithMaxAttempts(m int) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.MaxAttempts = m
	})
}

// WithProcessStatus Defines the name of the status for the task being processed, instead of the default one
func WithProcessStatus(s string) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.ProcessStatus = s
	})
}

// WithFailStatus Defines the name of the status for the failed task, instead of the default one
func WithFailStatus(s string) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.FailStatus = s
	})
}

// WithSuccessStatus Defines the name of the status for the successed task, instead of the default one
func WithSuccessStatus(s string) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.SuccessStatus = s
	})
}

// WithCanceledStatus Defines the name of the status for the canceled task, instead of the default one
func WithCanceledStatus(s string) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.CanceledStatus = s
	})
}

// WithTable Defines the name of the table for storing task information, instead of the default
func WithTable(t string) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.Table = t
	})
}

// WithMetrics Option to include metrics in Prometheus
func WithMetrics() WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.EnableMetrics = true
	})
}

// WithMetricsRequestTimeout Option to add context timeout inside collector on db requests
func WithMetricsRequestTimeout(dur time.Duration) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.MetricsCollectTimeout = dur
	})
}

// WithBlockTimeout Option to determine the duration of blocking work on one task by one service instance
func WithBlockTimeout(dur time.Duration) WorkerOption {
	return newFuncWorkerOption(func(o *options) {
		o.BlockTimeout = dur
	})
}
