package taskmanager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlow_Empty(t *testing.T) {
	flow := NewFlow("test")
	metas := flow.GetTasks("test", SuccessCondition)
	assert.Empty(t, metas)
}

func TestFlow_GetTasks(t *testing.T) {
	flow := NewFlow("test")

	task1 := TaskSetting{
		TaskID:  "1",
		RunType: SyncTask,
	}
	meta1 := TaskMeta{
		FlowID:  "test",
		TaskID:  "1",
		RunType: SyncTask,
	}

	task2 := TaskSetting{
		TaskID:  "2",
		RunType: AsyncTask,
	}
	meta2 := TaskMeta{
		FlowID:  "test",
		TaskID:  "2",
		RunType: AsyncTask,
	}

	task3 := TaskSetting{
		TaskID:     "3",
		RunType:    AsyncTask,
		RetryCount: 2,
	}
	meta3 := TaskMeta{
		FlowID:     "test",
		TaskID:     "3",
		RunType:    AsyncTask,
		RetryCount: 2,
	}

	flow.AddCondition("test", SuccessCondition, task1)
	flow.AddCondition("test", SuccessCondition, task2)
	flow.AddCondition("2", FailCondition, task3)

	metas := flow.GetTasks("test", SuccessCondition)
	assert.Len(t, metas, 2)
	assert.EqualValues(t, metas, []TaskMeta{meta1, meta2})

	metas = flow.GetTasks("2", SuccessCondition)
	assert.Empty(t, metas)

	metas = flow.GetTasks("2", FailCondition)
	assert.EqualValues(t, metas, []TaskMeta{meta3})
}
