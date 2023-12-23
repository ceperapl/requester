package repository

import (
	"context"
	"errors"

	"github.com/ceperapl/requester/pkg/domain"
)

var (
	// ErrTaskNotFound is an error that indicates that a task result is not found in the repository.
	ErrTaskNotFound = errors.New("task not found")
)

// TaskRepository is the interface that defines the methods for task result storage and retrieval.
type TaskRepository interface {
	CreateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error
	UpdateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error
	GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error)
}
