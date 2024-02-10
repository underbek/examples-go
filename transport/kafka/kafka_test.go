package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/testcontainers"
)

const (
	groupID   = "test-group"
	groupID2  = "test-group2"
	groupID3  = "test-group3"
	topic     = "test-topic"
	topic2    = "test-topic2"
	topic3    = "test-topic2"
	topic4    = "test-topic4"
	key       = "test-key"
	partition = 0
)

type TestSuiteKafkaTransport struct {
	suite.Suite
	container *testcontainer.KafkaContainer
	conn      *kafka.Conn
}

func (s *TestSuiteKafkaTransport) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
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

	producer, err := NewProducer(lg, ProducerConfig{
		Brokers:                s.container.GetBrokers(),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(producer.Close()) }()

	producer2, err := NewProducer(lg, ProducerConfig{
		Brokers:                s.container.GetBrokers(),
		Topic:                  topic2,
		AllowAutoTopicCreation: true,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(producer2.Close()) }()

	producer3, err := NewProducer(lg, ProducerConfig{
		Brokers:                s.container.GetBrokers(),
		AllowAutoTopicCreation: true,
		AllowManualTopic:       true,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(producer3.Close()) }()

	producer4, err := NewProducer(lg, ProducerConfig{
		Brokers:                s.container.GetBrokers(),
		Topic:                  topic4,
		AllowAutoTopicCreation: true,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(producer4.Close()) }()

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

	err = producer2.Publish(context.Background(), msg)
	s.Require().NoError(err)

	err = producer4.Publish(context.Background(), msg)
	s.Require().NoError(err)

	msgWithTopic := msg
	msgWithTopic.Topic = topic3

	err = producer3.Publish(context.Background(), msgWithTopic)
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

	consumer2, err := NewConsumer(lg, ConsumerConfig{
		Brokers:             s.container.GetBrokers(),
		GroupTopics:         []string{topic, topic2},
		GroupID:             groupID2,
		ManualRetryDuration: time.Second,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(consumer2.Close()) }()

	consumer3, err := NewConsumer(lg, ConsumerConfig{
		Brokers:             s.container.GetBrokers(),
		GroupTopics:         []string{topic3},
		GroupID:             groupID3,
		ManualRetryDuration: time.Second,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(consumer3.Close()) }()

	consumer4, err := NewConsumer(lg, ConsumerConfig{
		Brokers:             s.container.GetBrokers(),
		GroupTopics:         []string{topic4},
		GroupID:             groupID3,
		ManualRetryDuration: time.Second,
	})
	s.Require().NoError(err)
	defer func() { s.Assert().NoError(consumer4.Close()) }()

	expectedMessagesCount := 2
	var consumedMessages int
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	_ = consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		s.Assert().Equal(topic, msg.Topic)
		s.Assert().Equal(key, string(msg.Key))
		s.Assert().Equal(expected.Headers, msg.Headers)
		s.Assert().Equal(expected.Value, msg.Value)

		consumedMessages++
		if consumedMessages == expectedMessagesCount {
			cancel()
		}

		return nil
	})

	s.Require().Equal(expectedMessagesCount, consumedMessages)

	expectedMessagesCount = 3
	consumedMessages = 0
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	_ = consumer2.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		s.Assert().Contains([]string{topic, topic2}, msg.Topic)
		s.Assert().Equal(key, string(msg.Key))
		s.Assert().Equal(expected.Headers, msg.Headers)
		s.Assert().Equal(expected.Value, msg.Value)

		consumedMessages++
		if consumedMessages == expectedMessagesCount {
			cancel()
		}

		return nil
	})

	s.Require().Equal(expectedMessagesCount, consumedMessages)

	expectedMessagesCount = 2
	consumedMessages = 0
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	_ = consumer3.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		s.Assert().Contains([]string{topic3}, msg.Topic)
		s.Assert().Equal(key, string(msg.Key))
		s.Assert().Equal(expected.Headers, msg.Headers)
		s.Assert().Equal(expected.Value, msg.Value)

		consumedMessages++
		if consumedMessages == expectedMessagesCount {
			cancel()
		}

		return nil
	})

	s.Require().Equal(expectedMessagesCount, consumedMessages)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_ = consumer4.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		panic("some panic")
	})
}
