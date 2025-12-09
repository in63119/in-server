# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/api /app/api

ENV PORT=:4000
EXPOSE 4000

CMD ["/app/api"]
