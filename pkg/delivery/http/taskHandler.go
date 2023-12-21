package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/usecase"
	"github.com/ceperapl/requester/pkg/validation"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewTaskHandler(mux *mux.Router, taskService usecase.TaskService) (*mux.Router, error) {
	validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
	if err != nil {
		return nil, err
	}

	handler := &taskHandler{
		usecase:    taskService,
		validation: validation,
	}

	subRouter := mux.PathPrefix("/api/v1").Subrouter()

	subRouter.Handle("/task", logEndpoint(http.HandlerFunc(handler.CreateTask))).Methods(http.MethodPost)
	subRouter.Handle("/task/{id}", logEndpoint(http.HandlerFunc(handler.GetTaskResult))).Methods(http.MethodGet)

	return subRouter, nil
}

type taskHandler struct {
	usecase    usecase.TaskService
	validation validation.Validation
}

func (t *taskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task *domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: task validation

	taskId, err := t.usecase.CreateTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	taskJson, err := json.Marshal(domain.Task{ID: taskId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(taskJson))
}

func (t *taskHandler) GetTaskResult(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	taskResult, err := t.usecase.GetTaskResult(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskResultJson, err := json.Marshal(taskResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(taskResultJson))
}
