package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ceperapl/requester/pkg/domain"
	mq_mocks "github.com/ceperapl/requester/pkg/mq/mocks"
	task_mocks "github.com/ceperapl/requester/pkg/repository/mocks"
	exec_mocks "github.com/ceperapl/requester/pkg/taskexec/mocks"
	"github.com/ceperapl/requester/pkg/usecase"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	errCreateTaskInRepo      = errors.New("couldn't create task in repository")
	errUpdateTaskInRepo      = errors.New("couldn't update task in repository")
	errGetTaskResultFromRepo = errors.New("couldn't get task result from repository")
	errPublishTaskToMQ       = errors.New("couldn't publish task to queue")
	errExecuteTask           = errors.New("couldn't execute task")
)

func TestCreateTask(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name      string      // The name of the test case
		task      domain.Task // The task to pass to the CreateTask method
		mockRepo  func() *task_mocks.TaskRepository
		mockQueue func() *mq_mocks.TaskQueuer
		isIDEmpty bool
		err       error
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "Successful case",
			task: domain.Task{
				Method: "GET",
				URL:    "https://example.com",
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("CreateTaskResult", mock.Anything, mock.Anything).Return(nil)

				return repo
			},
			mockQueue: func() *mq_mocks.TaskQueuer {
				queue := &mq_mocks.TaskQueuer{}
				queue.On("Publish", mock.Anything, mock.Anything).Return(nil)

				return queue
			},
			isIDEmpty: false,
			err:       nil,
		},
		{
			name: "Error from repository",
			task: domain.Task{
				Method: "GET",
				URL:    "https://example.com",
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("CreateTaskResult", mock.Anything, mock.Anything).Return(errCreateTaskInRepo)

				return repo
			},
			mockQueue: func() *mq_mocks.TaskQueuer {
				queue := &mq_mocks.TaskQueuer{}
				queue.On("Publish", mock.Anything, mock.Anything).Return(nil)

				return queue
			},
			isIDEmpty: true,
			err:       errCreateTaskInRepo,
		},
		{
			name: "Error from message queue",
			task: domain.Task{
				Method: "GET",
				URL:    "https://example.com",
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("CreateTaskResult", mock.Anything, mock.Anything).Return(nil)

				return repo
			},
			mockQueue: func() *mq_mocks.TaskQueuer {
				queue := &mq_mocks.TaskQueuer{}
				queue.On("Publish", mock.Anything, mock.Anything).Return(errPublishTaskToMQ)

				return queue
			},
			isIDEmpty: true,
			err:       errPublishTaskToMQ,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a new task service with the mock repository and queue
			ts := usecase.NewTaskService(tc.mockRepo(), tc.mockQueue(), nil)
			// Call the CreateTask method with the test case task
			id, err := ts.CreateTask(context.TODO(), tc.task)
			// Check if the error matches the expectation
			require.ErrorIs(t, err, tc.err)
			// Check if the id matches the expectation
			if tc.isIDEmpty {
				assert.Empty(t, id)
			} else {
				assert.True(t, IsValidUUID(id))
			}
		})
	}
}

func IsValidUUID(id string) bool {
	_, err := uuid.FromString(id)

	return err == nil
}

func TestGetTaskResult(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name       string // The name of the test case
		mockRepo   func() *task_mocks.TaskRepository
		taskID     string
		taskResult *domain.TaskResult
		err        error
	}

	// Define some test cases
	testCases := []testCase{
		{
			name:   "Successful case",
			taskID: "b9d74525-b06d-4a40-a475-ce10af87c5db",
			taskResult: &domain.TaskResult{
				TaskID: "b9d74525-b06d-4a40-a475-ce10af87c5db",
				Status: domain.TaskInProgress,
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("GetTaskResult", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, id string) *domain.TaskResult {
						return &domain.TaskResult{
							TaskID: id,
							Status: domain.TaskInProgress,
						}
					},
					func(ctx context.Context, id string) error {
						return nil
					},
				)

				return repo
			},
			err: nil,
		},
		{
			name:       "Error from repository",
			taskResult: nil,
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("GetTaskResult", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, id string) *domain.TaskResult {
						return nil
					},
					func(ctx context.Context, id string) error {
						return errGetTaskResultFromRepo
					},
				)

				return repo
			},
			err: errGetTaskResultFromRepo,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a new task service with the mock repository and queue
			ts := usecase.NewTaskService(tc.mockRepo(), nil, nil)
			// Call the GetTaskResult method with the test case task id
			taskResult, err := ts.GetTaskResult(context.TODO(), tc.taskID)
			// Check if the error matches the expectation
			require.ErrorIs(t, err, tc.err)
			// Check if the taskResult matches the expectation
			assert.Equal(t, tc.taskResult, taskResult)
		})
	}
}

//nolint: funlen
func TestProcessTask(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name         string      // The name of the test case
		task         domain.Task // The task to pass to the CreateTask method
		taskResult   *domain.TaskResult
		mockRepo     func() *task_mocks.TaskRepository
		mockTaskExec func() *exec_mocks.TaskExecutor
		err          error
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "Successful case",
			task: domain.Task{
				ID:     "5a12cde5-42ba-45da-a102-a279bafc66ae",
				Method: "GET",
				URL:    "https://example.com",
			},
			taskResult: &domain.TaskResult{
				TaskID:         "5a12cde5-42ba-45da-a102-a279bafc66ae",
				Status:         domain.TaskDone,
				HTTPStatusCode: http.StatusOK,
				ContentLength:  12,
				Body:           "Hello World!",
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("UpdateTaskResult", mock.Anything, mock.Anything).Return(nil)

				return repo
			},
			//nolint: dupl
			mockTaskExec: func() *exec_mocks.TaskExecutor {
				taskExec := &exec_mocks.TaskExecutor{}
				taskExec.On("ExecuteTask", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, task domain.Task) *domain.TaskResult {
						return &domain.TaskResult{
							TaskID:         task.ID,
							Status:         domain.TaskDone,
							HTTPStatusCode: http.StatusOK,
							ContentLength:  12,
							Body:           "Hello World!",
						}
					},
					func(ctx context.Context, task domain.Task) error {
						return nil
					},
				)

				return taskExec
			},
			err: nil,
		},
		{
			name: "Error from repository",
			task: domain.Task{
				ID:     "5a12cde5-42ba-45da-a102-a279bafc66ae",
				Method: "GET",
				URL:    "https://example.com",
			},
			taskResult: nil,
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("UpdateTaskResult", mock.Anything, mock.Anything).Return(errUpdateTaskInRepo)

				return repo
			},
			//nolint: dupl
			mockTaskExec: func() *exec_mocks.TaskExecutor {
				taskExec := &exec_mocks.TaskExecutor{}
				taskExec.On("ExecuteTask", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, task domain.Task) *domain.TaskResult {
						return &domain.TaskResult{
							TaskID:         task.ID,
							Status:         domain.TaskDone,
							HTTPStatusCode: http.StatusOK,
							ContentLength:  12,
							Body:           "Hello World!",
						}
					},
					func(ctx context.Context, task domain.Task) error {
						return nil
					},
				)

				return taskExec
			},
			err: errUpdateTaskInRepo,
		},
		{
			name: "Failed to execute task",
			task: domain.Task{
				ID:     "5a12cde5-42ba-45da-a102-a279bafc66ae",
				Method: "GET",
				URL:    "https://example.com",
			},
			taskResult: &domain.TaskResult{
				TaskID: "5a12cde5-42ba-45da-a102-a279bafc66ae",
				Status: domain.TaskError,
				Error:  errExecuteTask.Error(),
			},
			mockRepo: func() *task_mocks.TaskRepository {
				repo := &task_mocks.TaskRepository{}
				repo.On("UpdateTaskResult", mock.Anything, mock.Anything).Return(nil)

				return repo
			},
			//nolint: dupl
			mockTaskExec: func() *exec_mocks.TaskExecutor {
				taskExec := &exec_mocks.TaskExecutor{}
				taskExec.On("ExecuteTask", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, task domain.Task) *domain.TaskResult {
						return &domain.TaskResult{
							TaskID:         task.ID,
							Status:         domain.TaskDone,
							HTTPStatusCode: http.StatusOK,
							ContentLength:  12,
							Body:           "Hello World!",
						}
					},
					func(ctx context.Context, task domain.Task) error {
						return errExecuteTask
					},
				)

				return taskExec
			},
			err: errExecuteTask,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a new task service with the mock repository and queue
			ts := usecase.NewTaskService(tc.mockRepo(), nil, tc.mockTaskExec())
			// Call the ProcessTask method with the test case task
			taskResult, err := ts.ProcessTask(context.TODO(), tc.task)
			// Check if the error matches the expectation
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.taskResult, taskResult)
		})
	}
}
