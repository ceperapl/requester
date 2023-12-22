package utils_test

// import (
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // nolint: goerr113
// func TestRetry(t *testing.T) {
// 	t.Parallel()

// 	t.Run("success - one attempt", func(t *testing.T) {
// 		t.Parallel()
// 		stopCondition := true
// 		atteptCount := 0
// 		retryFunc := func() (bool, error) {
// 			atteptCount++
// 			return stopCondition, nil
// 		}

// 		err := Retry("test", 100*time.Millisecond, 10, retryFunc)
// 		require.NoError(t, err)
// 		assert.Equal(t, 1, atteptCount)
// 	})

// 	t.Run("success - five attempts", func(t *testing.T) {
// 		t.Parallel()
// 		attemptCount := 0
// 		retryFunc := func() (bool, error) {
// 			attemptCount++
// 			return attemptCount == 5, nil
// 		}

// 		err := Retry("test", 100*time.Millisecond, 10, retryFunc)
// 		require.NoError(t, err)
// 		assert.Equal(t, 5, attemptCount)
// 	})

// 	t.Run("fail - attempts exceeded", func(t *testing.T) {
// 		t.Parallel()
// 		attemptCount := 0
// 		expectedError := errors.New("retry func")
// 		retryFunc := func() (bool, error) {
// 			attemptCount++
// 			return attemptCount == 500, expectedError
// 		}

// 		err := Retry("test", 100*time.Millisecond, 10, retryFunc)
// 		require.Error(t, err)
// 		assert.Equal(t, 10, attemptCount)
// 		require.ErrorIs(t, err, expectedError)
// 	})
// }
