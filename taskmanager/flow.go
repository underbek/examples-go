package taskmanager

type ConditionType string

const (
	SuccessCondition ConditionType = "success"
	FailCondition    ConditionType = "failed"
	AllCondition     ConditionType = "all"
)

type taskKey struct {
	taskID        string
	conditionType ConditionType
}

type TaskSetting struct {
	TaskID     string
	RunType    TaskRunType
	RetryCount int
}

type Flow struct {
	flowID string
	tasks  map[taskKey][]TaskSetting
}

func NewFlow(flowID string) *Flow {
	return &Flow{
		flowID: flowID,
		tasks:  make(map[taskKey][]TaskSetting),
	}
}

func (f *Flow) AddCondition(taskID string, conditionType ConditionType, settings ...TaskSetting) {

	if conditionType == AllCondition {
		key := taskKey{
			taskID:        taskID,
			conditionType: SuccessCondition,
		}
		f.tasks[key] = append(f.tasks[key], settings...)

		key = taskKey{
			taskID:        taskID,
			conditionType: FailCondition,
		}
		f.tasks[key] = append(f.tasks[key], settings...)

		return
	}
	key := taskKey{
		taskID:        taskID,
		conditionType: conditionType,
	}

	f.tasks[key] = append(f.tasks[key], settings...)
}

func (f *Flow) GetTasks(taskID string, conditionType ConditionType) []TaskMeta {
	settings := f.tasks[taskKey{
		taskID:        taskID,
		conditionType: conditionType,
	}]

	if len(settings) == 0 {
		return nil
	}

	metas := make([]TaskMeta, 0, len(settings))
	for _, setting := range settings {
		metas = append(metas, TaskMeta{
			TaskID:     setting.TaskID,
			FlowID:     f.flowID,
			RunType:    setting.RunType,
			RetryCount: setting.RetryCount,
		})
	}

	return metas
}
