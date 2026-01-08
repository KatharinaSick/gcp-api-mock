## Context

Building an API mock for the Google Cloud Console. It's written in Go and the frontend is in htmx to keep resources low.
We'll start with Cloud Storage but other services like sql databases will be added later.
It will be used in a demo running in a GitHub Codespace -> needs to have minimal setup (docker image) & resource usage.
There's no implementation yet.

## Intent

I need a solid foundation for the project structure, coding standards, and development practices to ensure
maintainability and scalability as the project grows.

## Constraints

- Follow best practices for Go project strcuture
- Stick to standard libraries where possible and don't use heavier frameworks like e.g. Gin or Echo
- Keep resource usage low (in-memory data store, minimal dependencies)
- Ensure code is clean, well-documented, and adheres to Go conventions

## Examples

Follow Goolang best practices & documentation available online

## Verification

- Make sure one can start the service
- Ensure the project structure is clear and follows Go conventions
- Check that coding standards are adhered to in the initial codebase
- Verify that development practices are documented and can be followed by new contributors