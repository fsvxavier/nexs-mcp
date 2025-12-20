.PHONY: help build test test-race test-coverage lint fmt vet clean run install-tools build-all docker-build docker-run release

# Variables
BINARY_NAME=nexs-mcp
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
VERSION=0.1.0
DIST_DIR=dist

# Build variables
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION)"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/nexs-mcp

run: build ## Build and run the server
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	@go test -v -timeout 30s ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -v -race -timeout 30s ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -timeout 30s -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'

lint: ## Run linters
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@gofmt -w -s .
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ $(DIST_DIR)/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@go clean

build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/nexs-mcp
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/nexs-mcp
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/nexs-mcp
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/nexs-mcp
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/nexs-mcp
	@echo "All builds completed successfully!"
	@ls -lh $(DIST_DIR)/

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t nexs-mcp:$(VERSION) -t nexs-mcp:latest .
	@echo "Docker image built: nexs-mcp:$(VERSION)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -it \
		-v $(PWD)/data:/app/data \
		nexs-mcp:latest

release: test-coverage lint build-all ## Prepare release artifacts
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release artifacts created in $(DIST_DIR)/"
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release $(VERSION) ready!"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Tools installed successfully"

security: ## Run security scan
	@echo "Running security scan..."
	@govulncheck ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy
	@go mod verify

verify: fmt vet lint test-race ## Run all verification steps
	@echo "All verification steps passed!"

ci: verify security ## Run CI pipeline locally
	@echo "CI pipeline completed successfully!"
