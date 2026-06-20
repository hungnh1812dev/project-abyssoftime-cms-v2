# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /src

# Cache module downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ── Run stage ────────────────────────────────────────────────────────────────
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /server
COPY content-types/ /content-types/

EXPOSE 8080
ENTRYPOINT ["/server"]

# ── Dev stage (docker-compose.dev.yml only) ──────────────────────────────────
FROM golang:1.26-alpine AS dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go install github.com/air-verse/air@latest

EXPOSE 8080
