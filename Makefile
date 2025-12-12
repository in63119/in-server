BINARY ?= bin/api

.PHONY: build run prod fmt tidy test sync-abi

build: fmt
	GO111MODULE=on go build -o $(BINARY) ./cmd/api

run:
	@PATH="$(shell go env GOPATH)/bin:$$PATH" command -v air >/dev/null 2>&1 || \
		{ echo "air not found; install with: go install github.com/air-verse/air@latest"; exit 1; }
	PATH="$(shell go env GOPATH)/bin:$$PATH" air -c .air.toml

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
