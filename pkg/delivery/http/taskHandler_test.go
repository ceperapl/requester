package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	deliveryhttp "github.com/ceperapl/requester/pkg/delivery/http"
	"github.com/ceperapl/requester/pkg/domain"
	usecase_mocks "github.com/ceperapl/requester/pkg/usecase/mocks"
	valid_mocks "github.com/ceperapl/requester/pkg/validator/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errUseCase = errors.New("usecase error")
var errValidation = errors.New("some validation error")

//nolint: funlen
func TestCreateTaskEndpoint(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name          string // The name of the test case
		requestBody   string
		mockUsecase   func() *usecase_mocks.TaskUsecaser
		mockValidator func() *valid_mocks.Validator
		respStatus    int    // The expected HTTP status code
		respBody      string // The expected HTTP response body
	}

	// Define some test cases
	testCases := []testCase{
		{
			name:        "Valid task",
			requestBody: `{"method":"GET", "url":"https://httpbin.org/delay/7", "headers": {"User-Agent": ["Bing"]}}`,
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}
				usecase.On("CreateTask", mock.Anything, mock.Anything).Return("52a5922f-f892-45bb-9885-bc4b723e14c4", nil)

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateStruct", mock.Anything).Return(nil)

				return valid
			},
			respStatus: http.StatusOK,
			respBody:   `{"id":"52a5922f-f892-45bb-9885-bc4b723e14c4"}`,
		},
		{
			name:        "Validation error",
			requestBody: `{"method":"GET", "url":"https://httpbin.org/delay/7", "headers": {"User-Agent": ["Bing"]}}`,
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateStruct", mock.Anything).Return(errValidation)

				return valid
			},
			respStatus: http.StatusBadRequest,
			respBody:   `{"error":"couldn't decode request: request validation error\nsome validation error","httpStatusCode":400}`,
		},
		{
			name:        "Error from usecase",
			requestBody: `{"method":"GET", "url":"https://httpbin.org/delay/7", "headers": {"User-Agent": ["Bing"]}}`,
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}
				usecase.On("CreateTask", mock.Anything, mock.Anything).Return("", errUseCase)

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateStruct", mock.Anything).Return(nil)

				return valid
			},
			respStatus: http.StatusInternalServerError,
			respBody:   `{"error":"couldn't create task: usecase error","httpStatusCode":500}`,
		},
		{
			name:        "Invalid request body",
			requestBody: `Invalid request body`,
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateStruct", mock.Anything).Return()

				return valid
			},
			respStatus: http.StatusBadRequest,
			//nolint: lll
			respBody: `{"error":"couldn't decode request: couldn't unmarshal json\ninvalid character 'I' looking for beginning of value","httpStatusCode":400}`,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a new task handler with the mock usecase
			handler := &deliveryhttp.TaskHandler{
				UseCase:   tc.mockUsecase(),
				Validator: tc.mockValidator(),
			}

			// Create a mock request and a response recorder
			req := httptest.NewRequest(http.MethodPost, "/task", strings.NewReader(tc.requestBody))
			rr := httptest.NewRecorder()
			// Call the CreateTaskEndpoint method with the mock request and response recorder
			deliveryhttp.ErrorHandler(handler.CreateTaskEndpoint(context.Background())).ServeHTTP(rr, req)
			// Check if the status code matches the expectation
			assert.Equal(t, tc.respStatus, rr.Code)
			// Check if the response body matches the expectation
			assert.Equal(t, tc.respBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

//nolint: funlen
func TestGetTaskResultEndpoint(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name          string // The name of the test case
		id            string // The id to pass in the request URL
		mockUsecase   func() *usecase_mocks.TaskUsecaser
		mockValidator func() *valid_mocks.Validator
		respStatus    int    // The expected HTTP status code
		respBody      string // The expected HTTP response body
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "Successful case",
			id:   "52a5922f-f892-45bb-9885-bc4b723e14c4",
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}
				usecase.On("GetTaskResult", mock.Anything, mock.Anything).Return(
					func(ctx context.Context, id string) *domain.TaskResult {
						return &domain.TaskResult{
							TaskID:         id,
							Status:         domain.TaskDone,
							HTTPStatusCode: 200,
							ContentLength:  12,
							Body:           "Hello World!",
						}
					},
					func(ctx context.Context, id string) error {
						return nil
					},
				)

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateVar", mock.Anything, mock.Anything).Return(nil)

				return valid
			},
			respStatus: http.StatusOK,
			respBody:   `{"id":"52a5922f-f892-45bb-9885-bc4b723e14c4","status":"done","httpStatusCode":200,"length":12,"body":"Hello World!"}`,
		},
		{
			name: "Validation error",
			id:   "52a5922f-f892-45bb-9885-bc4b723e14c4",
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateVar", mock.Anything, mock.Anything).Return(errValidation)

				return valid
			},
			respStatus: http.StatusBadRequest,
			respBody:   `{"error":"couldn't decode request: request validation error\nsome validation error","httpStatusCode":400}`,
		},
		{
			name: "Error from usecase",
			id:   "52a5922f-f892-45bb-9885-bc4b723e14c4",
			mockUsecase: func() *usecase_mocks.TaskUsecaser {
				usecase := &usecase_mocks.TaskUsecaser{}
				usecase.On("GetTaskResult", mock.Anything, mock.Anything).Return(nil, errUseCase)

				return usecase
			},
			mockValidator: func() *valid_mocks.Validator {
				valid := &valid_mocks.Validator{}
				valid.On("ValidateVar", mock.Anything, mock.Anything).Return(nil)

				return valid
			},
			respStatus: http.StatusInternalServerError,
			respBody:   `{"error":"couldn't get task result: usecase error","httpStatusCode":500}`,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a new task handler with the mock usecase
			handler := &deliveryhttp.TaskHandler{
				UseCase:   tc.mockUsecase(),
				Validator: tc.mockValidator(),
			}
			// Create a mock request and a response recorder
			req := httptest.NewRequest(http.MethodGet, "/task/"+tc.id, nil)
			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.Handle("/task/{id}", deliveryhttp.ErrorHandler(handler.GetTaskResultEndpoint(context.Background())))
			// Call the GetTaskResultEndpoint method with the mock request and response recorder
			router.ServeHTTP(rr, req)
			// Check if the status code matches the expectation
			assert.Equal(t, tc.respStatus, rr.Code)
			// Check if the response body matches the expectation
			assert.Equal(t, tc.respBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}
