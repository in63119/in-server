BINARY ?= bin/api

.PHONY: build run prod fmt tidy test sync-abi

build: fmt
	GO111MODULE=on go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

prod:
	docker compose up --build api

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

tidy:
	go mod tidy

test:
	go test ./...

sync-abi:
	go run ./scripts/sync-abi.go
