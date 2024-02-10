package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
)

func TestProducer_Publish(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		ctxProvider     func() context.Context
		exchange        string
		msg             PublishMessage
		channelProvider func() *ChannelMock
		wantErr         string
	}{
		{
			name: "Empty config",
			msg:  PublishMessage{},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("PublishWithContext", mock.Anything, "", "", false, false, mock.MatchedBy(func(msg amqp.Publishing) bool {
					require.WithinDuration(t, time.Now(), msg.Timestamp, time.Second)

					msg.Timestamp = time.Time{}
					require.Equal(t, amqp.Publishing{
						Headers: amqp.Table{},
					}, msg)
					return true
				})).
					Return(nil).
					Once()
				return c
			},
		},
		{
			name: "Error when channel call",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("PublishWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some error")).
					Once()
				return c
			},
			wantErr: "publish message: some error",
		},
		{
			name: "Happy path with injected headers",
			ctxProvider: func() context.Context {
				return logger.AddCtxMetaValues(context.Background(), map[string]string{
					"key1": "val1",
					"key2": "val2",
				})
			},
			exchange: "exchange_name",
			msg: PublishMessage{
				RoutingKey: "routing_key",
				Mandatory:  true,
				Immediate:  true,
				Message: amqp.Publishing{
					Headers: amqp.Table{
						"key1": "val1",
					},
					Timestamp: time.Date(2023, 10, 16, 11, 0, 0, 0, time.UTC),
				},
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("PublishWithContext", mock.Anything, "exchange_name", "routing_key", true, true, amqp.Publishing{
					Headers: amqp.Table{
						"key1":      "val1",
						logger.Meta: `{"key1":"val1","key2":"val2"}`,
					},
					Timestamp: time.Date(2023, 10, 16, 11, 0, 0, 0, time.UTC),
				}).
					Return(nil).
					Once()
				return c
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctxProvider != nil {
				ctx = tt.ctxProvider()
			}

			err := NewProducer(l, tt.channelProvider(), tt.exchange).Publish(ctx, tt.msg)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
