package usecase

import (
	"encoding/json"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/http"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/repository"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

type TaskService interface {
	CreateTask(task *domain.Task) (string, error)
	GetTaskResult(id string) (*domain.TaskResult, error)
	ProcessTask(task *domain.Task) error
}

func NewTaskService(repos repository.TaskRepository, mq mq.WorkQueue) (TaskService, error) {
	return &taskService{
		repos:     repos,
		workQueue: mq,
	}, nil
}

type taskService struct {
	repos     repository.TaskRepository
	workQueue mq.WorkQueue
}

func (t *taskService) CreateTask(task *domain.Task) (string, error) {
	task.ID = uuid.NewV4().String()

	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskNew,
	}

	if err := t.repos.CreateTaskResult(&taskResult); err != nil {
		return "", err
	}

	taskJson, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	t.workQueue.Publish(string(taskJson))

	return task.ID, nil
}

func (t *taskService) GetTaskResult(id string) (*domain.TaskResult, error) {
	// get task result from MongoDB
	taskResult, err := t.repos.GetTaskResult(id)
	if err != nil {
		return nil, err
	}

	return taskResult, nil
}

func (t *taskService) ProcessTask(task *domain.Task) error {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return err
	}

	log.Debug().Msg(string(taskJson))

	taskResult := domain.TaskResult{
		TaskID: task.ID,
		Status: domain.TaskInProgress,
	}

	// update task result state to "in progress" in MongoDB
	t.repos.UpdateTaskResult(&taskResult)

	// perform task request
	httpClient := http.NewClient()
	req := http.Request{
		Method:  task.Method,
		URL:     task.URL,
		Headers: task.Headers,
	}
	resp, err := httpClient.DoRequest(req)
	if err != nil {
		taskResult.Status = domain.TaskError
		// update task result state to "error" in MongoDB
		t.repos.UpdateTaskResult(&taskResult)
		return err
	}

	taskResult.Status = domain.TaskDone
	taskResult.HTTPStatusCode = &resp.StatusCode
	taskResult.Headers = resp.Headers
	taskResult.ContentLength = resp.ContentLength
	// TODO: update task result in MongoDB
	t.repos.UpdateTaskResult(&taskResult)

	return nil
}
