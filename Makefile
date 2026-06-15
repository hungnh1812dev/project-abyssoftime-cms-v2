.PHONY: dev dev-api dev-web test-api test-web

# Start API and web in parallel; Ctrl-C kills both
dev:
	@trap 'kill 0' INT; \
		(cd apps/api && go run ./cmd/server) & \
		(cd apps/web && npm run dev) & \
		wait

# Start only the Go API server (requires MongoDB running locally or via docker-compose)
dev-api:
	cd apps/api && go run ./cmd/server

# Start only the Vite dev server
dev-web:
	cd apps/web && npm run dev

# Run Go unit tests
test-api:
	cd apps/api && go test ./...

# Run frontend tests
test-web:
	cd apps/web && npm test
