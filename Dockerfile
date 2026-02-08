# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api

# --- Runtime stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./api"]
