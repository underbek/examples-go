package taskmanager

import (
	"context"
)

type TaskRunType string

const (
	SyncTask  TaskRunType = "sync"
	AsyncTask TaskRunType = "async"
)

type TaskTrace struct {
	TraceID [16]byte `json:"trace_id"`
	SpanID  [8]byte  `json:"span_id"`
}

type TaskMeta struct {
	TaskID     string      `json:"task_id"`
	FlowID     string      `json:"flow_id"`
	RunType    TaskRunType `json:"run_type"`
	RetryCount int         `json:"retry_count,omitempty"`

	FailCount      int         `json:"fail_count,omitempty"`
	PreviousResult interface{} `json:"previous_result,omitempty"`

	Trace      *TaskTrace  `json:"trace,omitempty"`
	Additional interface{} `json:"additional,omitempty"`
}

type Task interface {
	Run(context.Context, TaskMeta) (interface{}, error)
}
