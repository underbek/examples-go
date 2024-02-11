package taskmanager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlowManager_GetFlow(t *testing.T) {
	manager := NewFlowManager()
	flow1 := NewFlow("1")
	flow2 := NewFlow("2")

	manager.AddFlow(flow1, flow2)
	flow, err := manager.GetFlow("1")
	assert.NoError(t, err)
	assert.Equal(t, flow1, flow)

	flow, err = manager.GetFlow("2")
	assert.NoError(t, err)
	assert.Equal(t, flow2, flow)

	flow, err = manager.GetFlow("3")
	assert.Errorf(t, err, "flow with id 3 is not exists")
	assert.Nil(t, flow)
}
