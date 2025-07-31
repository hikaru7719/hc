# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HC (HTTP Client) is a GUI-based HTTP client built with Go backend and Next.js frontend. It runs as a local server (`hc serve`) providing a browser-based interface for making HTTP requests. The frontend is embedded in the Go binary for single-file distribution.

## Common Commands

```bash
# Build everything (frontend + backend)
make build

# Run the application
make run
# Or directly: ./build/hc serve

# Development mode with hot reload
make dev              # Backend with air hot reload
make dev-frontend     # Frontend dev server on port 3000

# Run tests
make test             # Runs both Go and frontend tests
go test ./...         # Go tests only
go test ./internal/storage -v  # Run specific package tests

# Code quality
make lint             # Run linters (go vet, go fmt, npm lint)
go fmt ./...          # Format Go code
go vet ./...          # Check Go code
go build -o hc main.go # Build go binary

# Clean build artifacts
make clean

# Install dependencies
make deps

# Build for multiple platforms
make build-all
```

## Architecture

### High-Level Structure
- **CLI Entry**: `main.go` uses Cobra for command handling, primarily `hc serve`
- **Server**: `internal/server/server.go` handles HTTP endpoints and serves embedded frontend
- **Proxy**: `internal/proxy/proxy.go` executes HTTP requests on behalf of the frontend
- **Storage**: `internal/storage/sqlite.go` manages SQLite database for request/folder persistence
- **Frontend**: Next.js SPA built to static files, embedded via `embed.go`

### API Endpoints
- `POST /api/request` - Execute HTTP request via proxy
- `GET/POST /api/requests` - List/create saved requests
- `GET/PUT/DELETE /api/requests/:id` - Manage specific request
- `GET/POST /api/folders` - List/create folders
- `GET/PUT/DELETE /api/folders/:id` - Manage specific folder

### Database Schema
- **folders**: Hierarchical folder structure (id, name, parent_id)
- **requests**: HTTP requests (id, name, folder_id, method, url, headers, body)

### Key Implementation Details
- Frontend files are embedded using `go:embed` in `embed.go`
- SQLite database is initialized in user's home directory (`~/.hc/hc.db`)
- Static file serving falls back to index.html for client-side routing

### Code Policy

#### Common

- Do not use comment if you can understand code to read program like simple logic.
  - Use comment with complex code.
- Do not define unnecessary variable.
  - Use direct assignment with literal
- Remove blank line in a function. You should not use blank line in a function.
- Keep considering best code.
- Keep function small.

#### Frontend

- Avoid using useEffect function.
- Avoid using a lot of useState function.
  - make useReducer function to aggregate useState.
- Avoid using useCallback.
- Avoid using try-catch code and handle error globally.
- Consider to split large React components.
- Define constant variable.
- Avoid using null or undefined or optional param.
- Use functional idiom like map, filter.
- Use async-await pattern, avoid using Promise chain like then, catch.

#### Backend

- Make test code and use table driven test.
- Do not use unnecessary wrap error by using fmt.Errorf
- Avoid using init function. init function order is difficult to understand.
