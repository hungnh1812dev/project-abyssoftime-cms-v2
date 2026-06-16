# Local Development Guide

Run API + web natively (Go + Node directly) with MongoDB in a container. No full Docker Compose stack required.

---

## Prerequisites

| Tool | Minimum version | Install |
|------|-----------------|---------|
| Go | 1.21+ | https://go.dev/dl |
| Node.js | 20 LTS | https://nodejs.org |
| Docker **or** Podman | any recent | https://docs.docker.com/get-docker / https://podman.io/getting-started/installation |

MongoDB runs as a container — no native install needed.

---

## Start MongoDB

Use the Makefile target (defaults to `docker`; override with `CONTAINER_CLI=podman` if needed):

```sh
make mongo-start
# or with Podman:
CONTAINER_CLI=podman make mongo-start
```

This creates a named container `cms-mongo` on port `27017` with a persistent volume `cms-mongo-data`. The command is **idempotent** — safe to run again if the container already exists.

To stop MongoDB when you're done:

```sh
make mongo-stop
# or:
CONTAINER_CLI=podman make mongo-stop
```

### Manual alternative (without make)

```sh
# Docker
docker run -d --name cms-mongo -p 27017:27017 -v cms-mongo-data:/data/db mongo:7

# Podman
podman run -d --name cms-mongo -p 27017:27017 -v cms-mongo-data:/data/db mongo:7
```

---

## Environment configuration

Copy the example file and fill in your values:

```sh
cp .env.example .env
```

Required variables (see `.env.example` for the full list):

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API listen port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/cms` |
| `JWT_SECRET` | Secret for signing JWTs | *(no default — must be set)* |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary account name | *(required for media upload)* |
| `CLOUDINARY_API_KEY` | Cloudinary API key | *(required for media upload)* |
| `CLOUDINARY_API_SECRET` | Cloudinary API secret | *(required for media upload)* |
| `VITE_API_URL` | API base URL for the Vite proxy | `http://localhost:8080` |

Export the API vars before running (or use [direnv](https://direnv.net)):

```sh
export PORT=8080
export MONGODB_URI=mongodb://localhost:27017/cms
export JWT_SECRET=your-secret-here
# Cloudinary vars only needed for media upload
export CLOUDINARY_CLOUD_NAME=...
export CLOUDINARY_API_KEY=...
export CLOUDINARY_API_SECRET=...
```

---

## Quick start

```sh
# Install frontend dependencies (first time only)
cd apps/web && npm install && cd ../..

# Start API + web concurrently
make dev
```

The web app opens at **http://localhost:5173**.  
The API is at **http://localhost:8080**.

`make dev` runs both processes in parallel and kills both when you press **Ctrl-C**.

---

## Individual targets

| Command | What it does |
|---------|--------------|
| `make dev` | API + web in parallel (Ctrl-C kills both) |
| `make dev-api` | Go API only (`go run ./cmd/server`) |
| `make dev-web` | Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |

---

## Docker mode

Prefer Docker? Use the compose files instead:

```sh
# First run (builds images)
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Subsequent runs
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

See the root `SPEC.md` §2 for the full command reference.

---

## Troubleshooting

**API exits immediately with "mongodb connect" error**  
→ MongoDB container is not running. Start it with `make mongo-start` (or `CONTAINER_CLI=podman make mongo-start`). Verify it's up with `docker ps | grep cms-mongo`.

**"JWT_SECRET not set" panic**  
→ Export `JWT_SECRET` before running `make dev-api`.

**Vite proxy returns 502**  
→ The API is not running. Start it with `make dev-api` in a separate terminal, or use `make dev` to run both together.

**Port already in use (8080 or 5173)**  
→ Find the process with `lsof -ti:8080 | xargs kill` and retry.
