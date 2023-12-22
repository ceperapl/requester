package http

import (
	"net/http"
	"sync"
)

type Check func() error

// // HealthChecker is the interface that wraps the methods for health checks.
// type HealthChecker interface {
// 	AddLivenessChecks(check Check)
// 	AddReadinessChecks(check Check)
// 	LivenessHandler() http.HandlerFunc
// 	ReadinessHandler() http.HandlerFunc
// }

// NewHealthChecker creates instance of HealthChecker
func NewHealthChecker() *HealthCheck {
	return &HealthCheck{}
}

type HealthCheck struct {
	lock sync.RWMutex

	livenessChecks  []Check
	readinessChecks []Check
}

func (h *HealthCheck) LivenessHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := h.checkLiveness(); err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *HealthCheck) ReadinessHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := h.checkReadiness(); err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *HealthCheck) AddLivenessChecks(check Check) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.livenessChecks = append(h.livenessChecks, check)
}

func (h *HealthCheck) AddReadinessChecks(check Check) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.readinessChecks = append(h.readinessChecks, check)
}

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
