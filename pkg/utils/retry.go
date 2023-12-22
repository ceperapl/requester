package utils

import (
	"errors"
	"time"
)

var (
	ErrRetryCountExceeded = errors.New("retry count exceeded")
)

type RetryFunc func(attempt int) (stop bool, err error)

// Retry runs function until true.
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
