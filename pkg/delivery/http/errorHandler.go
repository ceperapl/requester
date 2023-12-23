package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/validator"
)

// ErrorInfo represents the error information returned by the HTTP handler.
// It contains the error message and the HTTP status code.
type ErrorInfo struct {
	Error      string `json:"error"`
	StatusCode int    `json:"httpStatusCode"`
}

// errorHandler is a wrapper function that handles errors returned by the HTTP handler.
// It writes the error information as JSON to the response writer and sets the appropriate status code.
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

// statusCodeByError returns the HTTP status code corresponding to the given error.
// It uses the errors.Is function to check for specific errors and returns the appropriate code.
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
