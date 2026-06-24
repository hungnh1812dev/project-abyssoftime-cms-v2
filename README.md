# project-abyssoftime-cms-v2

A lightweight, code-first Personal Headless CMS. Go backend (Clean Architecture, PostgreSQL, REST) + React/Vite frontend. Self-hosted on Render.com.

---

## Local Development

Two modes are supported: **Native** (Go + Node directly on your machine) and **Docker** (everything in containers).

### Native (recommended for day-to-day development)

**Prerequisites:** Go 1.21+, Node.js 20 LTS, Docker or Podman (for PostgreSQL).

```sh
# Install frontend dependencies (first time)
cd apps/web && npm install && cd ../..

# Copy and fill in env vars
cp .env.example .env   # set JWT_SECRET, CLOUDINARY_*, etc.

# Start PostgreSQL container (idempotent — safe to rerun)
make postgres-start
# Podman: CONTAINER_CLI=podman make postgres-start

# Start API + web concurrently (Ctrl-C stops both)
make dev
```

The web app runs at **http://localhost:5173** and the API at **http://localhost:8080**.

| Command               | Description                |
| --------------------- | -------------------------- |
| `make postgres-start` | Start PostgreSQL container |
| `make postgres-stop`  | Stop PostgreSQL container  |
| `make dev`            | API + web in parallel      |
| `make dev-api`        | Go API only                |
| `make dev-web`        | Vite dev server only       |
| `make test-api`       | `go test ./...`            |
| `make test-web`       | `vitest run`               |

See [docs/local-dev.md](docs/local-dev.md) for full setup instructions, env var reference, and troubleshooting.

### Docker

```sh
# First run (builds images + starts all services with hot reload)
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Subsequent runs
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

---

## Adding a content panel

Panels are hard-coded React pages — no drag-and-drop engine. See [docs/guide.md](docs/guide.md) for a step-by-step walkthrough.

---

## Tech stack

| Layer    | Technology                                                  |
| -------- | ----------------------------------------------------------- |
| Backend  | Go, Chi router, Clean Architecture                          |
| Database | PostgreSQL (separate service on Render.com)                 |
| Auth     | JWT (access + HttpOnly refresh cookie)                      |
| Frontend | React, Vite, TanStack Query, react-hook-form, Shadcn UI     |
| Media    | Cloudinary                                                  |
| Deploy   | Render.com (Docker, PostgreSQL as separate managed service) |
