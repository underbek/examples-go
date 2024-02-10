package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ExchangeType string

const (
	ExchangeTypeDirect  ExchangeType = "direct"
	ExchangeTypeFanout  ExchangeType = "fanout"
	ExchangeTypeTopic   ExchangeType = "topic"
	ExchangeTypeHeaders ExchangeType = "headers"
)

type QueueDeclare struct {
	Queue      string
	Passive    bool
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	NoWait     bool
	Arguments  amqp.Table
	TTL        *time.Duration
	MessageTTL *time.Duration
	DLX        string
}

type QueueBind struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
	Arguments  amqp.Table
}

type Consume struct {
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool
	NoWait      bool
	Arguments   amqp.Table
}

type ExchangeDeclare struct {
	Exchange   string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  amqp.Table
}

type PublishMessage struct {
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Message    amqp.Publishing
}
