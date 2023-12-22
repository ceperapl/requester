package mongodb

import (
	"context"
	"errors"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCreateTaskResult = errors.New("couldn't create task result")
	ErrUpdateTaskResult = errors.New("couldn't update task result")
	ErrGetTaskResult    = errors.New("couldn't get task result")
)

type taskRepo struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewTaskRepo(client *mongo.Client, database string, collection string) repository.TaskRepository {
	return &taskRepo{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (t *taskRepo) CreateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	_, err := collection.InsertOne(ctx, taskResult)
	if err != nil {
		return errors.Join(ErrCreateTaskResult, err)
	}

	return nil
}

func (t *taskRepo) UpdateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	filter := bson.M{"taskid": taskResult.TaskID}

	if _, err := collection.ReplaceOne(ctx, filter, taskResult); err != nil {
		return errors.Join(ErrUpdateTaskResult, err)
	}

	return nil
}

func (t *taskRepo) GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error) {
	collection := t.client.Database(t.database).Collection(t.collection)

	filter := bson.M{"taskid": id}

	// create a value into which the result can be decoded
	var result domain.TaskResult

	if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.Join(repository.ErrTaskNotFound, err)
		}
		return nil, errors.Join(ErrGetTaskResult, err)
	}

	return &result, nil
}
