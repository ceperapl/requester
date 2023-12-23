package rabbitmq

import (
	"context"
	"errors"

	"github.com/ceperapl/requester/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	// ErrCreateWorkQueue is the error for failing to create a work queue.
	ErrCreateWorkQueue = errors.New("couldn't create work queue")
	// ErrPublishMsg is the error for failing to publish a message to the work queue.
	ErrPublishMsg = errors.New("couldn't publish message")
	// ErrConsumeMsgs is the error for failing to consume messages from the work queue.
	ErrConsumeMsgs = errors.New("couldn't consume messages")
	// ErrCloseWorkQueue is the error for failing to close the work queue.
	ErrCloseWorkQueue = errors.New("couldn't close work queue")
)

// New creates and returns a new WorkQueue instance with the given AMQP connection, channel and queue name.
// It declares a durable queue with the given name and returns an error if it fails to do so.
func New(conn *amqp.Connection, channel *amqp.Channel, queueName string) (*WorkQueue, error) {
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

	return &WorkQueue{
		conn:    conn,
		channel: channel,
		queue:   &queue,
	}, nil
}

// WorkQueue is a struct that implements the mq.WorkQueuer interface using RabbitMQ as the message broker.
type WorkQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

// Publish publishes a message to the work queue using the given context.
// It sets the delivery mode to persistent and the content type to application/json.
// It returns an error if it fails to publish the message.
func (w *WorkQueue) Publish(ctx context.Context, message string) error {
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

// Consume consumes messages from the work queue using the given context and processing function.
// It sets the auto-ack to false and manually acknowledges the messages after processing them.
// It returns an error if it fails to consume or acknowledge the messages.
func (w *WorkQueue) Consume(ctx context.Context, doFunc mq.ProcessingFunc) error {
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
