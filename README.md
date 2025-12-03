# Gin starter

Minimal Gin REST starter that mirrors a NestJS-style layer split (handler/service/repository) and keeps config/env first.

## Quick start

```bash
# install deps
go mod tidy

# run dev
go run ./cmd/api

# build static binary
make build
```

## Structure

- `cmd/api/main.go` – app entry
- `internal/config` – env config loading
- `internal/server` – Gin engine + middleware + route wiring
- `internal/handler` – HTTP handlers (controllers)
- `pkg/logger` – zap logger helper

Default endpoints:
- `GET /api/health`
- `GET /api/ready`
