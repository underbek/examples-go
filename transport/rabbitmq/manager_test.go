package rabbitmq

import (
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/utils"
)

func TestManager_DeclareQueue(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		qd              QueueDeclare
		channelProvider func() *ChannelMock
		want            amqp.Queue
		wantErr         string
	}{
		{
			name: "Empty config",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueDeclare", "", false, false, false, false, make(amqp.Table)).
					Return(amqp.Queue{}, nil).
					Once()
				return c
			},
		},
		{
			name: "Error when channel call",
			qd: QueueDeclare{
				Queue: "queue_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(amqp.Queue{}, errors.New("some error")).
					Once()
				return c
			},
			wantErr: "queue declare: some error",
		},
		{
			name: "Happy path",
			qd: QueueDeclare{
				Queue:      "queue_name",
				Passive:    true,
				Durable:    true,
				Exclusive:  true,
				AutoDelete: true,
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueDeclare", "queue_name", true, true, true, true, amqp.Table{
					"key1": "arg1",
					"key2": 123,
				}).
					Return(amqp.Queue{
						Name:      "queue_name",
						Messages:  2,
						Consumers: 1,
					}, nil).
					Once()
				return c
			},
			want: amqp.Queue{
				Name:      "queue_name",
				Messages:  2,
				Consumers: 1,
			},
		},
		{
			name: "Happy path with extra properties",
			qd: QueueDeclare{
				Queue:      "queue_name",
				TTL:        utils.ToPtr(time.Minute * 2),
				MessageTTL: utils.ToPtr[time.Duration](0),
				DLX:        "dlx_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueDeclare", "queue_name", false, false, false, false, amqp.Table{
					"x-expires":              2 * 60 * 1000,
					"x-message-ttl":          0,
					"x-dead-letter-exchange": "dlx_name",
				}).
					Return(amqp.Queue{
						Name:      "queue_name",
						Messages:  2,
						Consumers: 1,
					}, nil).
					Once()
				return c
			},
			want: amqp.Queue{
				Name:      "queue_name",
				Messages:  2,
				Consumers: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewManager(l, tt.channelProvider()).DeclareQueue(tt.qd)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestManager_DeclareExchange(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		ed              ExchangeDeclare
		channelProvider func() *ChannelMock
		wantErr         string
	}{
		{
			name: "Empty config",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", "", "direct", false, false, false, false, amqp.Table(nil)).
					Return(nil).
					Once()
				return c
			},
		},
		{
			name: "Error when channel call",
			ed: ExchangeDeclare{
				Exchange: "exchange_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some error")).
					Once()
				return c
			},
			wantErr: "exchange declare: some error",
		},
		{
			name: "Happy path",
			ed: ExchangeDeclare{
				Exchange:   "exchange_name",
				Type:       ExchangeTypeFanout,
				Durable:    true,
				AutoDelete: true,
				Internal:   true,
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", "exchange_name", "fanout", true, true, true, true, amqp.Table{
					"key1": "arg1",
					"key2": 123,
				}).
					Return(nil).
					Once()
				return c
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewManager(l, tt.channelProvider()).DeclareExchange(tt.ed)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestManager_BindQueue(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		qb              QueueBind
		channelProvider func() *ChannelMock
		wantErr         string
	}{
		{
			name: "Empty config",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueBind", "", "", "", false, amqp.Table(nil)).
					Return(nil).
					Once()
				return c
			},
		},
		{
			name: "Error when channel call",
			qb: QueueBind{
				Queue: "queue_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueBind", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some error")).
					Once()
				return c
			},
			wantErr: "queue bind: some error",
		},
		{
			name: "Happy path",
			qb: QueueBind{
				Queue:      "queue_name",
				Exchange:   "exchange_name",
				RoutingKey: "routing_key",
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("QueueBind", "queue_name", "routing_key", "exchange_name", true, amqp.Table{
					"key1": "arg1",
					"key2": 123,
				}).
					Return(nil).
					Once()
				return c
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewManager(l, tt.channelProvider()).BindQueue(tt.qb)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestManager_DeclareQueueAndExchange(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		qd              QueueDeclare
		ed              ExchangeDeclare
		qb              QueueBind
		channelProvider func() *ChannelMock
		wantErr         string
	}{
		{
			name: "Empty config",
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", "", "direct", false, false, false, false, amqp.Table(nil)).
					Return(nil).
					Once()
				c.On("QueueDeclare", "", false, false, false, false, make(amqp.Table)).
					Return(amqp.Queue{}, nil).
					Once()
				c.On("QueueBind", "", "", "", false, amqp.Table(nil)).
					Return(nil).
					Once()
				return c
			},
		},
		{
			name: "Error when channel call for exchange declare",
			qd: QueueDeclare{
				Queue: "queue_name",
			},
			ed: ExchangeDeclare{
				Exchange: "exchange_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some exchange error")).
					Once()
				return c
			},
			wantErr: "exchange declare: some exchange error",
		},
		{
			name: "Error when channel call for queue declare",
			qd: QueueDeclare{
				Queue: "queue_name",
			},
			ed: ExchangeDeclare{
				Exchange: "exchange_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				c.On("QueueDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(amqp.Queue{}, errors.New("some queue error")).
					Once()
				return c
			},
			wantErr: "queue declare: some queue error",
		},
		{
			name: "Error when channel call for bind",
			qd: QueueDeclare{
				Queue: "queue_name",
			},
			ed: ExchangeDeclare{
				Exchange: "exchange_name",
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				c.On("QueueDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(amqp.Queue{}, nil).
					Once()
				c.On("QueueBind", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some binding error")).
					Once()
				return c
			},
			wantErr: "queue bind: some binding error",
		},
		{
			name: "Happy path",
			qd: QueueDeclare{
				Queue:      "queue_name",
				Passive:    true,
				Durable:    true,
				Exclusive:  true,
				AutoDelete: true,
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
				TTL:        utils.ToPtr(time.Minute * 2),
				MessageTTL: utils.ToPtr[time.Duration](0),
				DLX:        "dlx_name",
			},
			ed: ExchangeDeclare{
				Exchange:   "exchange_name",
				Type:       ExchangeTypeFanout,
				Durable:    true,
				AutoDelete: true,
				Internal:   true,
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
			},
			qb: QueueBind{
				RoutingKey: "routing_key",
				NoWait:     true,
				Arguments: amqp.Table{
					"key1": "arg1",
					"key2": 123,
				},
			},
			channelProvider: func() *ChannelMock {
				c := NewChannelMock(t)
				c.On("ExchangeDeclare", "exchange_name", "fanout", true, true, true, true, amqp.Table{
					"key1": "arg1",
					"key2": 123,
				}).
					Return(nil).
					Once()
				c.On("QueueDeclare", "queue_name", true, true, true, true, amqp.Table{
					"key1":                   "arg1",
					"key2":                   123,
					"x-expires":              2 * 60 * 1000,
					"x-message-ttl":          0,
					"x-dead-letter-exchange": "dlx_name",
				}).
					Return(amqp.Queue{
						Name:      "queue_name",
						Messages:  2,
						Consumers: 1,
					}, nil).
					Once()
				c.On("QueueBind", "queue_name", "routing_key", "exchange_name", true, amqp.Table{
					"key1": "arg1",
					"key2": 123,
				}).
					Return(nil).
					Once()
				return c
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewManager(l, tt.channelProvider()).DeclareQueueAndExchange(tt.qd, tt.ed, tt.qb)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
