package pgtaskpool

import (
	"context"
	"errors"
	"github.com/underbek/examples-go/utils"
	"time"

	"github.com/jackc/pgx/v5"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/logger"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
	"github.com/underbek/examples-go/tracing"
	"golang.org/x/sync/errgroup"
)

type (
	handler func(context.Context, uint64) error

	Worker interface {
		Create(context.Context, Task) error
		Cancel(context.Context, uint64) error
		Run(context.Context, time.Duration, handler) error
		Reset(context.Context, Task) error
	}

	worker struct {
		logger  *logger.Logger
		storage goKitPgx.Storage
		config  options
		metrics taskSchedulerMetrics
	}
)

func New(logger *logger.Logger, storage goKitPgx.Storage, options ...WorkerOption) Worker {
	worker := &worker{
		logger:  logger,
		storage: storage,
		config:  defaultWorkerOptions(),
	}

	for _, opt := range options {
		opt.apply(&worker.config)
	}

	worker.metrics = newSchedulerMetrics(worker.config.EnableMetrics)
	newCounterMetrics(worker, worker.config.EnableMetrics)

	return worker
}

func (w *worker) Create(ctx context.Context, task Task) error {
	isExist, err := w.isTaskExists(ctx, task.TransactionID)
	if err != nil {
		return err
	}

	if isExist {
		w.logger.
			WithCtx(ctx).
			With("transaction_id", task.TransactionID).
			With("type", task.Type).
			Warn("task already exists")
		return nil
	}

	if err = w.checkTask(&task); err != nil {
		return err
	}

	return w.createTask(ctx, task)
}

func (w *worker) Cancel(ctx context.Context, id uint64) error {
	return w.cancelTask(ctx, id)
}

func (w *worker) Reset(ctx context.Context, task Task) error {
	tx, err := w.storage.Begin(ctx, nil)
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeInternal, "failed to begin transaction")
	}

	defer func() {
		rCtx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)

		if err = tx.Rollback(rCtx); err != nil {
			w.logger.
				WithCtx(rCtx).
				WithError(err).
				Error("rollback failed")
		}

		cancel()
	}()

	if err = w.cancelTask(ctx, task.TransactionID); err != nil {
		return err
	}

	if err = w.checkTask(&task); err != nil {
		return err
	}

	if err = w.createTask(ctx, task); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to commit transaction")
	}

	return nil
}

func (w *worker) Run(ctx context.Context, syncInterval time.Duration, handler handler) error {
	gr, ctx := errgroup.WithContext(ctx)

	gr.Go(func() error {
		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				err := w.execute(ctx, handler)
				if err != nil {
					w.logger.
						WithCtx(ctx).
						WithError(err).
						Error("execute failed")
				}
			}
		}
	})

	return gr.Wait()
}

func (w *worker) execute(ctx context.Context, handler handler) error {
	task, err := w.getTask(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.logger.
				WithCtx(ctx).
				Debug("no tasks")
			return nil
		}

		w.logger.
			WithCtx(ctx).
			WithError(err).
			Error("getTask returned error")
		return err
	}

	task, err = w.processTaskLock(ctx, task)
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeInternal, "failed to processTaskLock")
	}

	//The task processing itself
	task = w.process(ctx, task, handler)

	//Results of the task execution is being saved
	if err = w.updateTask(ctx, task); err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to update task")
	}

	//Releases the task and updates to further processes
	if err = w.updateTask(ctx, task); err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to unblock task")
	}

	return nil
}

func (w *worker) processTaskLock(ctx context.Context, task Task) (Task, error) {
	// Preparing for block task for other instances
	tx, err := w.storage.Begin(ctx, nil)
	if err != nil {
		return task, ctxerrors.Wrap(err, ctxerrors.TypeInternal, "failed to begin transaction")
	}

	defer func() {
		rCtx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)

		if err = tx.Rollback(rCtx); err != nil {
			w.logger.
				WithCtx(rCtx).
				WithError(err).
				Error("rollback failed")
		}

		cancel()
	}()

	// Block task for specific timeout
	lock := time.Now().Add(w.config.BlockTimeout)
	task.LockTime = &lock

	if err = w.blockTask(ctx, tx, task); err != nil {
		return task, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to block task")
	}

	if err = tx.Commit(ctx); err != nil {
		return task, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to commit transaction")
	}

	return task, nil
}

func (w *worker) process(ctx context.Context, task Task, handler handler) Task {
	var err error
	if task.Trace != nil {
		ctx = tracing.PutStringTraceInfoIntoContext(ctx, task.Trace.TraceID, task.Trace.SpanID)

		if task.Trace.Meta != nil {
			ctx = logger.AddCtxMetaValues(ctx, task.Trace.Meta)
		}
	}

	ctx, span := tracing.StartSpan(ctx, "pgtaskpool", "Execute task")
	defer span.End()

	w.logger.WithCtx(ctx).With("task", task).Debug("executing task")

	task.Attempts++

	startTime := time.Now()

	err = w.recoverHandler(ctx, handler, task)

	w.logger.
		WithCtx(ctx).
		With("task", task).
		WithError(err).
		Debug("task executed")

	if err != nil {
		span.RecordError(err)

		code := ctxerrors.ErrorType(err)
		if !errors.Is(err, ErrNeedJustToRetry) {
			task.LastErrorCode = utils.ToPtr(uint64(code))
			task.LastErrorMessage = utils.ToPtr(err.Error())
		}

		attempts := taskAttempts(task)
		if task.Attempts >= len(attempts) {
			task.Status = &w.config.FailStatus
		} else {
			nextTime := time.Now().Add(time.Second * time.Duration(attempts[task.Attempts]))
			task.NextProcessingTime = &nextTime
		}
	} else {
		span.AddEvent("Task succeeded")
		task.Status = &w.config.SuccessStatus
	}

	w.metrics.incTask(task)
	w.metrics.registerTaskDuration(task, time.Since(startTime).Seconds())
	return task
}
