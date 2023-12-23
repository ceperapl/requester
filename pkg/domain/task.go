package domain

// TaskStatus is a string type for storing task execution status.
// It can have one of the following values: TaskNew, TaskInProgress, TaskDone, or TaskError.
type TaskStatus string

// Task execution statuses.
const (
	TaskNew        TaskStatus = "new"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskError      TaskStatus = "error"
)

// Task represents a HTTP request that can be executed by a worker.
// It contains the ID of the task, the HTTP method, the URL, and the headers.
// The Method and URL fields are validated using struct tags.
type Task struct {
	ID      string              `json:"id"`
	Method  string              `json:"method,omitempty"  validate:"oneof=GET HEAD POST PUT PATCH DELETE CONNECT OPTIONS TRACE"`
	URL     string              `json:"url,omitempty"     validate:"url"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID         string              `json:"id"`
	Status         TaskStatus          `json:"status"`
	HTTPStatusCode int                 `json:"httpStatusCode,omitempty"`
	Headers        map[string][]string `json:"headers,omitempty"`
	ContentLength  int64               `json:"length,omitempty"`
	Body           string              `json:"body,omitempty"`
	Error          string              `json:"error,omitempty"`
}
