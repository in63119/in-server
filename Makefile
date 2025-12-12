BINARY ?= bin/api

.PHONY: build run prod fmt tidy test sync-abi

build: fmt
	GO111MODULE=on go build -o $(BINARY) ./cmd/api

run:
	if command -v air >/dev/null 2>&1; then \
		PATH="$(shell go env GOPATH)/bin:$$PATH" air -c .air.toml; \
	else \
		echo "air not found, falling back to go run ./cmd/api"; \
		go run ./cmd/api; \
	fi

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
