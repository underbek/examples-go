package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
)

type Manager struct {
	logger  *logger.Logger
	channel Channel
}

func NewManager(l *logger.Logger, ch Channel) *Manager {
	return &Manager{
		logger:  l,
		channel: ch,
	}
}

func (qm *Manager) DeclareQueue(qd QueueDeclare) (amqp091.Queue, error) {
	if qd.Arguments == nil {
		qd.Arguments = make(amqp091.Table)
	}
	if qd.TTL != nil {
		qd.Arguments["x-expires"] = int(qd.TTL.Milliseconds())
	}
	if qd.MessageTTL != nil {
		qd.Arguments["x-message-ttl"] = int(qd.MessageTTL.Milliseconds())
	}
	if qd.DLX != "" {
		qd.Arguments["x-dead-letter-exchange"] = qd.DLX
	}

	q, err := qm.channel.QueueDeclare(
		qd.Queue,
		qd.Durable,
		qd.AutoDelete,
		qd.Exclusive,
		qd.NoWait,
		qd.Arguments,
	)
	if err != nil {
		return q, fmt.Errorf("queue declare: %w", err)
	}

	qm.logger.With("QueueDeclare", qd).Info("queue declared")

	return q, nil
}

func (qm *Manager) DeclareExchange(ed ExchangeDeclare) error {
	if ed.Type == "" {
		ed.Type = ExchangeTypeDirect
	}

	if err := qm.channel.ExchangeDeclare(
		ed.Exchange,
		string(ed.Type),
		ed.Durable,
		ed.AutoDelete,
		ed.Internal,
		ed.NoWait,
		ed.Arguments,
	); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	qm.logger.With("ExchangeDeclare", ed).Info("exchange declared")

	return nil
}

func (qm *Manager) BindQueue(qb QueueBind) error {
	if err := qm.channel.QueueBind(qb.Queue, qb.RoutingKey, qb.Exchange, qb.NoWait, qb.Arguments); err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	qm.logger.With("QueueBind", qb).Info("queue bound")

	return nil
}

func (qm *Manager) DeclareQueueAndExchange(qd QueueDeclare, ed ExchangeDeclare, qb QueueBind) error {
	if err := qm.DeclareExchange(ed); err != nil {
		return err
	}

	if _, err := qm.DeclareQueue(qd); err != nil {
		return err
	}

	qb.Queue = qd.Queue
	qb.Exchange = ed.Exchange
	return qm.BindQueue(qb)
}
