package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/taskexec"
	uuid "github.com/satori/go.uuid"
)

// ErrCreateTask is an error that indicates a failure to create a task.
var ErrCreateTask = errors.New("create task failed")

// ErrGetTaskResult is an error that indicates a failure to get a task result.
var ErrGetTaskResult = errors.New("get task result failed")

// ErrProcessTask is an error that indicates a failure to process a task.
var ErrProcessTask = errors.New("process task failed")

// TaskUsecaser is the interface that defines the methods for task use cases.
type TaskUsecaser interface {
	CreateTask(ctx context.Context, task *domain.Task) (string, error)
	GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error)
	ProcessTask(ctx context.Context, task *domain.Task) error
}

// NewTaskService creates and returns a new TaskService instance.
func NewTaskService(repos repository.TaskRepository, mq mq.WorkQueuer) *TaskService {
	taskExec := taskexec.NewTaskExecution()

	return &TaskService{
		repos:        repos,
		workQueue:    mq,
		taskExecutor: taskExec,
	}
}

// TaskService is a struct that implements the TaskUsecaser interface.
// It uses a task repository and a work queue to manage tasks and their results.
type TaskService struct {
	repos        repository.TaskRepository
	workQueue    mq.WorkQueuer
	taskExecutor taskexec.TaskExecutor
}

// CreateTask creates a new task and returns its ID.
// It generates a unique ID for the task, creates a new task result with status "new", saves it to the task repository,
// and publishes the task to the work queue.
// It returns an error if it fails to marshal the task, save the task result, or publish the task.
func (ts *TaskService) CreateTask(ctx context.Context, task *domain.Task) (string, error) {
	task.ID = uuid.NewV4().String()

	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskNew,
	}

	if err := ts.repos.CreateTaskResult(ctx, &taskResult); err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}
	if err := ts.workQueue.Publish(ctx, string(taskJSON)); err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}

	return task.ID, nil
}

// GetTaskResult returns the task result for the given ID.
// It retrieves the task result from the task repository and returns it.
// It returns an error if it fails to get the task result from the repository.
func (ts *TaskService) GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error) {
	// get task result from MongoDB
	taskResult, err := ts.repos.GetTaskResult(ctx, id)
	if err != nil {
		//nolint: wrapcheck
		return nil, err
	}

	return taskResult, nil
}

// ProcessTask executes the task and updates its result.
func (ts *TaskService) ProcessTask(ctx context.Context, task *domain.Task) error {
	// update task result state to "in progress" in MongoDB
	if err := ts.repos.UpdateTaskResult(ctx, &domain.TaskResult{TaskID: task.ID, Status: domain.TaskInProgress}); err != nil {
		return errors.Join(ErrProcessTask, err)
	}

	// execute the task
	taskResult, err := ts.taskExecutor.ExecuteTask(ctx, task)
	if err != nil {
		// update task result state to "error" in MongoDB
		updTaskErr := ts.repos.UpdateTaskResult(ctx,
			&domain.TaskResult{
				TaskID: task.ID,
				Status: domain.TaskError,
				Error:  err.Error(),
			})
		if updTaskErr != nil {
			return errors.Join(ErrProcessTask, updTaskErr)
		}

		return errors.Join(ErrProcessTask, err)
	}

	// update task result in MongoDB
	if err := ts.repos.UpdateTaskResult(ctx, taskResult); err != nil {
		return errors.Join(ErrProcessTask, err)
	}

	return nil
}
