# GCP API Mock - Makefile
# Common commands for development and CI/CD

.PHONY: all build run test test-coverage lint clean docker-build docker-run help

# Default target
all: lint test build

# Build the server binary
build:
	@echo "Building server..."
	@go build -o bin/server ./cmd/server

# Run the server locally
run:
	@echo "Starting server..."
	@go run ./cmd/server

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t gcp-api-mock:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 gcp-api-mock:latest

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify

# Show help
help:
	@echo "GCP API Mock - Available commands:"
	@echo ""
	@echo "  make build          - Build the server binary"
	@echo "  make run            - Run the server locally"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter checks"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make deps           - Download dependencies"
	@echo "  make verify         - Verify dependencies"
	@echo "  make help           - Show this help message"

