package repository

import (
	"context"
	"errors"

	"github.com/ceperapl/requester/pkg/domain"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository interface {
	CreateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error
	UpdateTaskResult(ctx context.Context, taskResult *domain.TaskResult) error
	GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error)
}
