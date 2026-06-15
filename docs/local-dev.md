# Local Development Guide

Run the full stack natively (no Docker required) using `make dev`.

---

## Prerequisites

| Tool | Minimum version | Install |
|------|-----------------|---------|
| Go | 1.21+ | https://go.dev/dl |
| Node.js | 20 LTS | https://nodejs.org |
| MongoDB | 7.x | https://www.mongodb.com/try/download/community |

Start MongoDB before launching the app. On macOS with Homebrew:

```sh
brew services start mongodb-community
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
→ MongoDB is not running. Start it with `brew services start mongodb-community` (macOS) or `sudo systemctl start mongod` (Linux).

**"JWT_SECRET not set" panic**  
→ Export `JWT_SECRET` before running `make dev-api`.

**Vite proxy returns 502**  
→ The API is not running. Start it with `make dev-api` in a separate terminal, or use `make dev` to run both together.

**Port already in use (8080 or 5173)**  
→ Find the process with `lsof -ti:8080 | xargs kill` and retry.
