package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/taskexec"
	uuid "github.com/satori/go.uuid"
)

// ErrProcessTask is an error that indicates a failure to process a task.
var ErrProcessTask = errors.New("process task failed")

// TaskUsecaser is the interface that defines the methods for task use cases.
type TaskUsecaser interface {
	CreateTask(ctx context.Context, task domain.Task) (string, error)
	GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error)
	ProcessTask(ctx context.Context, task domain.Task) (*domain.TaskResult, error)
}

// NewTaskService creates and returns a new TaskService instance.
func NewTaskService(repos repository.TaskRepository, mq mq.TaskQueuer, taskExec taskexec.TaskExecutor) *TaskService {
	return &TaskService{
		repos:        repos,
		taskQueue:    mq,
		taskExecutor: taskExec,
	}
}

// TaskService is a struct that implements the TaskUsecaser interface.
// It uses a task repository and a work queue to manage tasks and their results.
type TaskService struct {
	repos        repository.TaskRepository
	taskQueue    mq.TaskQueuer
	taskExecutor taskexec.TaskExecutor
}

// CreateTask creates a new task and returns its ID.
// It generates a unique ID for the task, creates a new task result with status "new", saves it to the task repository,
// and publishes the task to the work queue.
// It returns an error if it fails to marshal the task, save the task result, or publish the task.
func (ts *TaskService) CreateTask(ctx context.Context, task domain.Task) (string, error) {
	task.ID = uuid.NewV4().String()

	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskNew,
	}
	if err := ts.repos.CreateTaskResult(ctx, &taskResult); err != nil {
		return "", fmt.Errorf("couldn't create task result in repository: %w", err)
	}

	if err := ts.taskQueue.Publish(ctx, task); err != nil {
		return "", fmt.Errorf("couldn't publish task to queue: %w", err)
	}

	return task.ID, nil
}

// GetTaskResult returns the task result for the given ID.
// It retrieves the task result from the task repository and returns it.
// It returns an error if it fails to get the task result from the repository.
func (ts *TaskService) GetTaskResult(ctx context.Context, taskID string) (*domain.TaskResult, error) {
	// get task result from MongoDB
	taskResult, err := ts.repos.GetTaskResult(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get task result from repository: %w", err)
	}

	return taskResult, nil
}

// ProcessTask executes the task and updates its result.
//nolint: nonamedreturns
func (ts *TaskService) ProcessTask(ctx context.Context, task domain.Task) (taskResult *domain.TaskResult, err error) {
	taskResult = &domain.TaskResult{TaskID: task.ID, Status: domain.TaskInProgress}
	// update task result in MongoDB
	if updateErr := ts.repos.UpdateTaskResult(ctx, taskResult); updateErr != nil {
		return nil, fmt.Errorf("couldn't update task result in repository: %w", updateErr)
	}

	var errExec error
	result, errExec := ts.taskExecutor.ExecuteTask(ctx, task)
	if errExec != nil {
		taskResult = &domain.TaskResult{
			TaskID: task.ID,
			Status: domain.TaskError,
			Error:  errExec.Error(),
		}
		err = fmt.Errorf("couldn't execute task: %w", errExec)
	} else {
		taskResult = result
	}

	// update task result in MongoDB
	if updateErr := ts.repos.UpdateTaskResult(ctx, taskResult); updateErr != nil {
		return nil, fmt.Errorf("couldn't update task result in repository: %w", updateErr)
	}

	return taskResult, err
}
