#!/bin/bash

# Run all tests with coverage

echo "Running Go tests with coverage..."

# Create coverage directory
mkdir -p coverage

# Run tests with coverage
go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...

# Generate coverage report
if [ -f coverage/coverage.out ]; then
    echo ""
    echo "Coverage Report:"
    go tool cover -func=coverage/coverage.out
    
    # Generate HTML coverage report
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    echo ""
    echo "HTML coverage report generated at: coverage/coverage.html"
fi

# Run go vet
echo ""
echo "Running go vet..."
go vet ./...

# Check for formatting issues
echo ""
echo "Checking formatting..."
gofmt -l .

echo ""
echo "Test run complete!"