package backoff

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	ErrTest = fmt.Errorf("test")
)

// TestBackoff shows that backoff is taking exponentially more time after each retry and finally fails if more than
// MaxTries are needed.
func TestBackoff(t *testing.T) {
	tests := []struct {
		backoff           *Backoff
		respondAfterTries int
		wantError         error
		wantDuration      time.Duration
		comment           string
	}{
		{
			backoff:           New(context.Background(), 0, time.Second),
			respondAfterTries: 0,
			wantError:         nil,
			wantDuration:      0,
			comment:           "normal call",
		},
		{
			backoff:           New(context.Background(), 2, time.Second),
			respondAfterTries: 1,
			wantError:         nil,
			wantDuration:      1 * time.Second,
			comment:           "1 retry",
		},
		{
			backoff:           New(context.Background(), 2, time.Second),
			respondAfterTries: 2,
			wantError:         nil,
			wantDuration:      2 * time.Second,
			comment:           "2 retries",
		},
		{
			backoff:           New(context.Background(), 2, time.Second),
			respondAfterTries: 3,
			wantError:         ErrTest,
			wantDuration:      4 * time.Second,
			comment:           "no success after MaxTries",
		},
		{
			backoff:           New(context.Background(), 99, time.Second),
			respondAfterTries: 9,
			wantError:         nil,
			wantDuration:      256 * time.Second,
			comment:           "9 retries",
		},
		{
			backoff:           New(context.Background(), 99, time.Second),
			respondAfterTries: 10,
			wantError:         nil,
			wantDuration:      256 * time.Second,
			comment:           "no increase in backoff time after 9 retries",
		},
	}

	// Mock TimeSleep during this test.
	bu := TimeSleep
	defer func() {
		TimeSleep = bu
	}()
	var gotDuration time.Duration
	TimeSleep = func(d time.Duration) {
		gotDuration = d
	}

	for _, tst := range tests {
		tries := 0
		retries := tst.respondAfterTries
		gotErr := tst.backoff.Do(func() error {
			tries++
			retries--
			if retries >= 0 {
				return ErrTest
			}
			return nil
		})
		assert.Equal(t, tst.wantError, gotErr, tst.comment)
		assert.Equal(t, tst.wantDuration, gotDuration, tst.comment)
	}
}

//TODO TestBackoffCancel shows that a Backoff can be cancelled.
