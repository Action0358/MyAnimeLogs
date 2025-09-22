# Makefile for MyAnimeLogs Backend

.PHONY: migrate migrate-up migrate-down build run test clean

# Database migration commands
migrate:
	@echo "Running database migrations..."
	cd $(shell pwd) && go run scripts/migrate.go

migrate-up: migrate

migrate-down:
	@echo "Rolling back migrations is not implemented yet"

# Build commands
build:
	@echo "Building application..."
	go build -o bin/server ./cmd/server

# Run commands
run:
	@echo "Starting server..."
	go run ./cmd/server

run-migrate: migrate run

# Development commands
dev: migrate
	@echo "Starting development server with hot reload..."
	air

# Test commands
test:
	@echo "Running tests..."
	go test ./...

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Database commands
db-reset: migrate-down migrate-up

# Clean commands
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t myanimelogs-backend .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env.local myanimelogs-backend

docker-run-prod:
	@echo "Running Docker container with production environment..."
	docker run -p 8080:8080 --env-file .env.production myanimelogs-backend

# Deployment verification commands
verify-staging:
	@echo "Verifying staging environment..."
	./scripts/verify-deployment.sh staging

verify-production:
	@echo "Verifying production environment..."
	./scripts/verify-deployment.sh production

verify-all:
	@echo "Verifying all environments..."
	./scripts/verify-deployment.sh all

# Fly.io deployment commands
deploy-staging:
	@echo "Deploying to staging..."
	fly deploy -a myanimelogs-staging

deploy-production:
	@echo "Deploying to production..."
	fly deploy -a myanimelogs-production

start-staging:
	@echo "Starting staging environment..."
	fly scale count 1 -a myanimelogs-staging

start-production:
	@echo "Starting production environment..."
	fly scale count 1 -a myanimelogs-production

restart-staging:
	@echo "Restarting staging environment..."
	fly machine restart -a myanimelogs-staging

restart-production:
	@echo "Restarting production environment..."
	fly machine restart -a myanimelogs-production

stop-staging:
	@echo "Stopping staging environment..."
	fly machine stop -a myanimelogs-staging

stop-production:
	@echo "Stopping production environment..."
	fly machine stop -a myanimelogs-production

# Help
help:
	@echo "Available commands:"
	@echo "  migrate           - Run database migrations"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application"
	@echo "  run-migrate       - Run migrations then start server"
	@echo "  dev               - Start development server with hot reload"
	@echo "  test              - Run tests"
	@echo "  clean             - Clean build artifacts"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container with local environment"
	@echo "  docker-run-prod   - Run Docker container with production environment"
	@echo ""
	@echo "Deployment commands:"
	@echo "  verify-staging    - Verify staging environment"
	@echo "  verify-production - Verify production environment"
	@echo "  verify-all        - Verify all environments"
	@echo "  deploy-staging    - Deploy to staging"
	@echo "  deploy-production - Deploy to production"
	@echo "  start-staging     - Start staging environment"
	@echo "  start-production  - Start production environment"
	@echo "  stop-staging      - Stop staging environment"
	@echo "  stop-production   - Stop production environment"
	@echo "  help              - Show this help message"