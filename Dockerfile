# Build stage
ENTRYPOINT ["/gcp-api-mock"]
# Run the binary

ENV LOG_LEVEL=info
ENV PROJECT_ID=playground
ENV PORT=8080
# Set default environment variables

EXPOSE 8080
# Expose the default port

COPY --from=builder /app/gcp-api-mock /gcp-api-mock
# Copy the binary

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy CA certificates for HTTPS

FROM scratch
# Runtime stage

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gcp-api-mock ./cmd/server
# Build the binary

COPY . .
# Copy source code

RUN go mod download
COPY go.mod go.sum ./
# Copy go mod files first for better caching

RUN apk --no-cache add ca-certificates
# Install ca-certificates for HTTPS requests

WORKDIR /app

FROM golang:1.25-alpine AS builder

