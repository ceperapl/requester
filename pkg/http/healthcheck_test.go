package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLivenessHandler tests the LivenessHandler method with different scenarios
func TestLivenessHandler(t *testing.T) {
	// Define a test case struct
	type testCase struct {
		name       string  // The name of the test case
		checks     []Check // The checks to pass to the AddLivenessChecks method
		wantStatus int     // The expected HTTP status code
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "All checks pass",
			checks: []Check{
				func() error { return nil },
				func() error { return nil },
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "One check fails",
			checks: []Check{
				func() error { return nil },
				func() error { return errors.New("check error") },
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "No checks",
			checks:     nil,
			wantStatus: http.StatusOK,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Create a new health check instance
			hc := NewHealthChecker()
			// Add the test case checks to the health check instance
			for _, check := range tc.checks {
				hc.AddLivenessChecks(check)
			}
			// Create a mock request and a response recorder
			req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
			rr := httptest.NewRecorder()
			// Call the LivenessHandler method with the mock request and response recorder
			hc.LivenessHandler().ServeHTTP(rr, req)
			// Check if the status code matches the expectation
			if status := rr.Code; status != tc.wantStatus {
				t.Errorf("LivenessHandler() status = %v, wantStatus %v", status, tc.wantStatus)
			}
		})
	}
}

// TestReadinessHandler tests the ReadinessHandler method with different scenarios
func TestReadinessHandler(t *testing.T) {
	// Define a test case struct
	type testCase struct {
		name       string  // The name of the test case
		checks     []Check // The checks to pass to the AddReadinessChecks method
		wantStatus int     // The expected HTTP status code
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "All checks pass",
			checks: []Check{
				func() error { return nil },
				func() error { return nil },
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "One check fails",
			checks: []Check{
				func() error { return nil },
				func() error { return errors.New("check error") },
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "No checks",
			checks:     nil,
			wantStatus: http.StatusOK,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Create a new health check instance
			hc := NewHealthChecker()
			// Add the test case checks to the health check instance
			for _, check := range tc.checks {
				hc.AddReadinessChecks(check)
			}
			// Create a mock request and a response recorder
			req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
			rr := httptest.NewRecorder()
			// Call the ReadinessHandler method with the mock request and response recorder
			hc.ReadinessHandler().ServeHTTP(rr, req)
			// Check if the status code matches the expectation
			if status := rr.Code; status != tc.wantStatus {
				t.Errorf("ReadinessHandler() status = %v, wantStatus %v", status, tc.wantStatus)
			}
		})
	}
}
