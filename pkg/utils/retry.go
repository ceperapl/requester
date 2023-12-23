package utils

import (
	"errors"
	"time"
)

// ErrRetryCountExceeded is an error that indicates that the maximum number of attempts for a retry function has been reached.
var ErrRetryCountExceeded = errors.New("retry count exceeded")

// RetryFunc is a function type that performs an action and returns a boolean value indicating whether to stop retrying and an error if any.
type RetryFunc func(attempt int) (stop bool, err error)

// Retry is a function that repeatedly calls a retry function with a given interval and a maximum number of attempts.
// It returns an error if the retry function fails after the maximum number of attempts or if it returns true.
func Retry(interval time.Duration, maxAttempts int, doFunc RetryFunc) error {
	var err error
	ticker := time.NewTicker(interval)
	attempt := 0
	for range ticker.C {
		attempt++
		if attempt > maxAttempts {
			return errors.Join(ErrRetryCountExceeded, err)
		}
		var stop bool
		stop, err = doFunc(attempt)
		if stop {
			break
		}
	}
	ticker.Stop()

	return nil
}
