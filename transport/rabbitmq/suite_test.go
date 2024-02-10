package rabbitmq

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/testcontainers"
)

type TestSuite struct {
	suite.Suite
	container *testcontainer.RabbitMQContainer
	conn      *amqp.Connection
}

func (s *TestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	var err error
	s.container, err = testcontainer.NewRabbitMQContainer(ctx)
	s.Require().NoError(err)

	s.conn, err = amqp.Dial(s.container.GetDSN())
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.Require().NoError(s.conn.Close())
	s.Require().NoError(s.container.Terminate(ctx))
}

func TestSuiteRabbitMQ_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) getChannel() (*amqp.Channel, func()) {
	ch, err := s.conn.Channel()
	s.Require().NoError(err)
	return ch, func() { s.Require().NoError(ch.Close()) }
}

func (s *TestSuite) checkQueue(ch *amqp.Channel, queueName string, messagesCount int) {
	s.Require().Eventually(func() bool {
		q, err := ch.QueueInspect(queueName)
		s.Require().NoError(err)
		return q.Messages == messagesCount
	}, time.Second*10, time.Millisecond*100)
}
