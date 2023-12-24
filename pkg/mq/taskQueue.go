package mq

import (
	"context"

	"github.com/ceperapl/requester/pkg/domain"
)

// ProcessingFunc is a type of function that takes a task as an argument and performs some processing on it.
type ProcessingFunc func(task domain.Task)

// TaskQueuer is an interface that defines two methods for publishing and consuming messages from a queue.
type TaskQueuer interface {
	// Publish is a method that sends a message to the queue.
	// It takes a context and a task as arguments and returns an error if something goes wrong.
	Publish(ctx context.Context, task domain.Task) error

	// Consume is a method that reads messages from the queue and passes them to the processing function.
	// It takes a context and a processing function as arguments and returns an error if something goes wrong.
	Consume(ctx context.Context, doFunc ProcessingFunc) error
}
