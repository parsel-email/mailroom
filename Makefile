# Simple Makefile for a Go project

REPO=github.com/parsel-email/mailroom
CONTAINER_REGISTRY=ghcr.io/parsel-email/mailroom
export CGO_ENABLED=1
# Get the current git branch
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
# Build flags for SQLite extensions

# Build the application
all: build test

build:
	@echo "Building with SQLite extensions"
	go build -o main main.go

# Run the application
run:
	go run main.go

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
	go test ./... -v

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main migrate

# Database migration commands
db-migrate:
	@echo "Running migrations..."
	go run cmd/migrate/main.go up

db-rollback:
	@echo "Rolling back migrations..."
	go run cmd/migrate/main.go down

db-status:
	@echo "Migration status..."
	go run cmd/migrate/main.go version

db-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Use 'make db-new name=migration_name'"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	go run cmd/migrate/main.go new $(name)

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

sqlc:
	sqlc generate

air:
	docker compose up