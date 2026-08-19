.PHONY: demo test build

# Dev convenience: reset bpmonitor.db and reseed it with fixture data, then
# launch the TUI against it. Lets you iterate on internal/tui without needing
# the real cuff nearby, and without worrying about having deleted/committed
# your way through the seeded readings in a previous run.
demo:
	rm -f bpmonitor.db
	go run ./cmd/seed
	go run ./cmd/bpterm

test:
	go test ./...

build:
	go build ./...
