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
	@go test -v -race -timeout 240s -coverprofile=$(COVERAGE_FILE) ./...
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

build-all: clean ## Build for all platforms (ONNX disabled for cross-compilation)
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/nexs-mcp
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/nexs-mcp
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/nexs-mcp
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/nexs-mcp
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/nexs-mcp
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

docker-publish: ## Publish Docker image to Docker Hub (requires .env with DOCKER_USER and DOCKER_TOKEN)
	@echo "Publishing Docker image to Docker Hub..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@set -a && . ./.env && set +a && \
	if [ -z "$$DOCKER_USER" ] || [ -z "$$DOCKER_TOKEN" ]; then \
		echo "Error: DOCKER_USER and DOCKER_TOKEN must be set in .env"; \
		exit 1; \
	fi; \
	echo "Logging in to Docker Hub as $$DOCKER_USER..."; \
	echo "$$DOCKER_TOKEN" | docker login -u "$$DOCKER_USER" --password-stdin; \
	if [ $$? -ne 0 ]; then \
		echo "Error: Docker login failed"; \
		exit 1; \
	fi; \
	echo "Building image $$DOCKER_IMAGE..."; \
	docker build -t $$DOCKER_IMAGE -t $${DOCKER_IMAGE%:*}:v$(VERSION) .; \
	if [ $$? -ne 0 ]; then \
		echo "Error: Docker build failed"; \
		exit 1; \
	fi; \
	echo "Pushing $$DOCKER_IMAGE..."; \
	docker push $$DOCKER_IMAGE; \
	if [ $$? -ne 0 ]; then \
		echo "Error: Docker push failed"; \
		exit 1; \
	fi; \
	echo "Pushing $${DOCKER_IMAGE%:*}:v$(VERSION)..."; \
	docker push $${DOCKER_IMAGE%:*}:v$(VERSION); \
	echo "Docker image published successfully!"; \
	echo "Images:"; \
	echo "  - $$DOCKER_IMAGE"; \
	echo "  - $${DOCKER_IMAGE%:*}:v$(VERSION)"

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

npm-publish: ## Publish package to NPM registry
	@echo "Publishing to NPM registry..."
	@npm publish --access public
	@echo "Package published successfully!"

npm-publish-github: ## Publish package to GitHub NPM registry
	@echo "Publishing to GitHub NPM registry..."
	@npm publish --registry=https://npm.pkg.github.com
	@echo "Package published successfully!"

github-publish: ## Create GitHub tag and release (usage: make github-publish VERSION=1.0.5 MESSAGE="Release notes")
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make github-publish VERSION=1.0.5"; \
		exit 1; \
	fi
	@echo "Checking if tag v$(VERSION) already exists..."
	@if git rev-parse v$(VERSION) >/dev/null 2>&1; then \
		echo "Warning: Tag v$(VERSION) already exists locally."; \
		read -p "Do you want to delete and recreate it? (y/N): " answer; \
		if [ "$$answer" = "y" ] || [ "$$answer" = "Y" ]; then \
			git tag -d v$(VERSION); \
			git push origin :refs/tags/v$(VERSION) 2>/dev/null || true; \
			echo "Deleted local and remote tag v$(VERSION)"; \
		else \
			echo "Aborted."; \
			exit 1; \
		fi; \
	fi
	@if gh release view v$(VERSION) >/dev/null 2>&1; then \
		echo "Warning: Release v$(VERSION) already exists on GitHub."; \
		read -p "Do you want to delete and recreate it? (y/N): " answer; \
		if [ "$$answer" = "y" ] || [ "$$answer" = "Y" ]; then \
			gh release delete v$(VERSION) -y; \
			echo "Deleted release v$(VERSION)"; \
		else \
			echo "Aborted."; \
			exit 1; \
		fi; \
	fi
	@echo "Creating tag and release v$(VERSION)..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@if [ -z "$(MESSAGE)" ]; then \
		gh release create v$(VERSION) --title "Release v$(VERSION)" --generate-notes; \
	else \
		gh release create v$(VERSION) --title "Release v$(VERSION)" --notes "$(MESSAGE)"; \
	fi
	@echo "Tag and release v$(VERSION) created successfully!"


