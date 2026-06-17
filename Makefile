.PHONY: dev dev-api dev-web test-api test-web mongo-start mongo-stop graphql-generate

CONTAINER_CLI ?= docker
MONGO_CONTAINER = cms-mongo
MONGO_IMAGE     = mongo:7
MONGO_PORT      = 27017
MONGO_VOLUME    = cms-mongo-data

# Start MongoDB container (idempotent — skips if already running)
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

# Stop the MongoDB container
mongo-stop:
	$(CONTAINER_CLI) stop $(MONGO_CONTAINER)

# Start API and web in parallel; Ctrl-C kills both
# Run `make mongo-start` first if MongoDB is not already up
dev:
	@set -a; [ -f .env ] && . ./.env; set +a; \
		trap 'kill 0' INT; \
		(cd apps/api && go run ./cmd/server) & \
		(cd apps/web && npm run dev) & \
		wait

# Start only the Go API server
dev-api:
	@set -a; [ -f .env ] && . ./.env; set +a; cd apps/api && go run ./cmd/server

# Start only the Vite dev server
dev-web:
	cd apps/web && npm run dev

# Run Go unit tests
test-api:
	cd apps/api && go test ./...

# Run frontend tests
test-web:
	cd apps/web && npm test

# Regenerate GraphQL boilerplate from graphql/schema.graphqls (never hand-edit graphql/generated/)
graphql-generate:
	cd apps/api && go run github.com/99designs/gqlgen generate
