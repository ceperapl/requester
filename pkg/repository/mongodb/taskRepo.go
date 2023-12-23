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
	// ErrCreateTaskResult is an error that indicates a failure to create a task result in MongoDB.
	ErrCreateTaskResult = errors.New("couldn't create task result")
	// ErrUpdateTaskResult is an error that indicates a failure to update a task result in MongoDB.
	ErrUpdateTaskResult = errors.New("couldn't update task result")
	// ErrGetTaskResult is an error that indicates a failure to get a task result from MongoDB.
	ErrGetTaskResult = errors.New("couldn't get task result")
)

// TaskRepo is a struct that implements the repository.TaskRepository interface using MongoDB as the storage.
type TaskRepo struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewTaskRepo creates and returns a new TaskRepo instance with the given MongoDB client, database and collection.
func NewTaskRepo(client *mongo.Client, database string, collection string) *TaskRepo {
	return &TaskRepo{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// CreateTaskResult creates a new task result document in MongoDB using the given task result.
// It returns an error if it fails to insert the document.
func (t *TaskRepo) CreateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	_, err := collection.InsertOne(ctx, taskResult)
	if err != nil {
		return errors.Join(ErrCreateTaskResult, err)
	}

	return nil
}

// UpdateTaskResult updates an existing task result document in MongoDB using the given task result.
// It replaces the document that matches the task ID of the task result.
// It returns an error if it fails to replace the document.
func (t *TaskRepo) UpdateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error {
	collection := t.client.Database(t.database).Collection(t.collection)

	filter := bson.M{"taskid": taskResult.TaskID}

	if _, err := collection.ReplaceOne(ctx, filter, taskResult); err != nil {
		return errors.Join(ErrUpdateTaskResult, err)
	}

	return nil
}

// GetTaskResult gets a task result document from MongoDB using the given task ID.
// It returns the task result and an error if any.
// If the task ID is not found, it returns repository.ErrTaskNotFound.
func (t *TaskRepo) GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error) {
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
