run:
	go run ./...

test:
	go test -v ./ratelimiter -count=1

.PHONY: run test
