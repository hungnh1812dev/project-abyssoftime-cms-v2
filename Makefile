.PHONY: dev dev-api dev-web test-api test-web \
       mongo-start mongo-stop pg-start pg-stop pg-rm \
       graphql-generate

CONTAINER_CLI ?= docker

# ── MongoDB ──────────────────────────────────────────────────────────────────
MONGO_CONTAINER = cms-mongo
MONGO_IMAGE     = mongo:7
MONGO_PORT      = 27017
MONGO_VOLUME    = cms-mongo-data

mongo-start:
	@if $(CONTAINER_CLI) ps --format '{{.Names}}' | grep -q '^$(MONGO_CONTAINER)$$'; then \
		echo "MongoDB already running ($(MONGO_CONTAINER))"; \
	elif $(CONTAINER_CLI) ps -a --format '{{.Names}}' | grep -q '^$(MONGO_CONTAINER)$$'; then \
		$(CONTAINER_CLI) start $(MONGO_CONTAINER); \
		echo "MongoDB resumed ($(MONGO_CONTAINER))"; \
	else \
		$(CONTAINER_CLI) run -d \
			--name $(MONGO_CONTAINER) \
			-p $(MONGO_PORT):27017 \
			-v $(MONGO_VOLUME):/data/db \
			$(MONGO_IMAGE); \
		echo "MongoDB started ($(MONGO_CONTAINER)) on port $(MONGO_PORT)"; \
	fi

mongo-stop:
	$(CONTAINER_CLI) stop $(MONGO_CONTAINER)

# ── PostgreSQL ───────────────────────────────────────────────────────────────
PG_CONTAINER = cms-postgres
PG_IMAGE     = postgres:16-alpine
PG_PORT      = 5432
PG_VOLUME    = cms-postgres-data
PG_USER      = cms
PG_PASSWORD  = cms
PG_DB        = cms

pg-start:
	@if $(CONTAINER_CLI) ps --format '{{.Names}}' | grep -q '^$(PG_CONTAINER)$$'; then \
		echo "PostgreSQL already running ($(PG_CONTAINER))"; \
	elif $(CONTAINER_CLI) ps -a --format '{{.Names}}' | grep -q '^$(PG_CONTAINER)$$'; then \
		$(CONTAINER_CLI) start $(PG_CONTAINER); \
		echo "PostgreSQL resumed ($(PG_CONTAINER))"; \
	else \
		$(CONTAINER_CLI) run -d \
			--name $(PG_CONTAINER) \
			-p $(PG_PORT):5432 \
			-v $(PG_VOLUME):/var/lib/postgresql/data \
			-e POSTGRES_USER=$(PG_USER) \
			-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
			-e POSTGRES_DB=$(PG_DB) \
			$(PG_IMAGE); \
		echo "PostgreSQL started ($(PG_CONTAINER)) on port $(PG_PORT)"; \
		echo "DSN: postgres://$(PG_USER):$(PG_PASSWORD)@localhost:$(PG_PORT)/$(PG_DB)?sslmode=disable"; \
	fi

pg-stop:
	$(CONTAINER_CLI) stop $(PG_CONTAINER)

pg-rm:
	-$(CONTAINER_CLI) stop $(PG_CONTAINER)
	-$(CONTAINER_CLI) rm -v $(PG_CONTAINER)
	-$(CONTAINER_CLI) volume rm $(PG_VOLUME)

# ── Development ──────────────────────────────────────────────────────────────

# Start API and web in parallel; Ctrl-C kills both
# Run `make mongo-start` (and/or `make pg-start`) first
dev:
	@set -a; [ -f .env ] && . ./.env; set +a; \
		trap 'kill 0' INT; \
		(cd apps/api && go run ./cmd/server) & \
		(cd apps/web && npm run dev) & \
		wait

dev-api:
	@set -a; [ -f .env ] && . ./.env; set +a; cd apps/api && go run ./cmd/server

dev-web:
	cd apps/web && npm run dev

# ── Tests ────────────────────────────────────────────────────────────────────

test-api:
	cd apps/api && go test ./...

test-web:
	cd apps/web && npm test

# ── Code generation ──────────────────────────────────────────────────────────

graphql-generate:
	cd apps/api && go run ./cmd/gqlcodegen --phase=schema
	cd apps/api && go run github.com/99designs/gqlgen generate
	cd apps/api && rm -f graphql/resolver/*.resolvers.go
	cd apps/api && go run ./cmd/gqlcodegen --phase=resolvers
