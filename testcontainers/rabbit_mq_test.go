package testcontainer

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

type TestRabbitMQSuite struct {
	suite.Suite
	container *RabbitMQContainer
}

func (s *TestRabbitMQSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	var err error
	s.container, err = NewRabbitMQContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestRabbitMQSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.Require().NoError(s.container.Terminate(ctx))
}

func TestSuiteRabbitMQ_Run(t *testing.T) {
	suite.Run(t, new(TestRabbitMQSuite))
}

func (s *TestRabbitMQSuite) Test_RabbitMQConn() {
	conn, err := amqp.Dial(s.container.GetDSN())
	s.Require().NoError(err)
	defer conn.Close()

	ch, err := conn.Channel()
	s.Require().NoError(err)
	defer ch.Close()

	args := make(amqp.Table)
	args["x-delayed-type"] = "direct"
	err = ch.ExchangeDeclare("delayed", "x-delayed-message", true, false, false, false, args)
	s.Require().NoError(err)

	delay := 2000 // 2 seconds
	headers := make(amqp.Table)
	headers["x-delay"] = delay

	err = ch.PublishWithContext(
		context.Background(),
		"delayed",
		"user.event.publish",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         []byte("delay"),
			Headers:      headers,
		},
	)
	s.Require().NoError(err)

	delayedQueue, err := ch.QueueDeclare("user-published-queue", true, false, false, false, nil)
	s.Require().NoError(err)

	err = ch.QueueBind(delayedQueue.Name, "user.event.publish", "delayed", false, nil)
	s.Require().NoError(err)

	ds, err := ch.Consume("user-published-queue", "user-published-consumer", false, false, false, false, nil)
	s.Require().NoError(err)

	start := time.Now()

	_, ok := <-ds
	s.Require().True(ok)

	elapsed := time.Since(start)
	s.Require().True(elapsed >= time.Duration(delay))
}
