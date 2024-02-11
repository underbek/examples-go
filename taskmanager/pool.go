package taskmanager

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type ResultTaskData struct {
	Meta      TaskMeta
	Condition ConditionType
	Err       error
	Result    interface{}
}

type Pool struct {
	creator    TaskCreator
	publisher  Publisher
	subscriber Subscriber
	manager    *FlowManager
}

func NewPool(creator TaskCreator, publisher Publisher, subscriber Subscriber, manager *FlowManager) *Pool {
	return &Pool{
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
		return ctx.Err()
	})

	for i := 0; i < size; i++ {
		group.Go(func() error {
			for {
				meta, err := p.subscriber.Subscribe(ctx)
				if err != nil {
					//log
					return err
				}

				results, _, err := p.runTask(ctx, meta)
				if err != nil {
					//log

					p.subscriber.NoAck(ctx, meta)
					return err
				}

				p.subscriber.Ack(ctx, meta)

				err = p.publisher.Publish(ctx, results...)
				if err != nil {
					//log

					return err
				}
			}
		})
	}

	return group.Wait()
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
				// log error
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
						// log error
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
	task, err := p.creator.Create(meta)
	if err != nil {
		//log
		return nil, ResultTaskData{}, err
	}

	flow, err := p.manager.GetFlow(meta.FlowID)
	if err != nil {
		//log
		return nil, ResultTaskData{}, err
	}

	condition := SuccessCondition

	result, err := task.Run(ctx, meta)
	if err != nil {
		meta.FailCount++
		condition = FailCondition

		if meta.FailCount < meta.RetryCount {
			// log retry
			return []TaskMeta{meta},
				ResultTaskData{
					Meta:      meta,
					Condition: condition,
					Err:       err,
					Result:    result,
				},
				nil
		}

		// log fail count > retry count
	}

	newTasks := flow.GetTasks(meta.TaskID, condition)
	for i := range newTasks {
		newTasks[i].PreviousResult = result
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
