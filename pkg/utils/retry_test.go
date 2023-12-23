package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ceperapl/requester/pkg/utils"
	"github.com/stretchr/testify/require"
)

var errSomethingWentWrong = errors.New("something went wrong")

// TestRetrySuccess tests that the Retry function returns nil when the doFunc succeeds.
func TestRetrySuccess(t *testing.T) {
	t.Parallel()
	// Define a doFunc that returns true and nil after 3 attempts.
	doFunc := func(attempt int) (bool, error) {
		if attempt == 3 {
			return true, nil
		}

		return false, nil
	}

	// Call the Retry function with a 10ms interval and 5 max attempts.
	err := utils.Retry(10*time.Millisecond, 5, doFunc)

	// Check that the error is nil.
	require.NoError(t, err)
}

// TestRetryFailure tests that the Retry function returns an error when the doFunc fails.
func TestRetryFailure(t *testing.T) {
	t.Parallel()

	// Define a doFunc that returns false and an error after 3 attempts.
	doFunc := func(attempt int) (bool, error) {
		if attempt == 3 {
			return false, errSomethingWentWrong
		}

		return false, nil
	}

	// Call the Retry function with a 10ms interval and 5 max attempts.
	err := utils.Retry(10*time.Millisecond, 5, doFunc)

	// Check that the error is not nil and contains the expected message.
	require.ErrorIs(t, err, utils.ErrRetryCountExceeded)
}
