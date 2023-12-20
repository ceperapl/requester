package mq

import "io"

type ProcessingFunc func(message string) error

type WorkQueue interface {
	io.Closer
	Publish(message string) error
	Consume(doFunc ProcessingFunc) error
}
