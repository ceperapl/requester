package mq

import (
	"context"
)

type ProcessingFunc func(message string)

type WorkQueue interface {
	Publish(ctx context.Context, message string) error
	Consume(ctx context.Context, doFunc ProcessingFunc) error
}
