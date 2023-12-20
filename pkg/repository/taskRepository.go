package repository

import (
	"io"

	"github.com/ceperapl/requester/pkg/domain"
)

type TaskRepository interface {
	io.Closer
	CreateTaskResult(taskResult *domain.TaskResult) error
	UpdateTaskResult(taskResult *domain.TaskResult) error
	GetTaskResult(id string) (*domain.TaskResult, error)
}
