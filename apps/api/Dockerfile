# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache module downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ── Run stage ────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS runner

COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]

# ── Dev stage (docker-compose.dev.yml only) ──────────────────────────────────
FROM golang:1.24-alpine AS dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go install github.com/air-verse/air@latest

EXPOSE 8080
