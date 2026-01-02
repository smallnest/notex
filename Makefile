.PHONY: build run test clean fmt vet lint help

# Binary name
BINARY_NAME=open-notebook

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOFMT=gofmt
GOVET=go vet

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) backend/*.go
	@echo "Build complete: $(BINARY_NAME)"

# Run the application in server mode
run:
	@echo "Starting $(BINARY_NAME) server..."
	$(GORUN) backend/*.go -server

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool coverage -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Check if code is formatted
fmt-check:
	@echo "Checking code formatting..."
	@test -z $$($(GOFMT) -l .) || (echo "Code is not formatted. Run 'make fmt'." && exit 1)

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run all checks
check: fmt-check vet lint
	@echo "All checks passed!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf data/
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Create data directory
init:
	@echo "Initializing data directories..."
	mkdir -p data/uploads
	@echo "Initialized"

# Build and run (development)
dev: init run

# Run with OpenAI
run-openai:
	@echo "Starting with OpenAI..."
	@if [ -z "$$OPENAI_API_KEY" ]; then \
		echo "Error: OPENAI_API_KEY not set"; \
		exit 1; \
	fi
	$(GORUN) backend/*.go -server

# Run with Ollama
run-ollama:
	@echo "Starting with Ollama..."
	$(GORUN) backend/*.go -server

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the server"
	@echo "  dev            - Initialize and run (development)"
	@echo "  run-openai     - Run with OpenAI (requires OPENAI_API_KEY)"
	@echo "  run-ollama     - Run with Ollama"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  fmt            - Format code"
	@echo "  fmt-check      - Check if code is formatted"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linter"
	@echo "  check          - Run all checks (fmt, vet, lint)"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  init           - Create data directories"
	@echo "  help           - Show this help message"
