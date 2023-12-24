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

// handlerFuncWithError is a function type that handles HTTP requests and returns an error if any.
type handlerFuncWithError func(http.ResponseWriter, *http.Request) error

// Handle registers the HTTP handlers for the task service using the given mux router.
// It creates a new validator and a new task handler and sets up the subrouter for the API endpoints.
// It returns an error if it fails to create the validator or the task handler.
func Handle(mux *mux.Router, taskService usecase.TaskUsecaser) error {
	valid, err := validator.New(
		validator.WithJSONNamesForStructFields(),
	)
	if err != nil {
		return fmt.Errorf("couldn't create validator: %w", err)
	}

	handler := &taskHandler{
		usecase:   taskService,
		validator: *valid,
	}

	subRouter := mux.PathPrefix("/api/v1").Subrouter()
	ctx := context.Background()

	subRouter.Handle("/task", logEndpoint(errorHandler(handler.CreateTaskEndpoint(ctx)))).Methods(http.MethodPost)
	subRouter.Handle("/task/{id}", logEndpoint(errorHandler(handler.GetTaskResultEndpoint(ctx)))).Methods(http.MethodGet)

	return nil
}

type taskHandler struct {
	usecase   usecase.TaskUsecaser
	validator validator.Validation
}

func (t *taskHandler) CreateTaskEndpoint(ctx context.Context) handlerFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		task, err := t.decodeCreateTaskEndpointReq(r)
		if err != nil {
			return err
		}

		taskID, err := t.usecase.CreateTask(ctx, *task)
		if err != nil {
			//nolint: wrapcheck
			return err
		}

		if err := t.encodeCreateTaskEndpointResp(w, taskID); err != nil {
			return err
		}

		return nil
	}
}

func (t *taskHandler) decodeCreateTaskEndpointReq(req *http.Request) (*domain.Task, error) {
	var task *domain.Task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		return nil, errors.Join(ErrJSONUnmarshal, err)
	}
	defer req.Body.Close()

	if err := t.validator.ValidateStruct(task); err != nil {
		//nolint: wrapcheck
		return nil, err
	}

	return task, nil
}

func (t *taskHandler) encodeCreateTaskEndpointResp(w http.ResponseWriter, taskID string) error {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(domain.Task{ID: taskID}); err != nil {
		return errors.Join(ErrJSONMarshal, err)
	}

	return nil
}

func (t *taskHandler) GetTaskResultEndpoint(ctx context.Context) handlerFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		taskID, err := t.decodeGetTaskResultEndpointReq(r)
		if err != nil {
			return err
		}

		taskResult, err := t.usecase.GetTaskResult(ctx, taskID)
		if err != nil {
			//nolint: wrapcheck
			return err
		}

		if err := t.encodeGetTaskResultEndpointResp(w, taskResult); err != nil {
			return err
		}

		return nil
	}
}

func (t *taskHandler) decodeGetTaskResultEndpointReq(req *http.Request) (string, error) {
	id := mux.Vars(req)["id"]

	if err := t.validator.ValidateVar(id, "uuid4"); err != nil {
		//nolint: wrapcheck
		return "", err
	}

	return id, nil
}

func (t *taskHandler) encodeGetTaskResultEndpointResp(w http.ResponseWriter, taskResult *domain.TaskResult) error {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(taskResult); err != nil {
		return errors.Join(ErrJSONMarshal, err)
	}

	return nil
}
