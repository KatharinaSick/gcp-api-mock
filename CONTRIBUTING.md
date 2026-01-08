# CONTRIBUTING.md

Guidelines for contributing to GCP API Mock.

## Development Workflow

### 1. Setup

```bash
# Clone the repository
git clone https://github.com/ksick/gcp-api-mock.git
cd gcp-api-mock

# Download dependencies
make deps

# Verify everything works
make test
```

### 2. Making Changes

1. Create a feature branch from `main`
2. Write tests for your changes (TDD encouraged)
3. Implement your changes
4. Ensure all tests pass: `make test`
5. Ensure code is formatted: `make fmt`
6. Ensure linting passes: `make lint`

### 3. Commit Guidelines

- Use clear, descriptive commit messages
- Reference issue numbers when applicable
- Keep commits focused and atomic

### 4. Pull Request

- Ensure CI passes
- Request review from maintainers
- Address feedback promptly

## Code Standards

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (enforced by CI)
- Use `go vet` for static analysis

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Files**: lowercase, snake_case (e.g., `request_id.go`)
- **Tests**: same file with `_test.go` suffix
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase
- **Constants**: PascalCase for exported, camelCase for unexported

### CSS Naming (BEM-style with prefix)

All CSS classes must be prefixed with `gcp-mock-` to prevent conflicts:

```css
/* Good */
.gcp-mock-header { }
.gcp-mock-header-title { }
.gcp-mock-button-primary { }

/* Bad - no prefix */
.header { }
.title { }
.button-primary { }
```

### Testing

- Write tests before implementing features (TDD)
- Use table-driven tests where appropriate
- Test both success and error cases
- Aim for high coverage, but prioritize meaningful tests

Example table-driven test:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "abc", "ABC", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## Project Architecture

### Package Organization

```
internal/           # Private application code
├── config/         # Configuration loading
├── handler/        # HTTP handlers (one file per domain)
├── middleware/     # HTTP middleware
├── requestid/      # Request ID utilities
├── server/         # Server setup and routing
└── store/          # In-memory data store
```

### Adding a New Handler

1. Create handler file: `internal/handler/{domain}.go`
2. Create test file: `internal/handler/{domain}_test.go`
3. Register routes in `internal/server/server.go`

### Adding a New Middleware

1. Create middleware file: `internal/middleware/{name}.go`
2. Create test file: `internal/middleware/{name}_test.go`
3. Add to middleware chain in `internal/server/server.go`

## CI/CD

GitHub Actions runs on every push and PR:

1. **Test** - Runs all tests with race detection
2. **Lint** - Checks formatting and runs go vet
3. **Build** - Verifies the application builds
4. **Docker** - Builds the Docker image

All checks must pass before merging.

