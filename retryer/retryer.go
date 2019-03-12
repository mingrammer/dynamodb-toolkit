package retryer

import (
	"math"
	"time"
)

const (
	minRetryBackoff = 64 * time.Millisecond
	maxRetryBackoff = 5 * time.Second
)

// RetryBackoff returns exponential retry backoff duration
func RetryBackoff(attempts int) time.Duration {
	backoff := time.Duration(math.Pow(2, float64(attempts))) * minRetryBackoff
	if backoff > maxRetryBackoff {
		return maxRetryBackoff
	}
	return backoff
}
