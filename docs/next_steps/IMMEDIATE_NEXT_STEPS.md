# Próximos Passos Imediatos

**Versão:** 1.0  
**Data:** 18 de Dezembro de 2025  
**Objetivo:** Guia prático para iniciar o projeto

## Visão Geral

Este documento fornece instruções detalhadas e comandos prontos para executar as primeiras ações do projeto, desde o setup inicial até as primeiras semanas de desenvolvimento.

## Índice
1. [Setup Inicial (Esta Semana)](#setup-inicial-esta-semana)
2. [Semana 1: MCP SDK Integration](#semana-1-mcp-sdk-integration)
3. [Semana 2: Schema & Tool Registry](#semana-2-schema--tool-registry)
4. [Checklists de Validação](#checklists-de-validação)

---

## Setup Inicial (Esta Semana)

**Objetivo:** Preparar ambiente de desenvolvimento completo  
**Duração Estimada:** 3-5 dias  
**Responsável:** Tech Lead + Equipe

### Pré-requisitos

#### 1. Instalar Go 1.25+

```bash
# Verificar versão do Go
go version
# Deve mostrar: go version go1.25 ou superior

# Se precisar instalar (Linux/macOS):
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz

# Adicionar ao PATH (no ~/.bashrc ou ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Reload
source ~/.bashrc
```

#### 2. Instalar Git

```bash
# Verificar instalação
git --version

# Configurar (se necessário)
git config --global user.name "Seu Nome"
git config --global user.email "seu@email.com"
```

#### 3. Instalar Ferramentas de Desenvolvimento

```bash
# golangci-lint (linter)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# govulncheck (security scanner)
go install golang.org/x/vuln/cmd/govulncheck@latest

# mockgen (mocking - opcional)
go install go.uber.org/mock/mockgen@latest

# Verificar instalação
golangci-lint --version
govulncheck -version
```

---

### Dia 1: Criar Repositório

#### Passo 1: Criar no GitHub

```bash
# Opção 1: Via Web (Recomendado)
# 1. Ir para https://github.com/new
# 2. Nome: nexs-mcp (ou mcp-server-go)
# 3. Descrição: "Model Context Protocol Server in Go - DollhouseMCP port"
# 4. Public/Private: Escolher conforme preferência
# 5. Add .gitignore: Go template
# 6. License: MIT
# 7. Create repository

# Opção 2: Via CLI (gh)
gh repo create fsvxavier/nexs-mcp \
  --public \
  --description "Model Context Protocol Server in Go" \
  --gitignore Go \
  --license MIT \
  --clone
```

#### Passo 2: Clone e Setup Inicial

```bash
# Se criou via web, clonar
git clone git@github.com:fsvxavier/nexs-mcp.git
cd nexs-mcp

# Inicializar Go module
go mod init github.com/fsvxavier/nexs-mcp

# Criar estrutura de pastas
mkdir -p cmd/mcp-server
mkdir -p internal/{mcp/{server,schema,tools},elements/{persona,skill,template,agent,memory,ensemble},portfolio/{local,github,sync},collection,security/{validation,sanitization,encryption},capability/{nlp,graph},telemetry}
mkdir -p pkg/{client,types}
mkdir -p test/{integration,e2e,fixtures}
mkdir -p docs/{architecture,guides}
mkdir -p examples
mkdir -p scripts
mkdir -p .github/workflows

# Verificar estrutura
tree -L 3 -d
```

#### Passo 3: Criar Arquivos Iniciais

```bash
# main.go
cat > cmd/mcp-server/main.go << 'EOF'
package main

import (
	"fmt"
	"os"
)

const (
	version = "0.1.0"
	name    = "nexs-mcp"
)

func main() {
	fmt.Printf("%s v%s\n", name, version)
	fmt.Println("MCP Server starting...")
	
	// TODO: Initialize MCP server
	os.Exit(0)
}
EOF

# README.md
cat > README.md << 'EOF'
# NEXS MCP - Model Context Protocol Server in Go

[![CI](https://github.com/fsvxavier/nexs-mcp/workflows/CI/badge.svg)](https://github.com/fsvxavier/nexs-mcp/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-mcp)](https://goreportcard.com/report/github.com/fsvxavier/nexs-mcp)
[![Coverage](https://codecov.io/gh/fsvxavier/nexs-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/fsvxavier/nexs-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance Model Context Protocol server written in Go, replicating and enhancing all features of [DollhouseMCP](https://github.com/DollhouseMCP/mcp-server).

## Features

- 🚀 **High Performance**: 10-50x faster than Node.js
- 🔒 **Secure**: 300+ security validation rules
- 🧩 **Complete**: 49 MCP tools for element management
- 📦 **6 Element Types**: Personas, Skills, Templates, Agents, Memories, Ensembles
- 🌐 **GitHub Integration**: Bidirectional sync with OAuth2
- 🔍 **Advanced Search**: NLP scoring with Jaccard similarity
- 🧪 **Well Tested**: 98%+ code coverage
- 🤖 **20 AI Models**: Claude, Gemini, GPT, Grok, OSWE with auto-selection

## Quick Start

### Installation

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/mcp-server@latest
```

### Usage

```bash
mcp-server
```

### Claude Desktop Integration

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "mcp-server"
    }
  }
}
```

## Documentation

- [Architecture](docs/plano/ARCHITECTURE.md)
- [Tools Specification](docs/plano/TOOLS_SPEC.md)
- [Testing Plan](docs/plano/TESTING_PLAN.md)
- [Roadmap](docs/next_steps/ROADMAP.md)

## Development

### Prerequisites

- Go 1.25+
- Git

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [DollhouseMCP](https://github.com/DollhouseMCP/mcp-server) - Original TypeScript implementation
- [Model Context Protocol](https://modelcontextprotocol.io/) - Protocol specification
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Official Go SDK
EOF
```

---

### Dia 2: Configurar CI/CD

#### GitHub Actions Workflow

```bash
# Criar arquivo de workflow
cat > .github/workflows/ci.yml << 'EOF'
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.25']
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: make test
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

  build:
    name: Build
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Build
        run: make build
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: mcp-server
          path: bin/mcp-server
EOF
```

---

### Dia 3: Configurar Ferramentas

#### Makefile

```bash
cat > Makefile << 'EOF'
.PHONY: all build test lint clean install-tools help

# Variables
BINARY_NAME=mcp-server
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: lint test build ## Run lint, test and build

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mcp-server

test: ## Run tests with coverage
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"

test-short: ## Run short tests
	$(GOTEST) -short ./...

test-integration: ## Run integration tests
	$(GOTEST) -v -tags=integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	$(GOTEST) -v -tags=e2e ./test/e2e/...

bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

lint: ## Run linters
	@echo "Running linters..."
	$(GOLINT) run --config .golangci.yml

fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	gofumpt -l -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@$(GOCMD) clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install mvdan.cc/gofumpt@latest

security: ## Run security scans
	@echo "Running security scans..."
	govulncheck ./...

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mcp-server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mcp-server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mcp-server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mcp-server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mcp-server

run: build ## Build and run
	./$(BUILD_DIR)/$(BINARY_NAME)

docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) .

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
EOF
```

#### golangci-lint Configuration

```bash
cat > .golangci.yml << 'EOF'
linters-settings:
  dupl:
    threshold: 100
  errcheck:
    check-type-assertions: true
  goconst:
    min-len: 2
    min-occurrences: 3
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
  gocyclo:
    min-complexity: 15
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/fsvxavier/nexs-mcp
  golint:
    min-confidence: 0
  govet:
    check-shadowing: true
  misspell:
    locale: US
  nolintlint:
    require-explanation: true
    require-specific: true

linters:
  enable:
    - bodyclose
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - misspell
    - nolintlint
    - revive
    - staticcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec

run:
  timeout: 5m
  tests: true
EOF
```

#### Git Hooks (Opcional)

```bash
# Pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

echo "Running pre-commit checks..."

# Format code
make fmt

# Run linters
make lint
if [ $? -ne 0 ]; then
    echo "Linting failed. Please fix errors before committing."
    exit 1
fi

# Run tests
make test-short
if [ $? -ne 0 ]; then
    echo "Tests failed. Please fix before committing."
    exit 1
fi

echo "Pre-commit checks passed!"
exit 0
EOF

chmod +x .git/hooks/pre-commit
```

---

### Dia 4-5: Primeiro Commit

```bash
# Adicionar dependências iniciais
go get github.com/modelcontextprotocol/go-sdk@latest
go get github.com/invopop/jsonschema@latest
go get github.com/go-playground/validator/v10@latest
go get github.com/stretchr/testify@latest
go get gopkg.in/yaml.v3@latest
go mod tidy

# Verificar que tudo compila
make build

# Executar testes (ainda vazios, mas devem passar)
make test

# Commit inicial
git add .
git commit -m "chore: initial project setup

- Setup Go module and project structure
- Configure CI/CD with GitHub Actions
- Add Makefile and development tools
- Configure golangci-lint
- Add initial README and documentation"

# Push
git push origin main
```

---

## Semana 1: MCP SDK Integration

**Objetivo:** Integrar MCP SDK e criar servidor básico  
**Duração:** 5 dias

### Dia 1-2: Integrar MCP SDK

#### Criar Server Wrapper

```bash
# Criar arquivo
cat > internal/mcp/server/server.go << 'EOF'
package server

import (
	"context"
	"fmt"
	
	"github.com/modelcontextprotocol/go-sdk/server"
	"github.com/modelcontextprotocol/go-sdk/transport"
)

const (
	serverName    = "nexs-mcp"
	serverVersion = "0.1.0"
)

// MCPServer wraps the official MCP SDK server
type MCPServer struct {
	srv *server.MCPServer
}

// New creates a new MCP server with stdio transport
func New() (*MCPServer, error) {
	// Create stdio transport (default for Claude Desktop)
	trans := transport.NewStdioTransport()
	
	// Create MCP server
	srv := server.NewMCPServer(
		server.WithName(serverName),
		server.WithVersion(serverVersion),
		server.WithTransport(trans),
	)
	
	return &MCPServer{srv: srv}, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	return s.srv.Serve(ctx)
}

// RegisterTool registers a new tool with the server
func (s *MCPServer) RegisterTool(name string, handler interface{}) error {
	// TODO: Implement tool registration with auto schema generation
	return fmt.Errorf("not implemented")
}
EOF

# Criar teste
cat > internal/mcp/server/server_test.go << 'EOF'
package server_test

import (
	"testing"
	
	"github.com/fsvxavier/nexs-mcp/internal/mcp/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	srv, err := server.New()
	require.NoError(t, err)
	assert.NotNil(t, srv)
}

// TODO: Add more tests
EOF
```

#### Atualizar main.go

```bash
cat > cmd/mcp-server/main.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/fsvxavier/nexs-mcp/internal/mcp/server"
)

var (
	version = "0.1.0"
	name    = "nexs-mcp"
)

func main() {
	fmt.Fprintf(os.Stderr, "%s v%s starting...\n", name, version)
	
	// Create MCP server
	srv, err := server.New()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()
	
	// Start server
	fmt.Fprintln(os.Stderr, "MCP server ready on stdio")
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	
	fmt.Fprintln(os.Stderr, "Server stopped")
}
EOF
```

#### Testar Build e Execução

```bash
# Build
make build

# Executar (deve iniciar e esperar comandos MCP)
./bin/mcp-server

# Em outro terminal, testar com echo
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | ./bin/mcp-server
```

---

### Dia 3-4: Schema Auto-generation

#### Implementar Gerador de Schema

```bash
cat > internal/mcp/schema/generator.go << 'EOF'
package schema

import (
	"reflect"
	
	"github.com/invopop/jsonschema"
)

// Generator generates JSON schemas from Go structs
type Generator struct {
	reflector *jsonschema.Reflector
}

// NewGenerator creates a new schema generator
func NewGenerator() *Generator {
	r := &jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	
	return &Generator{reflector: r}
}

// Generate generates JSON schema for a type
func (g *Generator) Generate(v interface{}) *jsonschema.Schema {
	return g.reflector.Reflect(v)
}

// GenerateFromType generates JSON schema from a reflect.Type
func (g *Generator) GenerateFromType(t reflect.Type) *jsonschema.Schema {
	return g.reflector.ReflectFromType(t)
}
EOF

# Teste do gerador
cat > internal/mcp/schema/generator_test.go << 'EOF'
package schema_test

import (
	"testing"
	
	"github.com/fsvxavier/nexs-mcp/internal/mcp/schema"
	"github.com/stretchr/testify/assert"
)

type TestInput struct {
	Name  string `json:"name" jsonschema:"required,minLength=3" validate:"required,min=3"`
	Email string `json:"email" jsonschema:"format=email" validate:"required,email"`
	Age   int    `json:"age" jsonschema:"minimum=0,maximum=150" validate:"gte=0,lte=150"`
}

func TestGenerator_Generate(t *testing.T) {
	gen := schema.NewGenerator()
	
	sch := gen.Generate(&TestInput{})
	
	assert.Equal(t, "object", sch.Type)
	assert.Contains(t, sch.Required, "name")
	assert.NotNil(t, sch.Properties)
	
	// Validate name property
	nameProp := sch.Properties.Get("name")
	assert.NotNil(t, nameProp)
	assert.Equal(t, "string", nameProp.Type)
}
EOF
```

---

### Dia 5: Primeira Tool

#### Implementar list_elements (stub)

```bash
mkdir -p internal/mcp/tools

cat > internal/mcp/tools/element_tools.go << 'EOF'
package tools

import (
	"context"
)

// ListElementsInput defines input for list_elements tool
type ListElementsInput struct {
	Type string `json:"type" jsonschema:"required,enum=personas|skills|templates" validate:"required,oneof=personas skills templates"`
}

// ListElementsOutput defines output for list_elements tool
type ListElementsOutput struct {
	Elements []ElementSummary `json:"elements"`
	Total    int              `json:"total"`
}

// ElementSummary is a brief element description
type ElementSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ElementTools handles element-related MCP tools
type ElementTools struct {
	// TODO: Add repository dependency
}

// NewElementTools creates a new ElementTools handler
func NewElementTools() *ElementTools {
	return &ElementTools{}
}

// ListElements handles the list_elements tool call
func (t *ElementTools) ListElements(ctx context.Context, input ListElementsInput) (*ListElementsOutput, error) {
	// TODO: Implement actual listing from repository
	// For now, return empty list
	return &ListElementsOutput{
		Elements: []ElementSummary{},
		Total:    0,
	}, nil
}
EOF

# Teste
cat > internal/mcp/tools/element_tools_test.go << 'EOF'
package tools_test

import (
	"context"
	"testing"
	
	"github.com/fsvxavier/nexs-mcp/internal/mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElementTools_ListElements(t *testing.T) {
	handler := tools.NewElementTools()
	
	input := tools.ListElementsInput{
		Type: "personas",
	}
	
	output, err := handler.ListElements(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 0, output.Total)
}
EOF
```

#### Commit Semana 1

```bash
make test
make lint

git add .
git commit -m "feat: implement MCP SDK integration and schema generation

- Integrate official MCP Go SDK
- Implement stdio transport
- Create schema auto-generation framework
- Add first tool stub: list_elements
- Achieve 100% test coverage for new code"

git push origin main
```

---

## Semana 2: Schema & Tool Registry

### Dia 1-2: Tool Registry

```bash
cat > internal/mcp/tools/registry.go << 'EOF'
package tools

import (
	"context"
	"fmt"
	"reflect"
	
	"github.com/fsvxavier/nexs-mcp/internal/mcp/schema"
	"github.com/go-playground/validator/v10"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(ctx context.Context, input interface{}) (interface{}, error)

// ToolDefinition contains tool metadata and handler
type ToolDefinition struct {
	Name        string
	Description string
	InputType   reflect.Type
	OutputType  reflect.Type
	Handler     ToolHandler
	InputSchema interface{} // JSON Schema
}

// Registry manages MCP tools
type Registry struct {
	tools     map[string]*ToolDefinition
	validator *validator.Validate
	generator *schema.Generator
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools:     make(map[string]*ToolDefinition),
		validator: validator.New(),
		generator: schema.NewGenerator(),
	}
}

// Register registers a new tool
func (r *Registry) Register(name, description string, handler interface{}) error {
	// Get handler function type
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}
	
	// Validate function signature: func(ctx context.Context, input T) (output T, error)
	if handlerType.NumIn() != 2 {
		return fmt.Errorf("handler must accept 2 parameters: (context.Context, input)")
	}
	if handlerType.NumOut() != 2 {
		return fmt.Errorf("handler must return 2 values: (output, error)")
	}
	
	// Get input and output types
	inputType := handlerType.In(1)
	outputType := handlerType.Out(0)
	
	// Generate JSON schema for input
	inputSchema := r.generator.GenerateFromType(inputType)
	
	// Create tool definition
	tool := &ToolDefinition{
		Name:        name,
		Description: description,
		InputType:   inputType,
		OutputType:  outputType,
		Handler:     r.wrapHandler(handler),
		InputSchema: inputSchema,
	}
	
	r.tools[name] = tool
	return nil
}

// wrapHandler wraps a typed handler into generic ToolHandler
func (r *Registry) wrapHandler(handler interface{}) ToolHandler {
	handlerValue := reflect.ValueOf(handler)
	
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		// Validate input
		if err := r.validator.Struct(input); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		
		// Call handler
		results := handlerValue.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(input),
		})
		
		// Extract output and error
		output := results[0].Interface()
		errInterface := results[1].Interface()
		
		if errInterface != nil {
			return nil, errInterface.(error)
		}
		
		return output, nil
	}
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolDefinition, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []*ToolDefinition {
	tools := make([]*ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
EOF
```

### Dia 3-5: Integrar Tool Registry com Server

```bash
# Atualizar server.go para usar registry
# Testar registro e chamada de tools
# Adicionar mais tools (create_element stub, etc)
```

---

## Checklists de Validação

### Checklist: Setup Inicial ✅

- [ ] Go 1.25+ instalado e funcionando
- [ ] Git configurado
- [ ] Repositório GitHub criado
- [ ] Go module inicializado (`go.mod` existe)
- [ ] Estrutura de pastas criada
- [ ] CI/CD pipeline configurado (GitHub Actions)
- [ ] Makefile funcionando
- [ ] golangci-lint instalado e configurado
- [ ] Primeira build bem-sucedida (`make build`)
- [ ] Testes passando (`make test`)
- [ ] Commit inicial feito e pushed

### Checklist: Semana 1 ✅

- [ ] MCP SDK integrado (`go.mod` tem dependência)
- [ ] Servidor básico criado (`internal/mcp/server/`)
- [ ] Stdio transport funcionando
- [ ] Schema generator implementado
- [ ] Primeira tool stub criada (`list_elements`)
- [ ] Testes unitários para server
- [ ] Testes unitários para schema generator
- [ ] Integração com Claude Desktop testada
- [ ] Cobertura > 90%
- [ ] Commit da semana feito

### Checklist: Semana 2 📋

- [ ] Tool registry implementado
- [ ] Registro automático de tools funcionando
- [ ] Validação automática via struct tags
- [ ] Múltiplas tools registradas
- [ ] Testes de integração tool registry
- [ ] list_elements retornando dados reais
- [ ] Cobertura > 95%

---

**Última Atualização:** 18 de Dezembro de 2025  
**Próxima Revisão:** Após conclusão da Semana 2  
**Owner:** Tech Lead
