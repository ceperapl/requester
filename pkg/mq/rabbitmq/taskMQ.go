package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ceperapl/requester/pkg/domain"
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
func New(conn *amqp.Connection, channel *amqp.Channel, queueName string) (*TaskQueue, error) {
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

	return &TaskQueue{
		conn:    conn,
		channel: channel,
		queue:   &queue,
	}, nil
}

// TaskQueue is a struct that implements the mq.TaskQueuer interface using RabbitMQ as the message broker.
type TaskQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

// Publish publishes a task to the queue using the given context.
// It sets the delivery mode to persistent and the content type to application/json.
// It returns an error if it fails to publish the message.
func (t *TaskQueue) Publish(ctx context.Context, task domain.Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("couldn't marshal task: %w", err)
	}
	err = t.channel.PublishWithContext(ctx,
		"",           // exchange
		t.queue.Name, // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         taskJSON,
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
func (t *TaskQueue) Consume(ctx context.Context, procFunc mq.ProcessingFunc) error {
	msgs, err := t.channel.ConsumeWithContext(ctx,
		t.queue.Name, // queue
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
		var task domain.Task
		if err := json.Unmarshal(d.Body, &task); err != nil {
			return fmt.Errorf("couldn't unmarshal message from queue: %w", err)
		}
		procFunc(task)
		if err := d.Ack(false); err != nil {
			return errors.Join(ErrConsumeMsgs, err)
		}
	}

	return nil
}
