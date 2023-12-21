package repository

import (
	"context"

	"github.com/ceperapl/requester/pkg/domain"
)

type TaskRepository interface {
	CreateTaskResult(taskResult *domain.TaskResult) error
	UpdateTaskResult(taskResult *domain.TaskResult) error
	GetTaskResult(id string) (*domain.TaskResult, error)
	Close(ctx context.Context) error
}
