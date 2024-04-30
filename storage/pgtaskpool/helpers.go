package pgtaskpool

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

func (w *worker) collectTasksMetrics() (uint32, uint32) {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.MetricsCollectTimeout)
	defer cancel()

	totalCount, err := w.countTasks(ctx, tasksFilter{
		status: statusProcess,
	})
	if err != nil {
		w.logger.
			WithCtx(ctx).
			WithError(err).
			Error("failed to count processing tasks")
	}

	overdueCount, err := w.countTasks(ctx, tasksFilter{
		status:        statusProcess,
		processBefore: time.Now(),
	})
	if err != nil {
		w.logger.
			WithCtx(ctx).
			WithError(err).
			Error("failed to count overdue processing tasks")
	}

	return totalCount, overdueCount
}

func (w *worker) checkTask(task *Task) error {
	if task.Type == ScheduleTypeCustom && len(task.CustomScheduleSlice) == 0 {
		return fmt.Errorf("declared custom schedule type but schedule slice is empty")
	}

	if task.Type == "" {
		task.Type = ScheduleTypeDefault
	}

	return nil
}

func taskAttempts(t Task) []uint {
	if len(t.CustomScheduleSlice) > 0 || t.Type == ScheduleTypeCustom {
		return t.CustomScheduleSlice
	}

	defaultTask := new(Task)
	defaultTask.GenerateDefaultSchedule(defaultLifetime)

	return defaultTask.CustomScheduleSlice
}

func (w *worker) recoverHandler(ctx context.Context, handler handler, task Task) error {
	defer func() {
		if r := recover(); r != nil {
			w.logger.WithCtx(ctx).
				With("task", task).
				With("panic", r).
				With("trace", string(debug.Stack())).
				Error(fmt.Sprintf("Recovered from pgtaskpool panic: %v", r))
		}
	}()

	return handler(ctx, task.TransactionID)
}
