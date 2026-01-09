# GCP API Mock

A lightweight mock for GCP APIs - test your Infrastructure as Code configs without a real GCP account.

## Quick Start

```bash
# Run the server
make run

# Or with Docker
docker run -p 8080:8080 ghcr.io/katharinasick/gcp-api-mock
```

## What's Supported

- **Cloud Storage API mock** - Buckets and objects (list, create, get, update, delete)
- **Web Dashboard** - See all your mock resources in real-time

## Configuration

| Variable     | Default      | Description         |
|--------------|--------------|---------------------|
| `PORT`       | `8080`       | Server port         |
| `PROJECT_ID` | `playground` | Default GCP project |

## License

MIT