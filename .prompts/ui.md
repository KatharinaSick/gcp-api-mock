## Context

Building an API mock for the Google Cloud Console. It's written in Go and the frontend is in htmx to keep resources low.
We'll start with Cloud Storage but other services like sql databases will be added later.
It will be used in a demo running in a GitHub Codespace -> needs to have minimal setup (docker image) & resource usage.
The backend works pretty well by now but there is basically no frontend yet.

## Intent

An UI that shows the resources that are managed via the Cloud SQL API mock.

## Constraints

- Keep the ui lightweight with htmx and minimal dependencies
- Follow all instructions in CLAUDE.md
- Ensure code is clean, well-documented, and adheres to Go conventions
- The UI should allow basic operations like listing, creating, and deleting gcs and cloud sql instances
- Just implement the frontend. The backend is already implemented
- Integrate the UI with the existing backend API
- The UI doesn't have to be responsive, it's fine if it looks good on a desktop only
- It should have a retro look like a terminal ui and show that it's a mock implementation
- Add a "request log" showing all requests that were sent to the backend on the right
