package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow() bool
	Wait(context.Context) error
}

type tokenBucketLimiter struct {
	// bucket capacity == max burst
	capacity float64

	// current token count - inits to bucket cap.
	tokens float64

	// token refill rate, as tokens/sec
	rate float64

	mu         sync.Mutex
	lastRefill time.Time
}

func NewTokenBucketLimiter(burst int, period time.Duration) *tokenBucketLimiter {
	rate := float64(burst) / period.Seconds()
	return &tokenBucketLimiter{
		capacity:   float64(burst),
		tokens:     float64(burst),
		rate:       rate,
		lastRefill: time.Now(),
	}
}

func (tbl *tokenBucketLimiter) Allow() bool {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	tbl.refill()
	if tbl.tokens >= 1 {
		tbl.tokens--
		return true
	}
	return false
}

func (tbl *tokenBucketLimiter) Wait(ctx context.Context) error {
	// Assertion: the context being canceled before we try to wait is an error
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for {
		// Check for available tokens, reusing existing behavior
		if tbl.Allow() {
			return nil
		}

		// Create a wait timer until there should be more tokens
		wait := tbl.timeToNextRefill()
		timer := time.Tick(wait)
		// Wait for either the context to be canceled or the timer to expire
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer:
		}
	}
}

func (tbl *tokenBucketLimiter) refill() {
	elapsed := time.Since(tbl.lastRefill).Seconds()
	refillTokens := elapsed * tbl.rate
	tbl.tokens += refillTokens
	if tbl.tokens > tbl.capacity {
		// Discard extra tokens
		tbl.tokens = tbl.capacity
	}
	tbl.lastRefill = time.Now()
}

func (tbl *tokenBucketLimiter) timeToNextRefill() time.Duration {
	tbl.mu.Lock()
	tokens := tbl.tokens
	rate := tbl.rate
	tbl.mu.Unlock()

	refillTokensNeeded := 1 - tokens
	waitTime := (refillTokensNeeded / rate) * float64(time.Second)

	return time.Duration(waitTime)
}
