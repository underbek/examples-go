package taskmanager

import (
	"context"
	"errors"
)

type TaskRunType string

const (
	SyncTask  TaskRunType = "sync"
	AsyncTask TaskRunType = "async"
)

var (
	ErrNotImplemented = errors.New("task not implemented")
)

type TaskMeta struct {
	TaskID     string      `json:"task_id"`
	FlowID     string      `json:"flow_id"`
	RunType    TaskRunType `json:"run_type"`
	RetryCount int         `json:"retry_count,omitempty"`

	FailCount      int         `json:"fail_count,omitempty"`
	Additional     interface{} `json:"additional,omitempty"`
	PreviousResult interface{} `json:"previous_result,omitempty"`
	//TODO: trace info
}

type Task interface {
	Run(context.Context, TaskMeta) (interface{}, error)
}
