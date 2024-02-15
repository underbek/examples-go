package taskmanager

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
	"sync"
	"testing"
	"time"
)

var (
	errTest = errors.New("test error")
)

type TestTask struct {
	completeCount *int
	mtx           *sync.Mutex
}

func (t *TestTask) Run(_ context.Context, meta TaskMeta) (interface{}, error) {
	t.mtx.Lock()
	*t.completeCount++
	t.mtx.Unlock()

	result := []string{meta.TaskID}
	if meta.PreviousResult != nil {
		result = append(meta.PreviousResult.([]string), result...)
	}

	return result, nil
}

type TestTaskSuccess struct {
	TestTask
}

type TestTaskFailed struct {
	TestTask
}

func (t *TestTaskFailed) Run(ctx context.Context, meta TaskMeta) (interface{}, error) {
	result, _ := t.TestTask.Run(ctx, meta)
	return result, errTest
}

type TestCreator struct {
	completeCount int
}

func (c *TestCreator) Create(meta TaskMeta) (Task, error) {
	mtx := sync.Mutex{}

	testTask := TestTask{completeCount: &c.completeCount, mtx: &mtx}

	switch meta.TaskID {
	case "head":
		return &TestTaskSuccess{testTask}, nil
	case "1":
		return &TestTaskSuccess{testTask}, nil
	case "2":
		return &TestTaskFailed{testTask}, nil
	case "3":
		return &TestTaskSuccess{testTask}, nil
	case "4":
		return &TestTaskFailed{testTask}, nil
	}

	return nil, errors.New("not implemented")
}

func TestPool_RunAsync(t *testing.T) {
	flow := NewFlow("flow_1")
	headTask := TaskMeta{
		FlowID:  "flow_1",
		TaskID:  "head",
		RunType: AsyncTask,
	}
	task1 := TaskSetting{
		TaskID:  "1",
		RunType: AsyncTask,
	}
	task2 := TaskSetting{
		TaskID:  "2",
		RunType: AsyncTask,
	}
	task3 := TaskSetting{
		TaskID:  "3",
		RunType: AsyncTask,
	}
	task4 := TaskSetting{
		TaskID:     "4",
		RunType:    AsyncTask,
		RetryCount: 3,
	}

	flow.AddCondition("head", SuccessCondition, task1)
	flow.AddCondition("1", SuccessCondition, task2, task3)
	flow.AddCondition("2", FailCondition, task3)
	flow.AddCondition("3", SuccessCondition, task4)

	manager := NewFlowManager()
	manager.AddFlow(flow)

	creator := &TestCreator{}

	transport := NewChannelTransport(10)

	lg, err := logger.New(true)
	require.NoError(t, err)

	pool := NewPool(lg, creator, transport, transport, manager)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var runError error

	go func() {
		runError = pool.Run(ctx, 10)
	}()

	err = transport.Publish(ctx, headTask)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return creator.completeCount == 11 // head-1, task1-1, task2-1, task3-2, task4-6
	}, time.Second*5, time.Millisecond*10)

	transport.Close()

	assert.Eventually(t, func() bool {
		return errors.Is(runError, ErrChannelClosed)
	}, time.Second*5, time.Millisecond*10)
}

func TestPool_RunSync(t *testing.T) {
	flow := NewFlow("flow_1")
	headTask := TaskMeta{
		FlowID:  "flow_1",
		TaskID:  "head",
		RunType: SyncTask,
	}
	task1 := TaskSetting{
		TaskID:  "1",
		RunType: SyncTask,
	}
	task2 := TaskSetting{
		TaskID:  "2",
		RunType: AsyncTask,
	}
	task3 := TaskSetting{
		TaskID:  "3",
		RunType: SyncTask,
	}
	task4 := TaskSetting{
		TaskID:     "4",
		RunType:    SyncTask,
		RetryCount: 3,
	}

	flow.AddCondition("head", SuccessCondition, task1)
	flow.AddCondition("1", SuccessCondition, task2, task3)
	flow.AddCondition("2", FailCondition, task3)
	flow.AddCondition("3", SuccessCondition, task4)

	manager := NewFlowManager()
	manager.AddFlow(flow)

	creator := &TestCreator{}

	transport := NewChannelTransport(10)

	lg, err := logger.New(true)
	require.NoError(t, err)

	pool := NewPool(lg, creator, transport, transport, manager)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var runError error

	go func() {
		runError = pool.Run(ctx, 10)
	}()

	results, err := pool.RunSync(ctx, headTask)
	assert.NoError(t, err)
	require.Len(t, results, 6)

	for i := range results {
		results[i].Meta.Trace = nil
	}

	assert.Equal(t, TaskMeta{
		TaskID:  "head",
		FlowID:  "flow_1",
		RunType: SyncTask,
	}, results[0].Meta)
	assert.Equal(t, SuccessCondition, results[0].Condition)
	assert.Equal(t, []string{"head"}, results[0].Result)

	assert.Equal(t, TaskMeta{
		TaskID:         "1",
		FlowID:         "flow_1",
		RunType:        SyncTask,
		PreviousResult: []string{"head"},
	}, results[1].Meta)
	assert.Equal(t, SuccessCondition, results[1].Condition)
	assert.Equal(t, []string{"head", "1"}, results[1].Result)

	assert.Equal(t, TaskMeta{
		TaskID:         "3",
		FlowID:         "flow_1",
		RunType:        SyncTask,
		PreviousResult: []string{"head", "1"},
	}, results[2].Meta)
	assert.Equal(t, SuccessCondition, results[2].Condition)
	assert.Equal(t, []string{"head", "1", "3"}, results[2].Result)

	assert.Equal(t, TaskMeta{
		TaskID:         "4",
		FlowID:         "flow_1",
		RunType:        SyncTask,
		RetryCount:     3,
		FailCount:      1,
		PreviousResult: []string{"head", "1", "3"},
	}, results[3].Meta)
	assert.Equal(t, FailCondition, results[3].Condition)
	assert.Equal(t, []string{"head", "1", "3", "4"}, results[3].Result)
	assert.Equal(t, errTest, results[3].Err)

	assert.Equal(t, TaskMeta{
		TaskID:         "4",
		FlowID:         "flow_1",
		RunType:        SyncTask,
		RetryCount:     3,
		FailCount:      2,
		PreviousResult: []string{"head", "1", "3"},
	}, results[4].Meta)
	assert.Equal(t, FailCondition, results[4].Condition)
	assert.Equal(t, []string{"head", "1", "3", "4"}, results[4].Result)
	assert.Equal(t, errTest, results[4].Err)

	assert.Equal(t, TaskMeta{
		TaskID:         "4",
		FlowID:         "flow_1",
		RunType:        SyncTask,
		RetryCount:     3,
		FailCount:      3,
		PreviousResult: []string{"head", "1", "3"},
	}, results[5].Meta)
	assert.Equal(t, FailCondition, results[5].Condition)
	assert.Equal(t, []string{"head", "1", "3", "4"}, results[5].Result)
	assert.Equal(t, errTest, results[5].Err)

	assert.Eventually(t, func() bool {
		return creator.completeCount == 11 // head-1, task1-1, task2-1, task3-2, task4-6
	}, time.Second*5, time.Millisecond*10)

	transport.Close()

	assert.Eventually(t, func() bool {
		return errors.Is(runError, ErrChannelClosed)
	}, time.Second*5, time.Millisecond*10)
}
