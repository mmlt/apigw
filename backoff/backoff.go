package backoff

import (
	"context"
	"time"
)

// TimeSleep typically equals time.Sleep(), only during testing it is mocked.
var TimeSleep func(d time.Duration)

func init() {
	TimeSleep = time.Sleep
}

// Backoff defines the parameters for exponential backoff.
// MaxRetries sets how many times the operation is retried.
type Backoff struct {
	ctx      context.Context
	MaxTries uint
	Duration time.Duration
	retry    uint
}

// New returns an initialized Backoff.
func New(ctx context.Context, maxTries uint, duration time.Duration) *Backoff {
	return &Backoff{
		ctx:      ctx,
		MaxTries: maxTries,
		Duration: duration,
	}
}

// BackoffFunc is the type of function that can retried.
type BackoffFunc func() error

// Do calls fn until it succeeds or MaxTries have been performed.
// If all attempts fail Backoff propagates the last error.
func (b *Backoff) Do(fn BackoffFunc) error {
	var err error
	for b.retry <= b.MaxTries {
		err = fn()
		if err == nil {
			break
		}

		m := b.retry
		if m > 8 {
			m = 8
		}
		//TODO implement ctx Done
		delay := b.Duration << m
		TimeSleep(delay)
		b.retry++
	}
	return err
}
