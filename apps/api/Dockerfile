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

# ── Run stage (default) ──────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /server
COPY content-types/ /content-types/

EXPOSE 8080
CMD ["/server"]
