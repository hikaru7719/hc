# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GUI-based HTTP client project called "hc" (HTTP Client) that integrates with Claude Code. The project is set up as a Go application with a devcontainer configuration.

## Development Environment

The project uses a Go devcontainer (Go 1.24) with Node.js 24 included. The devcontainer includes:
- Claude Code feature integration
- Biome for code formatting/linting
- Code spell checker VS Code extension

## Project Structure

Currently minimal - this is a new project in early development. The repository includes:
- `.devcontainer/` - Development container configuration
- `.claude/` - Claude Code specific settings
- Basic Go project structure (expected to be added)

## Common Commands

Since this is a new Go project, typical commands will include:

```bash
# Initialize Go module (if not already done)
go mod init github.com/[username]/hc

# Run the application
go run .

# Build the application
go build -o hc

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code
go vet ./...
```

## Architecture Notes

As a GUI HTTP client, the project will likely include:
- HTTP request/response handling logic
- GUI framework integration (to be determined)
- Request history and management
- Response parsing and display
- Authentication handling
- Request collections/environments

## Development Guidelines

- Follow standard Go conventions and idioms
- Use Go modules for dependency management
- Keep HTTP client logic separate from GUI code
- Consider using interfaces for testability
- Structure code with clear separation between business logic and UI