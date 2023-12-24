package taskexec

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ceperapl/requester/pkg/domain"
)

const (
	// reqTimeout is the duration to wait for a HTTP request to complete.
	reqTimeout = 10 * time.Second
)

var (
	// ErrRequestExecution is the error for failing to execute a HTTP request.
	ErrRequestExecution = errors.New("request execution failed")

	// ErrCreateRequest is the error for failing to create a HTTP request.
	ErrCreateRequest = errors.New("couldn't create request")

	// ErrReadFromBody is the error for failing to read from a HTTP response body.
	ErrReadFromBody = errors.New("couldn't read from http response body")
)

// TaskExecutor is an interface that defines the ExecuteTask method,
// which takes a context and a task as arguments and returns a task result or an error.
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task domain.Task) (*domain.TaskResult, error)
}

// TaskExecution is a struct that implements the TaskExecutor interface.
// It has a HTTPClient field that is used to send HTTP requests.
type TaskExecution struct {
	HTTPClient *http.Client
}

// NewTaskExecution is a constructor function that returns a pointer to a new TaskExecution instance.
func NewTaskExecution() *TaskExecution {
	httpClient := &http.Client{
		Timeout: reqTimeout,
	}
	taskExec := TaskExecution{
		HTTPClient: httpClient,
	}

	return &taskExec
}

// ExecuteTask creates a new HTTP request with the given method, url, body and headers,
// sends it using the HTTPClient field, reads and closes the response body,
// and creates a new task result with the relevant data.
func (te *TaskExecution) ExecuteTask(ctx context.Context, task domain.Task) (*domain.TaskResult, error) {
	// Create a new http.Request with the given method, url, body and headers
	req, err := http.NewRequestWithContext(ctx, task.Method, task.URL, strings.NewReader(task.Body))
	if err != nil {
		return nil, errors.Join(ErrCreateRequest, err)
	}
	for key, values := range task.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	// Send the request and get the response
	resp, err := te.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Join(ErrRequestExecution, err)
	}
	// Read and close the response body
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(ErrReadFromBody, err)
	}

	respHeaders := make(map[string][]string)
	for key, value := range resp.Header {
		respHeaders[key] = value
	}

	taskResult := domain.TaskResult{
		TaskID:         task.ID,
		Status:         domain.TaskDone,
		HTTPStatusCode: resp.StatusCode,
		Headers:        respHeaders,
		ContentLength:  resp.ContentLength,
		Body:           string(data),
	}

	return &taskResult, nil
}
