.PHONY: build run clean test deps lint fmt help dev install-air

# Variables
BINARY_NAME=fambot
BUILD_DIR=bin
MAIN_PATH=cmd/main.go

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting FamBot..."
	@go run $(MAIN_PATH)

# Development mode with auto-reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		echo "Starting FamBot in development mode..."; \
		air; \
	else \
		echo "Air not found. Install with 'make install-air' or run 'make run'"; \
		$(MAKE) run; \
	fi

# Install air for hot reloading
install-air:
	@echo "Installing air for hot reloading..."
	@go install github.com/cosmtrek/air@latest

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install from https://golangci-lint.run/"; \
	fi

# Install linting tools
install-lint:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f fambot.db
	@go clean

# Setup development environment
setup: deps install-air install-lint
	@echo "Development environment setup complete!"
	@echo "Copy .env.example to .env and configure your Slack tokens"

# Create release build
release:
	@echo "Creating release build..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Release builds created in $(BUILD_DIR)/"

# Database operations
db-reset:
	@echo "Resetting database..."
	@rm -f fambot.db
	@echo "Database reset complete. It will be recreated on next run."

# Show database info
db-info:
	@if [ -f fambot.db ]; then \
		echo "Database file: fambot.db"; \
		echo "Size: $$(du -h fambot.db | cut -f1)"; \
		echo "Tables:"; \
		sqlite3 fambot.db ".tables"; \
	else \
		echo "Database file not found. Run the bot first to create it."; \
	fi

# Backup database
db-backup:
	@if [ -f fambot.db ]; then \
		cp fambot.db fambot.db.backup; \
		echo "Database backed up to fambot.db.backup"; \
	else \
		echo "No database file to backup"; \
	fi

# Check environment setup
check-env:
	@echo "Checking environment setup..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Copy .env.example to .env and configure it."; \
	else \
		echo "✅ .env file found"; \
	fi
	@if [ -z "$$SLACK_BOT_TOKEN" ]; then \
		echo "❌ SLACK_BOT_TOKEN not set"; \
	else \
		echo "✅ SLACK_BOT_TOKEN is set"; \
	fi
	@if [ -z "$$SLACK_APP_TOKEN" ]; then \
		echo "❌ SLACK_APP_TOKEN not set"; \
	else \
		echo "✅ SLACK_APP_TOKEN is set"; \
	fi

# Docker operations
docker-build:
	@echo "Building Docker image..."
	@docker build -t fambot:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run --rm --env-file .env fambot:latest

# Help
help:
	@echo "FamBot Development Commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build          Build the application"
	@echo "  run            Run the application"
	@echo "  dev            Run in development mode with hot reload"
	@echo "  release        Create release builds for multiple platforms"
	@echo ""
	@echo "Development:"
	@echo "  deps           Install Go dependencies"
	@echo "  setup          Setup complete development environment"
	@echo "  fmt            Format Go code"
	@echo "  lint           Run linter (requires golangci-lint)"
	@echo "  install-lint   Install golangci-lint"
	@echo "  install-air    Install air for hot reloading"
	@echo ""
	@echo "Testing:"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo ""
	@echo "Database:"
	@echo "  db-reset       Reset the database"
	@echo "  db-info        Show database information"
	@echo "  db-backup      Backup the database"
	@echo ""
	@echo "Environment:"
	@echo "  check-env      Check environment configuration"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo ""
	@echo "Utility:"
	@echo "  clean          Clean build artifacts"
	@echo "  help           Show this help message"
