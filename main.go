package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NeilRoberts/censys-ratelimiter/ratelimiter"
)

const OpsTraffic = 50

func main() {
	// Create a new limiter w/ max rate of 10 ops/sec
	limiter := ratelimiter.NewTokenBucketLimiter(10, time.Second)

	// Simple rate limiting in a loop
	for i := range OpsTraffic {
		if limiter.Allow() {
			fmt.Printf("Op %3d allowed at %s\n", i, time.Now().Format("15:04:05.000"))
		} else {
			fmt.Printf("Op %3d denied at %s\n", i, time.Now().Format("15:04:05.000"))
		}

		// Sleep for less than the refill rate
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Printf("\n\n")

	// Create a new limiter w/ max rate of 5 ops/sec
	waitLimiter := ratelimiter.NewTokenBucketLimiter(5, time.Second)
	ctx := context.Background()
	ctx, _ = context.WithCancel(ctx)

	// Burn through the tokens quickly and cause waiting
	var err error
	for j := range 10 {
		err = waitLimiter.Wait(ctx)
		if err != nil {
			fmt.Println("*** Unexpected error waiting for limiter: %w", err)
		} else {
			fmt.Printf("Op %3d allowed at %s\n", j, time.Now().Format("15:04:05.000"))
		}
	}
}
