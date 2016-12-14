// Package backoff is a slightly changed version of the backoff algorithm that comes with gRPC
// See: vendor/google.golang.org/grpc/LICENSE for the license
package backoff

import (
	"math/rand"
	"time"
)

// DefaultConfig is the default backoff configuration
var (
	DefaultConfig = Config{
		MaxDelay:  120 * time.Second,
		BaseDelay: 1.0 * time.Second,
		Factor:    1.6,
		Jitter:    0.2,
	}
)

// Config defines the parameters for backoff
type Config struct {
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration

	// BaseDelay is the amount of time to wait before retrying after the first failure.
	BaseDelay time.Duration

	// factor is applied to the backoff after each retry.
	Factor float64

	// jitter provides a range to randomize backoff delays.
	Jitter float64
}

// Backoff returns the delay for the current amount of retries
func (bc Config) Backoff(retries int) time.Duration {
	if retries == 0 {
		return bc.BaseDelay
	}
	backoff, max := float64(bc.BaseDelay), float64(bc.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.Factor
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + bc.Jitter*(rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}

// Backoff returns the delay for the current amount of retries
func Backoff(retries int) time.Duration {
	return DefaultConfig.Backoff(retries)
}
