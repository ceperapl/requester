package domain

type TaskStatus string

const (
	TaskNew        TaskStatus = "new"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskError      TaskStatus = "error"
)

type Task struct {
	ID      string              `bson:"_id" json:"id"`
	Method  string              `json:"method,omitempty"`
	URL     string              `json:"url,omitempty" validate:"url"`
	Headers map[string][]string `json:"headers,omitempty"`
}

type TaskResult struct {
	TaskID         string              `json:"id"`
	Status         TaskStatus          `json:"status"`
	HTTPStatusCode *int                `json:"httpStatusCode,omitempty"`
	Headers        map[string][]string `json:"headers,omitempty"`
	ContentLength  int64               `json:"length,omitempty"`
}
