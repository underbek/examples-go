package rabbitmq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
)

func (s *TestSuite) Test_Main() {
	l, err := logger.New(true)
	s.Require().NoError(err)

	queueName := "main_test_queue"
	exchangeName := "main_test_exchange"
	routingKey := "main_test_key"

	ch, _ := s.getChannel()
	_, err = ch.QueueInspect(queueName)
	s.Require().ErrorContains(err, "NOT_FOUND - no queue") // error causes channel closing

	ch, cancel := s.getChannel()
	defer cancel()

	s.Require().NoError(
		NewManager(l, ch).DeclareQueueAndExchange(
			QueueDeclare{Queue: queueName}, ExchangeDeclare{Exchange: exchangeName}, QueueBind{RoutingKey: routingKey},
		),
	)

	expected := amqp.Publishing{
		Headers: map[string]interface{}{
			"test-header": "test-header",
		},
		ContentType:   "text",
		CorrelationId: "correlation_id",
		MessageId:     "message_id",
		AppId:         "app_id",
		Body:          []byte("test message"),
	}

	s.Require().NoError(
		NewProducer(l, ch, exchangeName).Publish(context.Background(), PublishMessage{
			RoutingKey: routingKey,
			Message:    expected,
		}),
	)

	s.checkQueue(ch, queueName, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		consumerCh, consumerCancel := s.getChannel()

		err = NewConsumer(l, consumerCh, queueName, true).
			Consume(
				context.Background(),
				Consume{
					ConsumerTag: "test-consumer",
				},
				func(ctx context.Context, msg amqp.Delivery) {
					actual := amqp.Publishing{
						Headers:       msg.Headers,
						ContentType:   msg.ContentType,
						CorrelationId: msg.CorrelationId,
						MessageId:     msg.MessageId,
						AppId:         msg.AppId,
						Body:          msg.Body,
					}

					s.Assert().NoError(msg.Ack(false))
					s.Assert().Equal(expected, actual)
					s.Assert().Less(time.Since(msg.Timestamp), time.Second*3)

					consumerCancel()
				})

		s.Assert().ErrorContains(err, "delivery channel was closed")
		wg.Done()
	}()

	wg.Wait()

	s.checkQueue(ch, queueName, 0)
}
