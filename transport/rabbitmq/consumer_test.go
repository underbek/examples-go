package rabbitmq

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
)

func TestConsumer_Consume(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name              string
		handlerWithCancel func(context.CancelFunc) HandleFunc
		channelProvider   func() *ChannelMock
		wantErr           string
	}{
		{
			name: "Error when channel call",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("Consume", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("some error")).
					Once()
				return c
			},
			wantErr: "consume message: some error",
		},
		{
			name: "Channel is closed",
			channelProvider: func() *ChannelMock {
				ch := make(chan amqp.Delivery)
				close(ch)

				c := NewChannelMock(t)
				c.On("Consume", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return((<-chan amqp.Delivery)(ch), nil).
					Once()
				return c
			},
			wantErr: "delivery channel was closed",
		},
		{
			name: "Handler panics",
			channelProvider: func() *ChannelMock {
				ch := make(chan amqp.Delivery, 1)
				ch <- amqp.Delivery{}

				c := NewChannelMock(t)
				c.On("Consume", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return((<-chan amqp.Delivery)(ch), nil).
					Once()
				return c
			},
			handlerWithCancel: func(cancel context.CancelFunc) func(context.Context, amqp.Delivery) {
				defer cancel()

				return func(context.Context, amqp.Delivery) {
					panic("test")
				}
			},
			wantErr: "context canceled",
		},
		{
			name: "Happy path",
			channelProvider: func() *ChannelMock {
				ch := make(chan amqp.Delivery, 1)
				ch <- amqp.Delivery{
					Body: []byte("test"),
				}

				c := NewChannelMock(t)
				c.On("Consume", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return((<-chan amqp.Delivery)(ch), nil).
					Once()
				return c
			},
			handlerWithCancel: func(cancel context.CancelFunc) func(context.Context, amqp.Delivery) {
				defer cancel()

				return func(_ context.Context, msg amqp.Delivery) {
					require.Equal(t, amqp.Delivery{
						Body: []byte("test"),
					}, msg)
				}
			},
			wantErr: "context canceled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var handler HandleFunc
			if tt.handlerWithCancel != nil {
				handler = tt.handlerWithCancel(cancel)
			}

			require.ErrorContains(
				t,
				NewConsumer(l, tt.channelProvider(), "queue_name", false).Consume(ctx, Consume{}, handler),
				tt.wantErr,
			)
		})
	}
}
