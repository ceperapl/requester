package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/usecase"
	"github.com/ceperapl/requester/pkg/validator"
	"github.com/gorilla/mux"
)

// ErrCreateTaskHandler is an error that indicates a failure to create a task handler.
var ErrCreateTaskHandler = errors.New("failed to create task handler")

// ErrJSONUnmarshal is an error that indicates a failure to unmarshal JSON data.
var ErrJSONUnmarshal = errors.New("couldn't unmarshal json")

// ErrJSONMarshal is an error that indicates a failure to marshal JSON data.
var ErrJSONMarshal = errors.New("couldn't marshal json")

// ErrReqValidation is an error that indicates a failure to validate the request data.
var ErrReqValidation = errors.New("request validation error")

// Handle registers the HTTP handlers for the task service using the given mux router.
func Handle(mux *mux.Router, taskService usecase.TaskUsecaser, valid validator.Validator) error {
	handler := &TaskHandler{
		UseCase:   taskService,
		Validator: valid,
	}

	subRouter := mux.PathPrefix("/api/v1").Subrouter()
	ctx := context.Background()

	subRouter.Handle("/task", logEndpoint(ErrorHandler(handler.CreateTaskEndpoint(ctx)))).Methods(http.MethodPost)
	subRouter.Handle("/task/{id}", logEndpoint(ErrorHandler(handler.GetTaskResultEndpoint(ctx)))).Methods(http.MethodGet)

	return nil
}

// TaskHandler is a struct that implements the HTTP handler methods for the task service.
type TaskHandler struct {
	UseCase   usecase.TaskUsecaser
	Validator validator.Validator
}

// CreateTaskEndpoint returns a HandlerFuncWithError that handles the creation of a new task.
func (t *TaskHandler) CreateTaskEndpoint(ctx context.Context) HandlerFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		task, err := t.decodeCreateTaskEndpointReq(r)
		if err != nil {
			return fmt.Errorf("couldn't decode request: %w", err)
		}

		taskID, err := t.UseCase.CreateTask(ctx, *task)
		if err != nil {
			return fmt.Errorf("couldn't create task: %w", err)
		}

		if err := t.encodeCreateTaskEndpointResp(w, taskID); err != nil {
			return fmt.Errorf("couldn't encode response: %w", err)
		}

		return nil
	}
}

func (t *TaskHandler) decodeCreateTaskEndpointReq(req *http.Request) (*domain.Task, error) {
	var task *domain.Task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		return nil, errors.Join(ErrJSONUnmarshal, err)
	}
	defer req.Body.Close()

	if err := t.Validator.ValidateStruct(task); err != nil {
		//nolint: wrapcheck
		return nil, err
	}

	return task, nil
}

func (t *TaskHandler) encodeCreateTaskEndpointResp(w http.ResponseWriter, taskID string) error {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(domain.Task{ID: taskID}); err != nil {
		return errors.Join(ErrJSONMarshal, err)
	}

	return nil
}

// GetTaskResultEndpoint returns a HandlerFuncWithError that handles the retrieval of a task result.
func (t *TaskHandler) GetTaskResultEndpoint(ctx context.Context) HandlerFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		taskID, err := t.decodeGetTaskResultEndpointReq(r)
		if err != nil {
			return fmt.Errorf("couldn't decode request: %w", err)
		}

		taskResult, err := t.UseCase.GetTaskResult(ctx, taskID)
		if err != nil {
			return fmt.Errorf("couldn't get task result: %w", err)
		}

		if err := t.encodeGetTaskResultEndpointResp(w, taskResult); err != nil {
			return fmt.Errorf("couldn't encode response: %w", err)
		}

		return nil
	}
}

func (t *TaskHandler) decodeGetTaskResultEndpointReq(req *http.Request) (string, error) {
	id := mux.Vars(req)["id"]

	if err := t.Validator.ValidateVar(id, "uuid4"); err != nil {
		//nolint: wrapcheck
		return "", err
	}

	return id, nil
}

func (t *TaskHandler) encodeGetTaskResultEndpointResp(w http.ResponseWriter, taskResult *domain.TaskResult) error {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(taskResult); err != nil {
		return errors.Join(ErrJSONMarshal, err)
	}

	return nil
}
