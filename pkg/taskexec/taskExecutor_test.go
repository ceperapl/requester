package taskexec_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/taskexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint: funlen
func TestExecuteTask(t *testing.T) {
	t.Parallel()
	// Create a test server that returns different responses depending on the request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			//nolint: errcheck
			w.Write([]byte("Hello, world!"))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			//nolint: errcheck
			w.Write([]byte("Something went wrong"))
		default:
			w.WriteHeader(http.StatusNotFound)
			//nolint: errcheck
			w.Write([]byte("Not found"))
		}
	}))

	taskExecutor := taskexec.NewTaskExecution()

	tests := []struct {
		name       string
		task       domain.Task
		taskResult domain.TaskResult
		err        error
	}{
		{
			name: "Valid request",
			task: domain.Task{
				Method: "GET",
				URL:    testServer.URL + "/ok",
				Headers: map[string][]string{
					"User-Agent": {"Test"},
				},
			},
			taskResult: domain.TaskResult{
				Status:         domain.TaskDone,
				HTTPStatusCode: http.StatusOK,
				ContentLength:  13,
				Body:           "Hello, world!",
			},
			err: nil,
		},
		{
			name: "Invalid request",
			task: domain.Task{
				Method: "invalid method",
			},
			err: taskexec.ErrCreateRequest,
		},
		{
			name: "Invalid request",
			task: domain.Task{
				URL: "invalid url",
			},
			err: taskexec.ErrRequestExecution,
		},
		{
			name: "Server error",
			task: domain.Task{
				Method: "GET",
				URL:    testServer.URL + "/error",
			},
			taskResult: domain.TaskResult{
				Status:         domain.TaskDone,
				HTTPStatusCode: http.StatusInternalServerError,
				ContentLength:  20,
				Body:           "Something went wrong",
			},
			err: nil,
		},
		{
			name: "Not found",
			task: domain.Task{
				Method: "GET",
				URL:    testServer.URL + "/notfound",
			},
			taskResult: domain.TaskResult{
				Status:         domain.TaskDone,
				HTTPStatusCode: http.StatusNotFound,
				ContentLength:  9,
				Body:           "Not found",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			taskResult, err := taskExecutor.ExecuteTask(context.TODO(), &tt.task)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.taskResult.HTTPStatusCode, taskResult.HTTPStatusCode)
				assert.Equal(t, tt.taskResult.Body, taskResult.Body)
				assert.Equal(t, tt.taskResult.ContentLength, taskResult.ContentLength)
				assert.Equal(t, tt.taskResult.Status, taskResult.Status)
			}
		})
	}
}
