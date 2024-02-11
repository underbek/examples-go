package taskmanager

type TaskCreator interface {
	Create(TaskMeta) (Task, error)
}
