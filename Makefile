.PHONY: help build test test-unit test-integration run clean deps fmt lint vet

# Default target
.DEFAULT_GOAL := help

# Build variables
BINARY_NAME=config-engine
MAIN_PATH=./main.go
BUILD_DIR=./bin

help: ## Display this help message
	@echo "Configuration Management Service - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download and install dependencies
	@echo "==> Installing dependencies..."
	go mod download
	go mod tidy
	@echo "==> Dependencies installed successfully"

build: deps ## Build the application
	@echo "==> Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run the application
	@echo "==> Starting $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run the application without building binary
	@echo "==> Running in development mode..."
	go run $(MAIN_PATH)

test: test-unit test-integration ## Run all tests (unit + integration)

test-unit: deps ## Run unit tests only
	@echo "==> Running unit tests..."
	go test -v -race -coverprofile=coverage-unit.out ./internal/...
	@echo "==> Unit tests completed"

test-integration: deps ## Run integration/functional tests only
	@echo "==> Running integration tests..."
	go test -v -race -coverprofile=coverage-integration.out ./tests/...
	@echo "==> Integration tests completed"

test-coverage: test ## Run tests and display coverage report
	@echo "==> Generating coverage report..."
	go tool cover -html=coverage-unit.out -o coverage-unit.html
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "==> Coverage reports generated: coverage-unit.html, coverage-integration.html"

fmt: ## Format the code
	@echo "==> Formatting code..."
	go fmt ./...
	@echo "==> Code formatted"

vet: ## Run go vet
	@echo "==> Running go vet..."
	go vet ./...
	@echo "==> Vet completed"

lint: ## Run golint (requires golint to be installed)
	@echo "==> Running golint..."
	@command -v golint >/dev/null 2>&1 || { echo "golint not installed. Install with: go install golang.org/x/lint/golint@latest"; exit 1; }
	golint ./...
	@echo "==> Lint completed"

clean: ## Remove build artifacts and coverage files
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage-unit.out coverage-integration.out
	rm -f coverage-unit.html coverage-integration.html
	@echo "==> Cleanup complete"

verify: fmt vet test ## Run format, vet, and all tests

install: build ## Install the binary to $GOPATH/bin
	@echo "==> Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)
	@echo "==> Installation complete"

docker-build: ## Build Docker image (if Dockerfile exists)
	@echo "==> Building Docker image..."
	docker build -t $(BINARY_NAME):latest .
	@echo "==> Docker image built"

docker-run: docker-build ## Run Docker container
	@echo "==> Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

benchmark: ## Run benchmarks
	@echo "==> Running benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "==> Benchmarks completed"