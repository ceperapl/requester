package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/http"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/repository"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrCreateTask    = errors.New("create task failed")
	ErrGetTaskResult = errors.New("get task result failed")
	ErrProcessTask   = errors.New("process task failed")
)

type TaskUsecaser interface {
	CreateTask(ctx context.Context, task *domain.Task) (string, error)
	GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error)
	ProcessTask(ctx context.Context, task *domain.Task) error
}

func NewTaskService(repos repository.TaskRepository, mq mq.WorkQueuer) (*TaskService, error) {
	return &TaskService{
		repos:     repos,
		workQueue: mq,
	}, nil
}

type TaskService struct {
	repos     repository.TaskRepository
	workQueue mq.WorkQueuer
}

func (t *TaskService) CreateTask(ctx context.Context, task *domain.Task) (string, error) {
	task.ID = uuid.NewV4().String()

	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskNew,
	}

	if err := t.repos.CreateTaskResult(ctx, &taskResult); err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}
	if err := t.workQueue.Publish(ctx, string(taskJSON)); err != nil {
		return "", errors.Join(ErrCreateTask, err)
	}

	return task.ID, nil
}

func (t *TaskService) GetTaskResult(ctx context.Context, id string) (*domain.TaskResult, error) {
	// get task result from MongoDB
	taskResult, err := t.repos.GetTaskResult(ctx, id)
	if err != nil {
		//nolint: wrapcheck
		return nil, err
	}

	return taskResult, nil
}

func (t *TaskService) ProcessTask(ctx context.Context, task *domain.Task) error {
	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskInProgress,
	}

	// update task result state to "in progress" in MongoDB
	if err := t.repos.UpdateTaskResult(ctx, &taskResult); err != nil {
		return errors.Join(ErrProcessTask, err)
	}

	// perform task request
	httpClient := http.NewClient()
	req := http.Request{
		Method:  task.Method,
		URL:     task.URL,
		Headers: task.Headers,
	}
	resp, err := httpClient.DoRequest(ctx, req)
	if err != nil {
		taskResult.Status = domain.TaskError
		taskResult.Error = err.Error()
		// update task result state to "error" in MongoDB
		if updTaskErr := t.repos.UpdateTaskResult(ctx, &taskResult); updTaskErr != nil {
			return errors.Join(ErrProcessTask, updTaskErr)
		}

		return errors.Join(ErrProcessTask, err)
	}

	taskResult.Status = domain.TaskDone
	taskResult.HTTPStatusCode = &resp.StatusCode
	taskResult.Headers = resp.Headers
	taskResult.ContentLength = resp.ContentLength
	// update task result in MongoDB
	if err := t.repos.UpdateTaskResult(ctx, &taskResult); err != nil {
		return errors.Join(ErrProcessTask, err)
	}

	return nil
}
