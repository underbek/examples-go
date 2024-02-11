package taskmanager

import "fmt"

type FlowManager struct {
	flows map[string]*Flow
}

func NewFlowManager() *FlowManager {
	return &FlowManager{
		flows: make(map[string]*Flow),
	}
}

func (m *FlowManager) AddFlow(flows ...*Flow) {
	for _, flow := range flows {
		m.flows[flow.flowID] = flow
	}
}

func (m *FlowManager) GetFlow(flowID string) (*Flow, error) {
	flow, ok := m.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow with id %s is not exists", flowID)
	}

	return flow, nil
}
