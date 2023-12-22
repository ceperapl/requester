package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ceperapl/requester/pkg/utils"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectTimeout = 1 * time.Second
	retryInterval  = time.Second
	attemptsCount  = 30
)

var (
	ErrCreateClient    = errors.New("couldn't create MongoDB client")
	ErrCloseConnection = errors.New("couldn't close MongoDB connection")
)

func NewClient(uri string) (*mongo.Client, error) {
	var client *mongo.Client

	if err := utils.Retry(retryInterval, attemptsCount, func(attempt int) (bool, error) {
		log.Info().Msg(fmt.Sprintf("connecting to MongoDB (attempt #%d/%d)", attempt, attemptsCount))

		// Set client options
		clientOptions := options.Client().ApplyURI(uri)

		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		// Connect to MongoDB
		var err error
		if client, err = mongo.Connect(ctx, clientOptions); err != nil {
			return false, errors.Join(ErrCreateClient, err)
		}

		// Check the connection
		if err := client.Ping(ctx, nil); err != nil {
			return false, errors.Join(ErrCreateClient, err)
		}

		return true, nil
	}); err != nil {
		return nil, errors.Join(ErrCreateClient, err)
	}

	return client, nil
}

func Close(ctx context.Context, client *mongo.Client) error {
	if err := client.Disconnect(ctx); err != nil {
		return errors.Join(ErrCloseConnection, err)
	}

	return nil
}
