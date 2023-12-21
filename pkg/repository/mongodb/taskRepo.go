package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/utils"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	retryInterval    = time.Second
	retryMaxAttempts = 30
)

type taskRepo struct {
	client     *mongo.Client
	database   string
	collection string
	ctx        context.Context
}

func NewTaskRepo(uri string, database string, collection string) (repository.TaskRepository, error) {
	// Connect to RabbitMQ with retry
	var client *mongo.Client
	if err := utils.Retry("connect to the MongoDB", retryInterval, retryMaxAttempts, func() (bool, error) {
		log.Info().Msg(fmt.Sprintf("connect to MongoDB; uri: %q", uri))

		// Set client options
		clientOptions := options.Client().ApplyURI(uri)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Connect to MongoDB
		var err error
		if client, err = mongo.Connect(ctx, clientOptions); err != nil {
			return false, fmt.Errorf("create MongoDB client: %w", err)
		}

		// Check the connection
		if err := client.Ping(ctx, nil); err != nil {
			return false, fmt.Errorf("connect to MongoDB: %w", err)
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("retry connecting to the RabbitMQ: %w", err)
	}

	return &taskRepo{
		client:     client,
		database:   database,
		collection: collection,
		ctx:        context.Background(),
	}, nil
}

func (t *taskRepo) CreateTaskResult(taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	_, err := collection.InsertOne(t.ctx, taskResult)
	if err != nil {
		return err
	}

	return nil
}

func (t *taskRepo) UpdateTaskResult(taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	filter := bson.M{"taskid": taskResult.TaskID}

	if _, err := collection.ReplaceOne(t.ctx, filter, taskResult); err != nil {
		return err
	}

	return nil
}

func (t *taskRepo) GetTaskResult(id string) (*domain.TaskResult, error) {
	collection := t.client.Database(t.database).Collection(t.collection)

	filter := bson.M{"taskid": id}

	// create a value into which the result can be decoded
	var result domain.TaskResult

	if err := collection.FindOne(t.ctx, filter).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *taskRepo) Close(ctx context.Context) error {
	if err := t.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}
