# Requirements

## Language

Go

## Expected Time

~2 hours

Implement a rate limiting package that utilizes a token bucket strategy. Your implementation should have the following capabilities:

1. Configurable Limits
    - Allow specifying a maximum number of operations per time period
    - Support different time windows (per second, per minute, etc.)
2. Multiple Operation Modes
    - `Allow()` - Non-blocking check that returns true/false
    - `Wait(ctx context.Context)` - Blocks until allowed or context cancelled
3. Thread Safety
    - Must be safe for concurrent use by multiple goroutines
    - Properly handle race conditions
4. Multiple Instances
    - Users should be able to create multiple independent rate limiters
    - Each limiter maintains its own state
  
### Your README must explain:
- Explain how the token bucket algorithm works.
- Pick ONE of these alternatives and compare it to token bucket:
    - Leaky Bucket
    - Fixed Window Counter
    - Sliding Window Counter
- What trade-offs does the other algorithm make?
- How does it compare to token bucket?
- What use cases suit that algorithm better vs token bucket?
- Usage example for your package

## Deliverables

Submit the following:
  - Source Code
  - Tests
  - README.md (Important!)

### Notes

Add any assumption you make into the Readme. Use the package name `ratelimiter` to help simplify the testing of your solution please.
