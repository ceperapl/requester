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
	// connectTimeout is the duration to wait for a connection to MongoDB.
	connectTimeout = 1 * time.Second
	// retryInterval is the duration to wait between retries of connecting to MongoDB.
	retryInterval = time.Second
	// attemptsCount is the maximum number of attempts to connect to MongoDB.
	attemptsCount = 30
)

var (
	// ErrCreateClient is an error that indicates a failure to create a MongoDB client.
	ErrCreateClient = errors.New("couldn't create MongoDB client")

	// ErrCloseConnection is an error that indicates a failure to close a MongoDB connection.
	ErrCloseConnection = errors.New("couldn't close MongoDB connection")
)

// NewClient creates and returns a new MongoDB client with the given URI.
// It retries to connect to MongoDB with the specified interval and number of attempts.
// It returns an error if it fails to create the client or connect to MongoDB.
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

// Close closes the MongoDB connection for the given client.
// It returns an error if it fails to disconnect the client.
func Close(ctx context.Context, client *mongo.Client) error {
	if err := client.Disconnect(ctx); err != nil {
		return errors.Join(ErrCloseConnection, err)
	}

	return nil
}
