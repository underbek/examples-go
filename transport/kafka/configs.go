package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	// AppName is a name of the app that will be written to metrics
	AppName string `env:"APP_NAME"`

	// Brokers is a list of kafka nodes(foo.wb.ru:9092)
	Brokers []string `env:"KAFKA_BROKERS"`

	// DialTimeout is the maximum amount of time a dial will wait for
	DialTimeout time.Duration `env:"KAFKA_DIAL_TIMEOUT"`

	// CommitInterval indicates the interval at which offsets are committed to
	// the broker.  If 0, commits will be handled synchronously.
	CommitInterval time.Duration `env:"KAFKA_COMMIT_INTERVAL"`

	// GroupID is a kafka consumer group id
	GroupID string `env:"KAFKA_GROUP_ID"`

	// User is a login for sasl plain auth.
	User string `env:"KAFKA_USER"`

	// Password is a password for sasl plain auth
	Password string `env:"KAFKA_PASSWORD"`

	// Topic is a kafka topic, all messages will be read from
	Topic string `env:"KAFKA_TOPIC"`

	// Topics is an array of topics for setup multiple readers
	Topics []string `env:"KAFKA_TOPICS"`

	// GroupTopics is an array of topic for one(!) reader
	GroupTopics []string `env:"KAFKA_GROUP_TOPICS"`

	// ManualRetries is a max count of reties for fetch and commit operations outside of driver logic
	ManualRetries int `env:"KAFKA_MANUAL_RETRIES"`

	// ManualRetryDuration duration between manual retries
	ManualRetryDuration time.Duration `env:"KAFKA_MANUAL_RETRY_DURATION"`

	BatchSize int `env:"KAFKA_BATCH_SIZE"`

	BatchTimeout time.Duration `env:"KAFKA_BATCH_TIMEOUT" default:"1s"`

	FlushTime time.Duration
}

type ProducerConfig struct {
	Brokers                []string
	Topic                  string
	Balancer               kafka.Balancer
	MaxAttempts            int
	WriteBackoffMin        time.Duration
	WriteBackoffMax        time.Duration
	BatchSize              int
	BatchBytes             int64
	BatchTimeout           time.Duration
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	RequiredAcks           kafka.RequiredAcks
	Async                  bool
	Compression            kafka.Compression
	AllowAutoTopicCreation bool
}
