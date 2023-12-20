package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	retryInterval    = time.Second
	retryMaxAttempts = 30
)

func New(uri string, queueName string) (mq.WorkQueue, error) {
	// Connect to RabbitMQ with retry
	var conn *amqp.Connection
	var channel *amqp.Channel
	var queue amqp.Queue

	if err := utils.Retry("connect to the RabbitMQ", retryInterval, retryMaxAttempts, func() (bool, error) {
		log.Info().Msg(fmt.Sprintf("connect to RabbitMQ; uri: %q; queue: %q", uri, queueName))

		var err error
		conn, err := amqp.Dial(uri)
		if err != nil {
			return false, fmt.Errorf("connect to RabbitMQ: %w", err)
		}

		channel, err = conn.Channel()
		if err != nil {
			return false, fmt.Errorf("open a channel: %w", err)
		}

		if queue, err = channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		); err != nil {
			return false, fmt.Errorf("declare a queue: %w", err)
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("retry connecting to the RabbitMQ: %w", err)
	}
	return &workQueue{
		conn:    conn,
		channel: channel,
		queue:   &queue,
	}, nil
}

type workQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func (w *workQueue) Publish(message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := w.channel.PublishWithContext(ctx,
		"",           // exchange
		w.queue.Name, // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("publish a message: %w", err)
	}

	return nil
}

func (w *workQueue) Consume(doFunc mq.ProcessingFunc) error {
	msgs, err := w.channel.Consume(
		w.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("register a consumer: %w", err)
	}

	for d := range msgs {
		if err := doFunc(string(d.Body)); err != nil {
			return fmt.Errorf("processing message: %w", err)
		}
		d.Ack(false)
	}

	return nil
}

func (w *workQueue) Close() error {
	if err := w.channel.Close(); err != nil {
		return fmt.Errorf("close RabbitMQ channel: %w", err)
	}
	if err := w.conn.Close(); err != nil {
		return fmt.Errorf("close RabbitMQ connection: %w", err)
	}

	return nil
}
