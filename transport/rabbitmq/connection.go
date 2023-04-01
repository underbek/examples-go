package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection interface {
	Channel() (*amqp.Channel, error)
	Close() error
}

type connection struct {
	conn *amqp.Connection
}

func NewConnection(url string) (Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	return &connection{
		conn: conn,
	}, nil
}

func (c *connection) Channel() (*amqp.Channel, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	return ch, nil
}

func (c *connection) Close() error {
	return c.conn.Close()
}
