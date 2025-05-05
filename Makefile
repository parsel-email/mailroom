# Simple Makefile for a Go project

REPO=github.com/parsel-email/mailroom
CONTAINER_REGISTRY=ghcr.io/parsel-email/mailroom
# Get the current git branch
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
# Build flags for SQLite extensions
BUILD_FLAGS=-tags=sqlite_fts5,sqlite_json1

# Build the application
all: build test

build:
	@echo "Building with SQLite extensions"
	@CGO_ENABLED=1 go build $(BUILD_FLAGS) -o main main.go

# Run the application
run:
	@CGO_ENABLED=1 go run $(BUILD_FLAGS) main.go

# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@CGO_ENABLED=1 go test $(BUILD_FLAGS) ./... -v

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@CGO_ENABLED=1 go test $(BUILD_FLAGS) ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main migrate

# Database migration commands
db-migrate:
	@echo "Running migrations..."
	@CGO_ENABLED=1 go run $(BUILD_FLAGS) cmd/migrate/main.go up

db-rollback:
	@echo "Rolling back migrations..."
	@CGO_ENABLED=1 go run $(BUILD_FLAGS) cmd/migrate/main.go down

db-status:
	@echo "Migration status..."
	@CGO_ENABLED=1 go run $(BUILD_FLAGS) cmd/migrate/main.go version

db-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Use 'make db-new name=migration_name'"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@CGO_ENABLED=1 go run $(BUILD_FLAGS) cmd/migrate/main.go new $(name)

# Test SQLite extensions
db-test:
	@echo "Testing SQLite extensions (FTS5, JSON)..."
	@CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/dbtest cmd/dbtest/main.go
	@./bin/dbtest

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down itest db-migrate db-rollback db-status db-new db-test

container-push:
	docker buildx build \
	-f Dockerfile.buildx \
  --platform linux/amd64,linux/arm64 \
  --build-arg GH_TOKEN=${GITHUB_TOKEN} \
  --build-arg COMMIT_SHA=$(shell git rev-parse HEAD) \
  --build-arg BRANCH=${BRANCH} \
  -t ${CONTAINER_REGISTRY}:latest \
  --push \
  .

client-build:
	cd client && npm run build