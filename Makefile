.PHONY: build test vet check run demo seed clean

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

check: vet test build

run:
	go run ./cmd/bpterm

demo:
	rm -f dev.db
	go run ./cmd/seed -db dev.db
	go run ./cmd/bpterm -db dev.db

seed:
	rm -f dev.db
	go run ./cmd/seed -db dev.db

clean:
	rm -f dev.db dev.db-journal
