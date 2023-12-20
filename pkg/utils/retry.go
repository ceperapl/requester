package utils

import (
	"fmt"
	"time"
)

type RetryFunc func() (stop bool, err error)

// Retry runs function until true
func Retry(operation string, interval time.Duration, maxAttempts int, doFunc RetryFunc) error {
	var err error
	ticker := time.NewTicker(interval)
	currentAttempt := 0
	for range ticker.C {
		currentAttempt++
		if currentAttempt > maxAttempts {
			return fmt.Errorf("number of attempts (%d) to %s exceeded: %w", maxAttempts, operation, err)
		}
		var stop bool
		stop, err = doFunc()
		if stop {
			break
		}
	}
	ticker.Stop()
	return nil
}
