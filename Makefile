.PHONY: help build test test-race test-coverage lint fmt vet clean run install-tools build-all docker-build docker-run release

# Variables
BINARY_NAME=nexs-mcp
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
VERSION=1.3.0
DIST_DIR=dist
ONNX?=0

# Build variables
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION)"

# Build configuration based on ONNX flag
ifeq ($(ONNX),1)
	BUILD_CGO_ENABLED=1
	BUILD_TAGS=
	BUILD_CFLAGS=-I/usr/local/include
	BUILD_LDFLAGS=-L/usr/local/lib -lonnxruntime
	BUILD_MODE=with ONNX support
else
	BUILD_CGO_ENABLED=0
	BUILD_TAGS=-tags noonnx
	BUILD_CFLAGS=
	BUILD_LDFLAGS=
	BUILD_MODE=portable (without ONNX)
endif

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary (default: portable without ONNX, use ONNX=1 for ONNX support)
	@echo "Building $(BINARY_NAME) $(BUILD_MODE)..."
	@CGO_ENABLED=$(BUILD_CGO_ENABLED) CGO_CFLAGS="$(BUILD_CFLAGS)" CGO_LDFLAGS="$(BUILD_LDFLAGS)" go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME) ./cmd/nexs-mcp

build-noonnx: ## Build the binary without ONNX support (portable, no CGO)
	@echo "Building $(BINARY_NAME) without ONNX (using fallback chain)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -tags noonnx -o bin/$(BINARY_NAME) ./cmd/nexs-mcp

build-onnx: ## Build the binary with ONNX support (requires ONNX Runtime installed)
	@echo "Building $(BINARY_NAME) with ONNX support..."
	@echo "Note: Requires ONNX Runtime installed (see 'make install-onnx' or docs/development/ONNX_SETUP.md and docs/development/ONNX_ENVIRONMENT_SETUP.md)"
	@CGO_ENABLED=1 \
		CGO_CFLAGS="-I/usr/local/include" \
		CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime" \
		go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/nexs-mcp
	@echo "✓ Build with ONNX complete"

run: build ## Build and run the server
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	@go test -v -timeout 10m ./...

# Run the integration server test script with tracing and an overall timeout
test-mcp-trace: ## Run integration server test with bash -x and a timeout (default TIMEOUT=60s)
	@TIMEOUT=${TIMEOUT:-60s}; \
	if ! command -v timeout >/dev/null 2>&1; then \
		echo "Error: 'timeout' command not found. Install coreutils (Linux) or gtimeout via coreutils on macOS (brew install coreutils)."; exit 1; \
	fi; \
	@echo "Running test_mcp_server.sh with timeout=$$TIMEOUT (bash -x)..."; \
	timeout $$TIMEOUT bash -x test_mcp_server.sh

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -v -race -timeout 10m ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -timeout 10m -coverprofile=$(COVERAGE_FILE) ./...
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
	@find bin/ -type f ! -name 'nexs-mcp.js' -delete 2>/dev/null || true
	@find $(DIST_DIR)/ -type f ! -name 'nexs-mcp.js' -delete 2>/dev/null || true
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@go clean

build-all: clean ## Build for all platforms (default: portable, use ONNX=1 for ONNX support)
	@echo "⚠️  WARNING: Cross-compilation currently has issues with HNSW library dependencies"
	@echo "⚠️  Use 'make build' for native builds or build on target platform"
	@echo "Building for all platforms $(BUILD_MODE)..."
	@mkdir -p bin
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/nexs-mcp
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/nexs-mcp
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/nexs-mcp
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/nexs-mcp
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/nexs-mcp
	@echo "Building for Windows (arm64)..."
	@GOOS=windows GOARCH=arm64 CGO_ENABLED=$(BUILD_CGO_ENABLED) \
		CGO_CFLAGS="$(BUILD_CFLAGS)" \
		CGO_LDFLAGS="$(BUILD_LDFLAGS)" \
		go build $(LDFLAGS) $(BUILD_TAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe ./cmd/nexs-mcp
	@echo "All builds completed successfully!"
	@ls -lh bin/

dist: build ## Copy binary to dist folder
	@echo "Creating dist folder..."
	@mkdir -p $(DIST_DIR)
	@echo "Copying binary to dist..."
	@cp bin/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)
	@echo "Distribution ready in $(DIST_DIR)/"

dist-all: build-all ## Copy all platform binaries to dist folder
	@echo "Creating dist folder..."
	@mkdir -p $(DIST_DIR)
	@echo "Copying binaries to dist..."
	@cp bin/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/$(BINARY_NAME)-linux-amd64
	@cp bin/$(BINARY_NAME)-linux-arm64 $(DIST_DIR)/$(BINARY_NAME)-linux-arm64
	@cp bin/$(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64
	@cp bin/$(BINARY_NAME)-darwin-arm64 $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64
	@cp bin/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@cp bin/$(BINARY_NAME)-windows-arm64.exe $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe
	@echo "All binaries copied to $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

docker-build: ## Build Docker image with ONNX support and models
	@echo "Building Docker image with ONNX Runtime and models..."
	@echo "Note: This will copy pre-downloaded ONNX models (~550MB) from models/ directory"
	@docker build -t nexs-mcp:$(VERSION) -t nexs-mcp:latest .
	@echo "Docker image built: nexs-mcp:$(VERSION)"
	@echo "Image includes:"
	@echo "  - ONNX Runtime v1.23.2"
	@echo "  - MS MARCO MiniLM-L-6-v2 model"
	@echo "  - Paraphrase-Multilingual-MiniLM-L12-v2 model"

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
	echo "Building image $$DOCKER_IMAGE with ONNX support..."; \
	echo "Note: This will copy pre-downloaded ONNX models (~550MB) from models/ directory"; \
	docker build -t $$DOCKER_IMAGE -t $${DOCKER_IMAGE%:*}:v$(VERSION) .; \
	if [ $$? -ne 0 ]; then \
		echo "Error: Docker build failed"; \
		exit 1; \
	fi; \
	echo "Image includes ONNX Runtime v1.23.2 and pre-loaded models"; \
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

release: test-coverage lint dist-all ## Prepare release artifacts
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && \
		zip $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe
	@echo "Release artifacts created in $(DIST_DIR)/"
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release $(VERSION) ready!"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Tools installed successfully"

install-onnx: ## Install ONNX Runtime (cross-platform, requires sudo on Linux/macOS)
	@echo "Installing ONNX Runtime v1.23.2..."
	@UNAME=$$(uname -s); \
	if [ "$$UNAME" = "Linux" ]; then \
		echo "Detected Linux platform"; \
		echo "Note: This requires sudo privileges"; \
		cd /tmp && \
		wget -q https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz && \
		tar -xzf onnxruntime-linux-x64-1.23.2.tgz && \
		sudo cp -r onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/ && \
		sudo cp -r onnxruntime-linux-x64-1.23.2/include/* /usr/local/include/ && \
		sudo ldconfig && \
		rm -rf onnxruntime-linux-x64-1.23.2* && \
		echo "✓ ONNX Runtime installed successfully to /usr/local"; \
	elif [ "$$UNAME" = "Darwin" ]; then \
		echo "Detected macOS platform"; \
		echo "Note: This requires sudo privileges"; \
		cd /tmp && \
		curl -LO https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-osx-universal2-1.23.2.tgz && \
		tar -xzf onnxruntime-osx-universal2-1.23.2.tgz && \
		sudo cp -r onnxruntime-osx-universal2-1.23.2/lib/* /usr/local/lib/ && \
		sudo cp -r onnxruntime-osx-universal2-1.23.2/include/* /usr/local/include/ && \
		sudo update_dyld_shared_cache 2>/dev/null || true && \
		rm -rf onnxruntime-osx-universal2-1.23.2* && \
		echo "✓ ONNX Runtime installed successfully to /usr/local"; \
	elif [ "$$UNAME" = "MINGW64_NT" ] || [ "$$UNAME" = "MSYS_NT" ] || echo "$$UNAME" | grep -q "^MINGW"; then \
		echo "Detected Windows platform"; \
		echo "Please install ONNX Runtime manually on Windows:"; \
		echo "1. Download: https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-win-x64-1.23.2.zip"; \
		echo "2. Extract to C:\\onnxruntime"; \
		echo "3. Add C:\\onnxruntime\\lib to your PATH"; \
		echo "See docs/development/ONNX_SETUP.md for detailed instructions"; \
		exit 1; \
	else \
		echo "Error: Unsupported platform: $$UNAME"; \
		echo "Supported platforms: Linux, macOS (Darwin)"; \
		echo "For Windows, see docs/development/ONNX_SETUP.md"; \
		exit 1; \
	fi
	@echo "You can now build with ONNX support: make build-onnx"

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
	@PACKAGE_VERSION=$$(node -p "require('./package.json').version"); \
	PACKAGE_NAME=$$(node -p "require('./package.json').name"); \
	echo "Checking if version $$PACKAGE_VERSION already exists on NPM..."; \
	EXISTING_VERSION=$$(npm view $$PACKAGE_NAME versions --json 2>/dev/null | grep -o "\"$$PACKAGE_VERSION\"" | head -1 | tr -d '"' || echo ""); \
	if [ -n "$$EXISTING_VERSION" ]; then \
		echo ""; \
		echo "⚠️  Warning: Version $$PACKAGE_VERSION already exists on NPM registry."; \
		echo ""; \
		echo "Options:"; \
		echo "  1) Delete the existing version and publish again (npm unpublish)"; \
		echo "  2) Cancel the publication"; \
		echo ""; \
		read -p "Choose an option (1/2): " option; \
		if [ "$$option" = "1" ]; then \
			echo "Deleting version $$PACKAGE_VERSION from NPM..."; \
			npm unpublish $$PACKAGE_NAME@$$PACKAGE_VERSION; \
			if [ $$? -ne 0 ]; then \
				echo "Error: Failed to unpublish version $$PACKAGE_VERSION"; \
				exit 1; \
			fi; \
			echo "Version $$PACKAGE_VERSION deleted successfully."; \
			echo "Publishing new version..."; \
			npm publish --access public; \
			if [ $$? -ne 0 ]; then \
				echo "Error: Failed to publish package"; \
				echo "If authentication is required, it will open in your browser"; \
				exit 1; \
			fi; \
			echo "Package published successfully!"; \
		elif [ "$$option" = "2" ]; then \
			echo "Publication cancelled."; \
			exit 0; \
		else \
			echo "Invalid option. Publication cancelled."; \
			exit 1; \
		fi; \
	else \
		echo "Version $$PACKAGE_VERSION does not exist on NPM. Publishing..."; \
		echo "If authentication is required, it will open in your browser..."; \
		npm publish --access public; \
		if [ $$? -ne 0 ]; then \
			echo "Error: Failed to publish package"; \
			echo "Try: npm login (it will open browser authentication)"; \
			exit 1; \
		fi; \
		echo "Package published successfully!"; \
	fi

npm-publish-github: ## Publish package to GitHub NPM registry
	@echo "Publishing to GitHub NPM registry..."
	@PACKAGE_VERSION=$$(node -p "require('./package.json').version"); \
	PACKAGE_NAME=$$(node -p "require('./package.json').name"); \
	echo "Checking if version $$PACKAGE_VERSION already exists on GitHub NPM..."; \
	EXISTING_VERSION=$$(npm view $$PACKAGE_NAME@$$PACKAGE_VERSION version --registry=https://npm.pkg.github.com 2>/dev/null || echo ""); \
	if [ -n "$$EXISTING_VERSION" ]; then \
		echo ""; \
		echo "⚠️  Warning: Version $$PACKAGE_VERSION already exists on GitHub NPM registry."; \
		echo ""; \
		echo "Options:"; \
		echo "  1) Delete the existing version and publish again (npm unpublish)"; \
		echo "  2) Cancel the publication"; \
		echo ""; \
		read -p "Choose an option (1/2): " option; \
		if [ "$$option" = "1" ]; then \
			echo "Deleting version $$PACKAGE_VERSION from GitHub NPM..."; \
			npm unpublish $$PACKAGE_NAME@$$PACKAGE_VERSION --registry=https://npm.pkg.github.com; \
			if [ $$? -ne 0 ]; then \
				echo "Error: Failed to unpublish version $$PACKAGE_VERSION"; \
				exit 1; \
			fi; \
			echo "Version $$PACKAGE_VERSION deleted successfully."; \
			echo "Publishing new version..."; \
			npm publish --registry=https://npm.pkg.github.com; \
			if [ $$? -ne 0 ]; then \
				echo "Error: Failed to publish package"; \
				exit 1; \
			fi; \
			echo "Package published successfully!"; \
		elif [ "$$option" = "2" ]; then \
			echo "Publication cancelled."; \
			exit 0; \
		else \
			echo "Invalid option. Publication cancelled."; \
			exit 1; \
		fi; \
	else \
		echo "Version $$PACKAGE_VERSION does not exist on GitHub NPM. Publishing..."; \
		npm publish --registry=https://npm.pkg.github.com; \
		if [ $$? -ne 0 ]; then \
			echo "Error: Failed to publish package"; \
			exit 1; \
		fi; \
		echo "Package published successfully!"; \
	fi

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


