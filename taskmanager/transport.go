package taskmanager

import (
	"context"
	"errors"
	"sync"
)

var ErrChannelClosed = errors.New("channel closed")

type Publisher interface {
	Publish(context.Context, ...TaskMeta) error
}

type Subscriber interface {
	Subscribe(context.Context) (TaskMeta, error)
	Ack(context.Context, TaskMeta)
	NoAck(context.Context, TaskMeta)
}

type ChannelTransport struct {
	tasks  chan TaskMeta
	closed bool
	mtx    sync.RWMutex
}

func NewChannelTransport(size int) *ChannelTransport {
	return &ChannelTransport{
		tasks: make(chan TaskMeta, size),
	}
}

func (t *ChannelTransport) Publish(_ context.Context, tasks ...TaskMeta) error {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	if t.closed {
		return errors.New("channel closed")
	}

	for _, task := range tasks {
		t.tasks <- task
	}

	return nil
}

func (t *ChannelTransport) Subscribe(_ context.Context) (TaskMeta, error) {
	task, ok := <-t.tasks
	if !ok {
		return TaskMeta{}, ErrChannelClosed
	}

	return task, nil
}

func (t *ChannelTransport) Ack(_ context.Context, _ TaskMeta) {}

func (t *ChannelTransport) NoAck(ctx context.Context, meta TaskMeta) {
	err := t.Publish(ctx, meta)
	if err != nil {
		return
	}
}

func (t *ChannelTransport) Close() {
	t.mtx.Lock()
	t.closed = true
	t.mtx.Unlock()
	close(t.tasks)
}
