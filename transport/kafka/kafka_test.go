package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/logger"
	testcontainer "github.com/underbek/examples-go/testcontainers"
)

const (
	groupID   = "test-group"
	topic     = "test-topic"
	key       = "test-key"
	partition = 0
)

type TestSuiteKafkaTransport struct {
	suite.Suite
	container *testcontainer.KafkaContainer
	conn      *kafka.Conn
}

func (s *TestSuiteKafkaTransport) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	var err error
	s.container, err = testcontainer.NewKafkaContainer(ctx)
	s.Require().NoError(err)

	s.conn, err = kafka.DialLeader(context.Background(), "tcp", s.container.GetBrokers()[0], topic, partition)
	s.Require().NoError(err)
}

func (s *TestSuiteKafkaTransport) TearDownSuite() {
	s.Require().NoError(s.conn.Close())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.container.Terminate(ctx)
	s.Require().NoError(err)
}

func TestSuiteKafkaTransport_Run(t *testing.T) {
	suite.Run(t, new(TestSuiteKafkaTransport))
}

func (s *TestSuiteKafkaTransport) Test_Kafka() {
	lg, err := logger.New(true)
	s.Require().NoError(err)

	producer := NewProducer(lg, ProducerConfig{
		Brokers:                s.container.GetBrokers(),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
	})
	defer func() { s.Assert().NoError(producer.Close()) }()

	msg := kafka.Message{
		Headers: []kafka.Header{
			{
				Key:   "test-header",
				Value: []byte("test-header"),
			},
		},
		Key:   []byte(key),
		Value: []byte("test message"),
	}

	err = producer.Publish(context.Background(), msg)
	s.Require().NoError(err)

	err = producer.Publish(context.Background(), msg)
	s.Require().NoError(err)

	expected := msg

	consumer, err := NewConsumer(lg, ConsumerConfig{
		Brokers:             s.container.GetBrokers(),
		Topic:               topic,
		GroupID:             groupID,
		ManualRetryDuration: time.Second,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(consumer.Close()) }()

	isRead := false

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	_ = consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		s.Assert().Equal(topic, msg.Topic)
		s.Assert().Equal(key, string(msg.Key))
		s.Assert().Equal(expected.Headers, msg.Headers)
		s.Assert().Equal(expected.Value, msg.Value)

		if isRead {
			cancel()
		}

		isRead = true

		return nil
	})

	s.Require().True(isRead)
}
