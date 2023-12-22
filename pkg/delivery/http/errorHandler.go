package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/validator"
)

type ErrorInfo struct {
	Error      string `json:"error"`
	StatusCode int    `json:"httpStatusCode"`
}

func errorHandler(handler handlerFuncWithError) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}
		statusCode := statusCodeByError(err)
		errorInfo := ErrorInfo{
			Error:      err.Error(),
			StatusCode: statusCode,
		}
		errorInfoJSON, _ := json.Marshal(errorInfo)
		w.WriteHeader(statusCodeByError(err))
		fmt.Fprint(w, string(errorInfoJSON))
	})
}

func statusCodeByError(err error) int {
	switch {
	case errors.Is(err, ErrJSONUnmarshal), errors.Is(err, validator.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrTaskNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
