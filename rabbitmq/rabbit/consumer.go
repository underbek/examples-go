package rabbit

import (
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	errChan      chan *amqp.Error
	host         string
	exchangeName string
	queueName    string
	bindingKey   string
	threads      int
}

func New(host, exchangeName, queueName, bindingKey string, threads int) *Consumer {
	return &Consumer{
		host:         host,
		exchangeName: exchangeName,
		queueName:    queueName,
		bindingKey:   bindingKey,
		threads:      threads,
	}
}

func (c *Consumer) connect() (queue amqp.Queue, err error) {

	// создаем соединение
	c.conn, err = amqp.Dial(c.host)
	if err != nil {
		return
	}

	// канал для уведомления об ошибках
	c.errChan = c.conn.NotifyClose(make(chan *amqp.Error, 1))

	// создаем уникальный канал для получения сообщений
	c.channel, err = c.conn.Channel()
	if err != nil {
		return
	}

	// декларируем exchange
	if err = c.channel.ExchangeDeclare(
		c.exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return
	}

	// декларируем queue
	if queue, err = c.channel.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return
	}

	// соединяем queue с exchange
	err = c.channel.QueueBind(
		queue.Name,
		c.bindingKey,
		c.exchangeName,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "tasks",
			"x-dead-letter-routing-key": "delete_dead",
		},
	)

	return
}

func (c *Consumer) Handle(handler func(amqp.Delivery)) error {

	// подключаемся
	queue, err := c.connect()
	if err != nil {
		return err
	}

	// создаем consumer
	msgChan, err := c.channel.Consume(
		queue.Name,
		"test_consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// обрабатываем в несколько горутин
	for i := 0; i < c.threads; i++ {
		go func() {
			for msg := range msgChan {
				handler(msg)
			}
		}()
	}

	return <-c.errChan
}

//{"user_id":"1", "urls":["http://localhost:15672/#/exchanges"]}
