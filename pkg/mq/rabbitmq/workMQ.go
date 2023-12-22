package rabbitmq

import (
	"context"
	"errors"

	"github.com/ceperapl/requester/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrCreateWorkQueue = errors.New("couldn't create work queue")
	ErrPublishMsg      = errors.New("couldn't publish message")
	ErrConsumeMsgs     = errors.New("couldn't consume messages")
	ErrCloseWorkQueue  = errors.New("couldn't close work queue")
)

func New(conn *amqp.Connection, channel *amqp.Channel, queueName string) (mq.WorkQueue, error) {
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, errors.Join(ErrCreateWorkQueue, err)
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

func (w *workQueue) Publish(ctx context.Context, message string) error {
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
		return errors.Join(ErrPublishMsg, err)
	}

	return nil
}

func (w *workQueue) Consume(ctx context.Context, doFunc mq.ProcessingFunc) error {
	msgs, err := w.channel.ConsumeWithContext(ctx,
		w.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return errors.Join(ErrConsumeMsgs, err)
	}

	for d := range msgs {
		doFunc(string(d.Body))
		if err := d.Ack(false); err != nil {
			return errors.Join(ErrConsumeMsgs, err)
		}
	}

	return nil
}
