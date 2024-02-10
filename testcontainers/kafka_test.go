package testcontainer

import (
	"context"
	"testing"
	"time"

	gokafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
)

const (
	topic     = "dummy-topic"
	key       = "dummy-key"
	value     = "dummy-value"
	partition = 0
)

type TestSuiteKafka struct {
	suite.Suite
	container *KafkaContainer
}

func (s *TestSuiteKafka) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	var err error
	s.container, err = NewKafkaContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestSuiteKafka) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.container.Terminate(ctx)
	s.NoError(err)
}

func TestSuiteKafka_Run(t *testing.T) {
	suite.Run(t, new(TestSuiteKafka))
}

func (s *TestSuiteKafka) Test_ConnectKafka() {
	_, err := gokafka.DialLeader(context.Background(), "tcp", s.container.GetBrokers()[0], topic, partition)
	s.NoError(err)

	//publishes a message to the kafka topic
	publish(s)
	//reads a message from the kafka topic
	consume(s)

}

func publish(s *TestSuiteKafka) {
	publisher := gokafka.Writer{
		Addr:  gokafka.TCP(s.container.GetBrokers()[0]),
		Topic: topic,
	}
	defer publisher.Close()

	err := publisher.WriteMessages(
		context.Background(),
		gokafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
	s.NoError(err)
}

func consume(s *TestSuiteKafka) {
	consumer := gokafka.NewReader(gokafka.ReaderConfig{
		Brokers:     s.container.GetBrokers(),
		Topic:       topic,
		Partition:   partition,
		StartOffset: gokafka.LastOffset,
	})
	defer consumer.Close()

	m, err := consumer.ReadMessage(context.Background())
	s.NoError(err)

	s.Equal(key, string(m.Key))
	s.Equal(value, string(m.Value))
}
