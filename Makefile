BINARY ?= bin/api

.PHONY: build run fmt tidy test

build: fmt
	GO111MODULE=on go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

tidy:
	go mod tidy

test:
	go test ./...
