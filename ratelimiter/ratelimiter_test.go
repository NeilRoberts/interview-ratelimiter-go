package ratelimiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test the constructor with a variety of periods
func TestNewTokenBucketLimiter(t *testing.T) {
	tests := []struct {
		name   string
		burst  int
		period time.Duration
		want   *tokenBucketLimiter
	}{
		{
			name:   "5 tokens per millisecond",
			burst:  5,
			period: time.Millisecond,
			want: &tokenBucketLimiter{
				capacity: float64(5),
				tokens:   float64(5),
				rate:     tokenRate(5, time.Millisecond),
			},
		},
		{
			name:   "5 tokens per second",
			burst:  5,
			period: time.Second,
			want: &tokenBucketLimiter{
				capacity: float64(5),
				tokens:   float64(5),
				rate:     tokenRate(5, time.Second),
			},
		},
		{
			name:   "5 tokens per minute",
			burst:  5,
			period: time.Minute,
			want: &tokenBucketLimiter{
				capacity: float64(5),
				tokens:   float64(5),
				rate:     tokenRate(5, time.Minute),
			},
		},
		{
			name:   "5 tokens per hour",
			burst:  5,
			period: time.Hour,
			want: &tokenBucketLimiter{
				capacity: float64(5),
				tokens:   float64(5),
				rate:     tokenRate(5, time.Hour),
			},
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.name)
		got := NewTokenBucketLimiter(tt.burst, tt.period)

		assert.Equal(t, tt.want.capacity, got.capacity, fmt.Sprintf("Mismatched capacity. Want: %f, Got: %f", tt.want.capacity, got.capacity))
		assert.Equal(t, tt.want.tokens, got.tokens, fmt.Sprintf("Mismatched tokens. Want: %f, Got: %f", tt.want.tokens, got.tokens))
		assert.Equal(t, tt.want.rate, got.rate, fmt.Sprintf("Mismatched capacity. Want: %f, Got: %f", tt.want.rate, got.rate))
	}
}

func TestAllow(t *testing.T) {
	tests := []struct {
		name        string
		burst       int
		period      time.Duration
		opSleep     time.Duration
		opVolume    int
		wantAllowed int
		wantDenied  int
	}{
		{
			name:        "50 ops @ 10/sec, no sleep",
			burst:       10,
			period:      time.Second,
			opSleep:     0 * time.Millisecond,
			opVolume:    50,
			wantAllowed: 10,
			wantDenied:  40,
		},
		{
			name:        "50 ops @ 10/sec, exact refresh sleep",
			burst:       10,
			period:      time.Second,
			opSleep:     100 * time.Millisecond,
			opVolume:    50,
			wantAllowed: 50,
			wantDenied:  0,
		},
		{
			name:        "50 ops @ 10/sec, half refresh sleep",
			burst:       10,
			period:      time.Second,
			opSleep:     50 * time.Millisecond,
			opVolume:    50,
			wantAllowed: 34,
			wantDenied:  16,
		},
		{
			name:        "50 ops @ 10/sec, greater than refresh sleep",
			burst:       10,
			period:      time.Second,
			opSleep:     200 * time.Millisecond,
			opVolume:    50,
			wantAllowed: 50,
			wantDenied:  0,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.name)

		lim := NewTokenBucketLimiter(tt.burst, tt.period)
		allowed := 0
		denied := 0
		for _ = range tt.opVolume {
			if lim.Allow() {
				allowed++
			} else {
				denied++
			}
			time.Sleep(tt.opSleep)
		}

		assert.Equal(t, tt.wantAllowed, allowed, fmt.Sprintf("Allowed ops - wanted: %3d, got: %3d", tt.wantAllowed, allowed))
		assert.Equal(t, tt.wantDenied, denied, fmt.Sprintf("Denied ops - wanted: %3d, got: %3d", tt.wantDenied, denied))
	}
}

func TestWait(t *testing.T) {
	burst := 5 // token refresh will be 200ms
	period := time.Second
	waitLimiter := NewTokenBucketLimiter(burst, period)

	opVolume := 10

	ctx := context.Background()
	ctx, _ = context.WithCancel(ctx)

	allowTimes := make([]time.Time, opVolume)

	// Exercise the Waiter
	for i := range opVolume {
		_ = waitLimiter.Wait(ctx)
		allowTimes[i] = time.Now()
	}

	// There should be effectively no delay for ops up to burst
	for j := 1; j < burst; j++ {
		delta := allowTimes[j].Sub(allowTimes[j-1])
		assert.Less(t, delta.Milliseconds(), int64(5), fmt.Sprintf("Pre-burst: expected delta < 5ms, got: %d", delta.Milliseconds()))
	}
	// but the delay should ~200ms from burst to the end
	for k := burst; k < opVolume; k++ {
		delta := allowTimes[k].Sub(allowTimes[k-1])
		assert.Less(t, delta.Milliseconds(), int64(205), fmt.Sprintf("Post-burst: expected delta < 205ms, got: %d", delta.Milliseconds()))
	}
}

// Helpers
func tokenRate(burst int, period time.Duration) float64 {
	return float64(burst) / period.Seconds()
}
