package http

import (
	"net/http"
	"sync"
)

// Check is a function type that performs a health check and returns an error if any.
type Check func() error

// NewHealthChecker creates and returns a new HealthCheck instance.
func NewHealthChecker() *HealthCheck {
	return &HealthCheck{}
}

// HealthCheck is a struct that implements the HealthChecker interface.
// It manages a list of liveness and readiness checks and provides HTTP handlers for them.
type HealthCheck struct {
	lock sync.RWMutex

	livenessChecks  []Check
	readinessChecks []Check
}

// LivenessHandler returns a http.HandlerFunc that performs the liveness checks and writes the result to the response writer.
// It writes http.StatusOK if all the checks pass, or http.StatusServiceUnavailable if any of them fails.
func (h *HealthCheck) LivenessHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := h.checkLiveness(); err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)

			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

// ReadinessHandler returns a http.HandlerFunc that performs the readiness checks and writes the result to the response writer.
// It writes http.StatusOK if all the checks pass, or http.StatusServiceUnavailable if any of them fails.
func (h *HealthCheck) ReadinessHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := h.checkReadiness(); err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)

			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

// AddLivenessChecks adds one or more liveness checks to the HealthCheck instance.
// Liveness checks are used to determine if the service is running and able to handle requests.
func (h *HealthCheck) AddLivenessChecks(check Check) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.livenessChecks = append(h.livenessChecks, check)
}

// AddReadinessChecks adds one or more readiness checks to the HealthCheck instance.
// Readiness checks are used to determine if the service is ready to serve traffic, such as having all the dependencies available.
func (h *HealthCheck) AddReadinessChecks(check Check) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.readinessChecks = append(h.readinessChecks, check)
}

// checkReadiness performs all the readiness checks and returns an error if any of them fails.
func (h *HealthCheck) checkReadiness() error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for _, check := range h.readinessChecks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

// checkLiveness performs all the liveness checks and returns an error if any of them fails.
func (h *HealthCheck) checkLiveness() error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for _, check := range h.livenessChecks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}
