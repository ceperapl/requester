package rabbitmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/ceperapl/requester/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	retryInterval = time.Second
	attemptsCount = 30
)

var (
	ErrCreateConnection = errors.New("coudn't connect to RabbitMQ")
)

func NewConnection(uri string) (*amqp.Connection, *amqp.Channel, error) {
	var conn *amqp.Connection
	var channel *amqp.Channel

	if err := utils.Retry(retryInterval, attemptsCount, func(attempt int) (bool, error) {
		log.Info().Msg(fmt.Sprintf("connecting to RabbitMQ (attempt #%d/%d)", attempt, attemptsCount))

		var err error
		conn, err = amqp.Dial(uri)
		if err != nil {
			return false, errors.Join(ErrCreateConnection, err)
		}

		channel, err = conn.Channel()
		if err != nil {
			return false, errors.Join(ErrCreateConnection, err)
		}

		return true, nil
	}); err != nil {
		return nil, nil, errors.Join(ErrCreateConnection, err)
	}

	return conn, channel, nil
}

func CloseConnection(conn *amqp.Connection, channel *amqp.Channel) error {
	if err := channel.Close(); err != nil {
		return errors.Join(ErrCloseWorkQueue, err)
	}
	if err := conn.Close(); err != nil {
		return errors.Join(ErrCloseWorkQueue, err)
	}

	return nil
}
