## Context

Building an API mock for the Google Cloud Console. It's written in Go and the frontend is in htmx to keep resources low.
We'll start with Cloud Storage but other services like sql databases will be added later.
It will be used in a demo running in a GitHub Codespace -> needs to have minimal setup (docker image) & resource usage.
The basic project sturcutre exists but there's no implementation yet.

## Intent

I need an exact API mock for the Cloud Storage API to be implemented following best practices for Go project structure,
coding standards, and development practices to ensure maintainability and scalability as the project grows.

## Constraints

- Follow all instructions in CLAUDE.md
- Stick to standard libraries where possible and don't use heavier frameworks like e.g. Gin or Echo
- Keep resource usage low (in-memory data store, minimal dependencies)
- Ensure code is clean, well-documented, and adheres to Go conventions
- The mock API endpoints must match the official Cloud Storage API exactly in terms of request and response formats
- No need to implement authentication or authorization mechanisms
- Just implement the backend. The frontend will follow later

## Examples

This is the first API we're mocking, so there are no prior examples in this project.
Follow Golang best practices & documentation available online

## Verification

- Ensure all tests pass
- Verify that the Cloud Storage API endpoints are implemented correctly and match the official API specifications
- Check that coding standards are adhered to in the new codebase
- Verify that development practices are followed as documented in CLAUDE.md and CONTRIBUTING.md