# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gcp-api-mock ./cmd/server

# Runtime stage
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/gcp-api-mock /gcp-api-mock

# Expose the default port
EXPOSE 8080

# Set default environment variables
ENV PORT=8080
ENV PROJECT_ID=playground
ENV LOG_LEVEL=info

# Run the binary
ENTRYPOINT ["/gcp-api-mock"]

