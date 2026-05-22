.PHONY: run build test lint

run:
	go run ./cmd/spotify-tui

build:
	go build -o bin/spotify-tui ./cmd/spotify-tui

test:
	go test ./...

lint:
	golangci-lint run