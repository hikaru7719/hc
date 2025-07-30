# HC - HTTP Client Makefile

# Variables
BINARY_NAME=hc
FRONTEND_DIR=frontend
GO_CMD=go
NPM_CMD=npm
NEXT_CMD=npx next

# Go build variables
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
BUILD_DIR=build
LDFLAGS=-ldflags="-s -w"

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: all build build-backend build-frontend clean run dev test lint help

# Default target
all: build

# Build both frontend and backend
build: build-frontend build-backend
	@echo "$(GREEN)✓ Build completed successfully$(NC)"

# Build backend
build-backend:
	@echo "$(YELLOW)Building backend...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO_CMD) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Backend built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build frontend
build-frontend:
	@echo "$(YELLOW)Building frontend...$(NC)"
	@cd $(FRONTEND_DIR) && $(NPM_CMD) install
	@cd $(FRONTEND_DIR) && $(NPM_CMD) run lint
	@cd $(FRONTEND_DIR) && $(NPM_CMD) run format
	@cd $(FRONTEND_DIR) && $(NPM_CMD) run build
	@echo "$(GREEN)✓ Frontend built: $(FRONTEND_DIR)/out$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@cd $(FRONTEND_DIR) && rm -rf out .next
	@echo "$(GREEN)✓ Clean completed$(NC)"

# Run the application
run: build
	@echo "$(YELLOW)Starting HC...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Development mode - run backend with hot reload
dev:
	@echo "$(YELLOW)Starting development mode...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air for hot reload..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

# Run frontend development server
dev-frontend:
	@echo "$(YELLOW)Starting frontend development server...$(NC)"
	@cd $(FRONTEND_DIR) && $(NPM_CMD) run dev

# Run tests
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO_CMD) test -v ./...
	@cd $(FRONTEND_DIR) && $(NPM_CMD) test

# Run tests with coverage
test-coverage:
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	$(GO_CMD) test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	@echo "$(GREEN)✓ Coverage report: coverage/coverage.out$(NC)"
	@echo ""
	@echo "Coverage summary:"
	@$(GO_CMD) tool cover -func=coverage/coverage.out | tail -1
	@$(GO_CMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "$(GREEN)✓ HTML report: coverage/coverage.html$(NC)"

# Lint code
lint:
	@echo "$(YELLOW)Running linters...$(NC)"
	$(GO_CMD) vet ./...
	$(GO_CMD) fmt ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	fi
	@cd $(FRONTEND_DIR) && $(NPM_CMD) run lint

# Install dependencies
deps:
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(GO_CMD) mod download
	@cd $(FRONTEND_DIR) && $(NPM_CMD) install
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

# Cross-platform builds
build-all: build-frontend
	@echo "$(YELLOW)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO_CMD) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "$(GREEN)✓ Multi-platform build completed$(NC)"

# Help
help:
	@echo "HC - HTTP Client Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            Build both frontend and backend (default)"
	@echo "  build          Build both frontend and backend"
	@echo "  build-backend  Build only the backend"
	@echo "  build-frontend Build only the frontend"
	@echo "  clean          Remove build artifacts"
	@echo "  run            Build and run the application"
	@echo "  dev            Run backend in development mode with hot reload"
	@echo "  dev-frontend   Run frontend development server"
	@echo "  test           Run all tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  lint           Run linters"
	@echo "  deps           Install all dependencies"
	@echo "  build-all      Build for multiple platforms"
	@echo "  help           Show this help message"
