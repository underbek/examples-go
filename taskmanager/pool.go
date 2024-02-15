package taskmanager

import (
	"context"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type ResultTaskData struct {
	Meta      TaskMeta
	Condition ConditionType
	Err       error
	Result    interface{}
}

type Pool struct {
	logger     *logger.Logger
	creator    TaskCreator
	publisher  Publisher
	subscriber Subscriber
	manager    *FlowManager
}

func NewPool(
	logger *logger.Logger,
	creator TaskCreator,
	publisher Publisher,
	subscriber Subscriber,
	manager *FlowManager,
) *Pool {
	return &Pool{
		logger:     logger.Named("task_manager_pool"),
		creator:    creator,
		publisher:  publisher,
		subscriber: subscriber,
		manager:    manager,
	}
}

func (p *Pool) Run(ctx context.Context, size int) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		<-ctx.Done()
		p.logger.WithCtx(ctx).Info("pool stopping")
		return ctx.Err()
	})

	for i := 0; i < size; i++ {
		group.Go(func() error {
			for {
				meta, err := p.subscriber.Subscribe(ctx)
				if err != nil {
					p.logger.WithCtx(ctx).WithError(err).Error("failed to subscribe")

					return err
				}

				results, _, err := p.runTask(ctx, meta)
				if err != nil {
					p.logger.WithCtx(ctx).
						WithError(err).
						With("task_meta", meta).
						Error("failed to run task")

					p.subscriber.NoAck(ctx, meta)
					return err
				}

				p.subscriber.Ack(ctx, meta)

				err = p.publisher.Publish(ctx, results...)
				if err != nil {
					p.logger.WithCtx(ctx).
						WithError(err).
						With("task_id", meta.TaskID).
						With("flow_id", meta.FlowID).
						Error("failed to publish")

					return err
				}
			}
		})
	}

	resErr := group.Wait()
	p.logger.WithCtx(ctx).WithError(resErr).Error("pool stopped")

	return resErr
}

func (p *Pool) RunSync(ctx context.Context, head TaskMeta) ([]ResultTaskData, error) {
	metas := make(chan TaskMeta, 10)
	defer close(metas)

	metas <- head

	var results []ResultTaskData
	for {
		select {
		default:
			return results, nil
		case meta := <-metas:
			newMetas, result, err := p.runTask(ctx, meta)
			if err != nil {
				p.logger.WithCtx(ctx).
					WithError(err).
					With("task_id", meta.TaskID).
					With("flow_id", meta.FlowID).
					Error("failed to run sync task")

				return nil, err
			}

			results = append(results, result)

			if len(newMetas) == 0 {
				continue
			}

			for _, newMeta := range newMetas {
				if newMeta.RunType == AsyncTask {
					err = p.publisher.Publish(ctx, newMeta)
					if err != nil {
						p.logger.WithCtx(ctx).
							WithError(err).
							With("task_id", newMeta.TaskID).
							With("flow_id", newMeta.FlowID).
							Error("failed to publish")

						return nil, err
					}

					continue
				}

				metas <- newMeta
			}
		}
	}
}

func (p *Pool) runTask(ctx context.Context, meta TaskMeta) ([]TaskMeta, ResultTaskData, error) {
	if meta.Trace != nil {
		ctx = tracing.PutTraceInfoIntoContext(ctx, meta.Trace.TraceID, meta.Trace.SpanID)
	}

	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindServer, "task_manager", "RunSync")
	defer span.End()

	task, err := p.creator.Create(meta)
	if err != nil {
		p.logger.WithCtx(ctx).
			WithError(err).
			With("task_id", meta.TaskID).
			With("flow_id", meta.FlowID).
			Error("failed to create task")

		return nil, ResultTaskData{}, err
	}

	flow, err := p.manager.GetFlow(meta.FlowID)
	if err != nil {
		p.logger.WithCtx(ctx).
			WithError(err).
			With("flow_id", meta.FlowID).
			Error("failed to get flow")

		return nil, ResultTaskData{}, err
	}

	condition := SuccessCondition

	p.logger.WithCtx(ctx).
		With("task_id", meta.TaskID).
		With("flow_id", meta.FlowID).
		Debug("run task")

	result, err := task.Run(ctx, meta)
	if err != nil {
		meta.FailCount++
		condition = FailCondition

		if meta.FailCount < meta.RetryCount {
			p.logger.WithCtx(ctx).
				WithError(err).
				With("task_id", meta.TaskID).
				With("flow_id", meta.FlowID).
				With("fail_count", meta.FailCount).
				With("retry_count", meta.RetryCount).
				Info("retry task after fail")

			return []TaskMeta{meta},
				ResultTaskData{
					Meta:      meta,
					Condition: condition,
					Err:       err,
					Result:    result,
				},
				nil
		}
	}

	p.logger.WithCtx(ctx).
		WithError(err).
		With("task_id", meta.TaskID).
		With("flow_id", meta.FlowID).
		With("fail_count", meta.FailCount).
		With("retry_count", meta.RetryCount).
		Error("task failed")

	newTasks := flow.GetTasks(meta.TaskID, condition)
	for i := range newTasks {
		newTasks[i].PreviousResult = result

		if span.IsRecording() {
			traceData := &TaskTrace{
				TraceID: tracing.GetSpanContextFromContext(ctx).TraceID(),
				SpanID:  tracing.GetSpanContextFromContext(ctx).SpanID(),
			}
			newTasks[i].Trace = traceData
		}
	}

	return newTasks,
		ResultTaskData{
			Meta:      meta,
			Condition: condition,
			Err:       err,
			Result:    result,
		},
		nil
}
