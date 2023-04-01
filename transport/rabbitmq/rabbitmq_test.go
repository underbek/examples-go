package rabbitmq

import (
	"context"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/logger"
	testcontainer "github.com/underbek/examples-go/testcontainers"
)

const (
	exchange = "test-exchange"
	queue    = "test-queue"
	key      = "test-key"
)

type TestSuiteRabbitMQ struct {
	suite.Suite
	container *testcontainer.RabbitMQContainer
	conn      Connection
}

func (s *TestSuiteRabbitMQ) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	var err error
	s.container, err = testcontainer.NewRabbitMQContainer(ctx)
	s.Require().NoError(err)

	s.conn, err = NewConnection(s.container.GetDSN())
	s.Require().NoError(err)
}

func (s *TestSuiteRabbitMQ) TearDownSuite() {
	s.Require().NoError(s.conn.Close())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.container.Terminate(ctx)
	s.Require().NoError(err)
}

func TestSuiteRabbitMQ_Run(t *testing.T) {
	suite.Run(t, new(TestSuiteRabbitMQ))
}

func (s *TestSuiteRabbitMQ) Test_RabbitMQ() {
	lg, err := logger.New(true)
	s.Require().NoError(err)

	producer, err := NewProducer(lg, s.conn, ExchangeDeclare{
		Exchange: exchange,
		Type:     ExchangeTypeDirect,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(producer.Close()) }()

	consumer, err := NewConsumer(lg, s.conn,
		QueueDeclare{
			Queue: queue,
		},
		QueueBind{
			Queue:      queue,
			Exchange:   exchange,
			RoutingKey: key,
		},
	)
	s.Require().NoError(err)

	msg := amqp.Publishing{
		Headers: map[string]interface{}{
			"test-header": "test-header",
		},
		ContentType:   "text",
		CorrelationId: "correlation_id",
		MessageId:     "message_id",
		Timestamp:     time.Now().Truncate(time.Second),
		AppId:         "app_id",
		Body:          []byte("test message"),
	}

	err = producer.Publish(context.Background(), PublishMessage{
		RoutingKey: key,
		Message:    msg,
	})
	s.Require().NoError(err)

	channel, err := s.conn.Channel()
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(channel.Close()) }()

	s.Assert().Eventually(func() bool {
		q, inErr := channel.QueueInspect(queue)
		s.Assert().NoError(inErr)

		return 1 == q.Messages
	}, time.Second*10, time.Millisecond*100)

	expected := msg

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = consumer.Consume(context.Background(), Consume{
			ConsumerTag: "test-consumer",
		},
			func(ctx context.Context, msg amqp.Delivery) {
				actual := amqp.Publishing{
					Headers:       msg.Headers,
					ContentType:   msg.ContentType,
					CorrelationId: msg.CorrelationId,
					MessageId:     msg.MessageId,
					Timestamp:     msg.Timestamp,
					AppId:         msg.AppId,
					Body:          msg.Body,
				}

				s.Assert().NoError(msg.Ack(false))

				s.Assert().Equal(expected, actual)

				s.Assert().NoError(consumer.Close())
			})

		s.Assert().ErrorContains(err, "delivery channel was closed")
		wg.Done()
	}()

	wg.Wait()

	s.Assert().Eventually(func() bool {
		q, err := channel.QueueInspect(queue)
		s.Assert().NoError(err)

		return 0 == q.Messages
	}, time.Second*10, time.Millisecond*100)
}
