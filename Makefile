.PHONY: build run clean deps tidy

BINARY=server

## build: compile the server binary
build:
	go build -o $(BINARY) ./cmd/server

## run: build then run with sudo (required for cgroup access)
run: build
	sudo ./$(BINARY)

## deps: download all Go module dependencies
deps:
	go mod download

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove binary and temporary directories
clean:
	rm -f $(BINARY)
	rm -rf tmp/

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'
