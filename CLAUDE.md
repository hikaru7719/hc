# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HC (HTTP Client) is a GUI-based HTTP client built with Go backend and Next.js frontend. It runs as a local server (`hc serve`) providing a browser-based interface for making HTTP requests. The frontend is embedded in the Go binary for single-file distribution.

## Technology Stack

### Backend
- Go 1.24+ - Core backend language
- Cobra - CLI framework for command handling
- Chi - HTTP router for REST API
- SQLite - Embedded database for persistence
- go:embed - Static file embedding for single binary distribution

### Frontend
- Next.js 15 - React framework with App Router
- React 19 - UI library
- TypeScript - Type-safe JavaScript
- Tailwind CSS - Utility-first CSS framework
- Biome - Fast formatter and linter
- Lucide React - Icon library

### Development Tools
- Make - Build automation
- Air - Hot reload for Go development
- npm/pnpm - Node.js package management

## Common Commands

```bash
# Build everything (frontend + backend)
make build

# Run the application
make run
# Or directly: ./build/hc serve

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
```sql
-- Folders table for organizing requests
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Requests table for storing HTTP requests
CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    folder_id INTEGER,
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    headers TEXT,  -- JSON string
    body TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);
```

### Key Implementation Details
- Frontend files are embedded using `go:embed` in `embed.go`
- SQLite database is initialized in user's home directory (`~/.hc/hc.db`)
- Static file serving falls back to index.html for client-side routing

## Directory Structure

### Project Root
```
.
├── cmd/                    # CLI command definitions
├── frontend/              # Next.js frontend application
├── internal/              # Go internal packages
├── docs/                  # Documentation
├── .claude/               # Claude AI configuration and commands
├── main.go               # Application entry point
├── embed.go              # Frontend static files embedding
├── Makefile              # Build and development commands
├── go.mod, go.sum        # Go dependencies
└── README.md             # Project documentation
```

### Backend Structure (Go)
```
cmd/                      # CLI command definitions
internal/
├── logger/               # Logging utilities
├── middleware/           # HTTP middleware
├── models/              # Data models
├── proxy/               # HTTP proxy functionality
├── server/              # HTTP server and handlers
└── storage/             # Database layer
```

### Frontend Structure (Next.js)
```
frontend/
├── app/                  # Next.js App Router
├── components/          # React components
├── api/                 # API client logic
├── hooks/               # Custom React hooks
├── utils/               # Utility functions
├── constants/           # Application constants
├── types/               # TypeScript type definitions
├── biome.json          # Biome formatter/linter config
├── tailwind.config.ts  # Tailwind CSS configuration
├── tsconfig.json       # TypeScript configuration
└── package.json        # Node.js dependencies
```

### Build Artifacts (git-ignored)
```
build/                   # Go binary output
frontend/out/           # Next.js static export
frontend/.next/         # Next.js build cache
node_modules/           # Node.js dependencies
```

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
