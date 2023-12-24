package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryhttp "github.com/ceperapl/requester/pkg/delivery/http"
	"github.com/ceperapl/requester/pkg/repository"
)

var errUnknown = errors.New("unknown error")

func TestErrorHandler(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name       string                            // The name of the test case
		handler    deliveryhttp.HandlerFuncWithError // The handler to pass to the errorHandler function
		wantStatus int                               // The expected HTTP status code
		wantBody   string                            // The expected HTTP response body
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "No error from handler",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return nil
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name: "Error from handler (ErrJSONUnmarshal)",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return deliveryhttp.ErrJSONUnmarshal
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"couldn't unmarshal json","httpStatusCode":400}`,
		},
		{
			name: "Error from handler (ErrTaskNotFound)",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return repository.ErrTaskNotFound
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"task not found","httpStatusCode":404}`,
		},
		{
			name: "Error from handler (unknown error)",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errUnknown
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"unknown error","httpStatusCode":500}`,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a mock request and a response recorder
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			// Call the errorHandler function with the mock request and response recorder
			deliveryhttp.ErrorHandler(tc.handler).ServeHTTP(rr, req)
			// Check if the status code matches the expectation
			if status := rr.Code; status != tc.wantStatus {
				t.Errorf("errorHandler() status = %v, wantStatus %v", status, tc.wantStatus)
			}
			// Check if the response body matches the expectation
			if body := rr.Body.String(); body != tc.wantBody {
				t.Errorf("errorHandler() body = %v, wantBody %v", body, tc.wantBody)
			}
		})
	}
}
