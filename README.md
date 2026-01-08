# GCP API Mock

A lightweight mock for GCP APIs - test your Infrastructure as Code configs without a real GCP account.

## Overview

GCP API Mock provides a local, in-memory implementation of Google Cloud Platform APIs. It's designed to be:

- **Lightweight** - Minimal dependencies, uses only Go standard library
- **Fast** - In-memory data store, instant startup
- **Container-ready** - Optimized Docker image for GitHub Codespaces

## Quick Start

### Running Locally

```bash
# Run the server
make run

# Or build and run
make build
./bin/server
```

### Running with Docker

```bash
# Build the image
make docker-build

# Run the container
make docker-run
```

The server starts on `http://localhost:8080` by default.

## Configuration

Configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `GCP_MOCK_HOST` | `0.0.0.0` | Server host address |
| `GCP_MOCK_PORT` | `8080` | Server port |
| `GCP_MOCK_ENV` | `development` | Environment (development/production) |

## API Endpoints

### Health Checks

- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe

### UI

- `GET /` - Web dashboard (HTMX)

### GCP APIs (Coming Soon)

- Cloud Storage API

## Project Structure

```
gcp-api-mock/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── requestid/       # Request ID utilities
│   ├── server/          # Server setup and routing
│   └── store/           # In-memory data store
├── web/
│   ├── static/          # Static assets (CSS, JS)
│   └── templates/       # HTML templates
├── .github/
│   └── workflows/       # CI/CD pipelines
├── Dockerfile           # Container image
├── Makefile            # Development commands
└── go.mod              # Go module
```

## Development

### Prerequisites

- Go 1.23 or later
- Docker (optional, for containerized deployment)

### Available Commands

```bash
make help           # Show all available commands
make run            # Run the server locally
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make lint           # Run linter checks
make fmt            # Format code
make build          # Build binary
make clean          # Clean build artifacts
```

### Testing

Tests follow the standard Go testing conventions:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/config/...
```

### Code Style

- Follow standard Go conventions and formatting
- Run `make fmt` before committing
- All code must pass `make lint`

## Contributing

1. Ensure all tests pass: `make test`
2. Ensure code is formatted: `make fmt`
3. Ensure linting passes: `make lint`
4. Write tests for new features

## License

MIT License - see [LICENSE](LICENSE) for details.

