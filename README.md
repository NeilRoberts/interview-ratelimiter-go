# Censys :: ratelimiter

This project creates a `ratelimiter` package that implements a rate limiter using the "token bucket" algorithm. The code is available on [GitHub](https://github.com/NeilRoberts/censys-ratelimiter).

## Usage

Create a new rate limiter with `ratelimiter.NewTokenBucketLimiter(burst int, period time.Duration)`. The bucket size is set by `burst` - this is the maximum number of immediately sequential operations that will be allowed. The token refill rate is set by both `burst` and `period` and is reduced to tokens/sec, regardless of the actual unit of `period`.

`Allow()` does a spot check of available tokens after first attempting to add tokens accumulated since the last refill. If there are any, one is removed and the method returns `true`. Otherwise, the method returns `false` and the token count is unchanged. i.e. it will not go negative.

`Wait(ctx context.Context)` requires a cancelable `Context` parameter. It confirms that `ctx` has not been canceled before checking the available tokens. If there are tokens in the bucket, the method returns and the caller can continue. Otherwise, the method waits until tokens are available before returning. Canceling `ctx` will result in the method returning an error.

## Examples

```go
// Create a new limiter w/ max rate of 10 ops/sec
limiter := NewTokenBucketLimiter(10, time.Second)

// Simple rate limiting in a loop
for i := 0; i < 20; i++ {
	if limiter.Allow() {
		fmt.Printf("Op %d allowed at %s\n", i, time.Now().Format("15:04:05.000"))
	} else {
		fmt.Printf("Op %d not allowed at %s\n", i, time.Now().Format("15:04:05.000"))
	}
	
	// Sleep for less than the refill rate
	time.Sleep(20 * time.Millisecond)
}


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
```

### Details

- The limiter accepts a max number of operations and a time period. These are converted into a rate of ops/sec on the premise that this averages over time to the specified ops/period. The max ops also determines the bucket size.
- The bucket is opportunistically refilled as each request is made, based on the time elapsed since the last refill. This is true for calls to both `Allow` and `Wait`.

## Algorithm Explained

The token bucket algorithm is based on the metaphor of spending tokens to perform tasks, such as processing (server) or sending (client) HTTP requests. For convenience, the tokens are kept in a bucket where they are easy to access, but limited in number. If the bucket has enough tokens in it, tasks can be performed. If not, tasks can be rejected or wait until tokens are available. More tokens are periodically added to the bucket, but the bucket can never hold more than some fixed number of tokens - extra tokens are discarded.

The limiter exposes two methods for interaction: `Allow` and `Wait`.

### Allow

`Allow` first attempts to add tokens to the bucket based on its refill rate and time since last addition. Then it checks whether there are sufficient tokens in the bucket to allow the operation. If there are, one token is removed and `Allow` returns `true`. If there aren't, `Allow` returns false and the caller is responsible for blocking the operation.

### Wait

`Wait` requires a cancelable context and first confirms that it has not been canceled. Then it checks for available tokens. Only if there are no tokens does it actually wait. Once tokens have been refilled, `Wait` returns to the caller. Canceling the context results in an error being returned to the caller.

## Comparison of Token Bucket and Fixed Window algorithms

Fixed Window is a simpler algorithm to implement. Within the time period of the window, a counter is decremented for each processed operation. If the counter reaches zero, subsequent operations are denied. At the beginning of the new window, the counter is reset to its limit. Compared to Token Bucket, Fixed Window can be less smooth in its rate of processing operations. All of the allowance can be consumed early, leading to starvation for the remainder of the window. Alternately, two maximum bursts on either side of the window reset can lead to an instantaneous processing rate that is double what was allotted.

Token Bucket continually allows more operations to be processed by partially refilling tokens based on the time elapsed since the last refill. A surge of operations may temporarily deplete tokens, but they will regenerate in time (that is likely less than the duration of a Fixed Window). On average, the expected processing rate will be achieved.

An example of where one algorithm might be chosen over another is a tiered service. Users on a free tier could be given a fixed number of operations per day (or hour, or whatever) and if they use them up, then they just have to wait until the next cycle. Users on a paid tier could be given more operations, but with a Token Bucket approach, the time waiting to process additional operations can appear shorter with a judicious choice of `burst` and `period`.

## Running and Testing

A Makefile has been provided for running and testing the code. `make run` will run the program that exercises the `ratelimiter` package. `make test` will run the package tests. Keep in mind that the various sleeps used to exercise time-evolving behavior results in the tests taking noticeable time (currently close to 20 seconds).
