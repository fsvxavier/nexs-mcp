# Plano de Testes - MCP Server Go

**Versão:** 1.0  
**Data:** 18 de Dezembro de 2025  
**Meta:** Cobertura mínima de 98%

## Visão Geral

Este plano de testes cobre:
- **MCP SDK Integration:** Testes de integração com `modelcontextprotocol/go-sdk`
- **Schema Auto-generation:** Validação de schemas gerados via reflection
- **Multiple Transports:** Stdio, SSE, HTTP (todos os 3 transportes do SDK)
- **Validation Framework:** Testes de struct tags e validation automática
- **Domain Logic:** Testes unitários de regras de negócio

## Índice
1. [Estratégia de Testes](#estratégia-de-testes)
2. [Pirâmide de Testes](#pirâmide-de-testes)
3. [Tipos de Testes](#tipos-de-testes)
4. [Estrutura de Testes](#estrutura-de-testes)
5. [Ferramentas e Frameworks](#ferramentas-e-frameworks)
6. [CI/CD Integration](#cicd-integration)

---

## Estratégia de Testes

### Objetivos

1. **Cobertura:** ≥ 98% de cobertura de código
2. **Qualidade:** Zero bugs críticos em produção
3. **Performance:** Testes devem rodar em < 60 segundos
4. **Confiabilidade:** Testes determinísticos (sem flakiness)

### Princípios

- **Test-Driven Development (TDD):** Escrever testes antes do código
- **Isolation:** Testes unitários completamente isolados
- **Fast Feedback:** Testes rápidos (< 30s timeout por teste)
- **Readable:** Testes como documentação viva

---

## Pirâmide de Testes

```
                    ┌──────────────┐
                    │   E2E Tests  │  5% (50 tests)
                    │  Slow, Brittle│
                    └──────────────┘
                ┌────────────────────┐
                │ Integration Tests   │  15% (150 tests)
                │  Medium Speed       │
                └────────────────────┘
          ┌──────────────────────────────┐
          │      Unit Tests               │  80% (800 tests)
          │   Fast, Reliable, Isolated    │
          └──────────────────────────────┘
```

### Distribuição

| Tipo | Quantidade | % do Total | Tempo Médio |
|------|-----------|-----------|-------------|
| Unit | ~800 | 80% | < 10ms |
| Integration | ~150 | 15% | < 100ms |
| E2E | ~50 | 5% | < 1s |
| **Total** | **~1000** | **100%** | **< 60s** |

---

## Tipos de Testes

### 1. Testes Unitários (80%)

**Objetivo:** Testar unidades individuais (funções, métodos) em isolamento total

**Características:**
- Zero I/O (filesystem, network, database)
- Mocks para todas as dependências
- Extremamente rápidos (< 10ms cada)
- Determinísticos (sempre mesmo resultado)

**Exemplo:**
```go
// internal/domain/element_test.go
package domain_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/fsvxavier/mcp-server/internal/domain"
)

func TestElement_Validate_ValidPersona(t *testing.T) {
    // Arrange
    elem := &domain.Element{
        ID:          "persona_test_alice_20251218-120000",
        Type:        domain.PersonaElement,
        Name:        "test-persona",
        Version:     "1.0.0",
        Author:      "alice",
        Description: "Test persona",
        Content:     "# Test\n\nValid content",
    }
    
    validator := domain.NewElementValidator()
    
    // Act
    err := validator.Validate(elem)
    
    // Assert
    assert.NoError(t, err)
}

func TestSchemaGeneration_ListElementsInput(t *testing.T) {
    // Arrange
    reflector := jsonschema.NewReflector()
    
    // Act
    schema := reflector.Reflect(&ListElementsInput{})
    
    // Assert
    assert.Equal(t, "object", schema.Type)
    assert.Contains(t, schema.Required, "type")
    assert.Equal(t, []string{"personas", "skills", "templates", "agents", "memories", "ensembles"}, 
        schema.Properties["type"].Enum)
}

func TestToolValidation_AutomaticFromTags(t *testing.T) {
    // Arrange
    input := ListElementsInput{
        Type: "invalid_type", // should fail validation
    }
    validator := validator.New()
    
    // Act
    err := validator.Struct(input)
    
    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "oneof")
}
```

**Características de Testes Unitários:**
- Zero I/O (filesystem, network, database)
- Mocks para todas as dependências
- Extremamente rápidos (< 10ms cada)
- Determinísticos (sempre mesmo resultado)
- **Testes de schema generation via reflection**
- **Testes de validation tags**
    result := validator.Validate(elem)
    
    // Assert
    assert.True(t, result.Valid)
    assert.Empty(t, result.Errors)
}

func TestElement_Validate_InvalidName_TooLong(t *testing.T) {
    // Arrange
    elem := &domain.Element{
        Name: strings.Repeat("a", 101), // > 100 chars
        Type: domain.PersonaElement,
    }
    
    validator := domain.NewElementValidator()
    
    // Act
    result := validator.Validate(elem)
    
    // Assert
    assert.False(t, result.Valid)
    assert.Contains(t, result.Errors, "name exceeds maximum length")
}

func TestElement_Validate_SecurityVulnerability_PathTraversal(t *testing.T) {
    // Arrange
    elem := &domain.Element{
        Content: "../../etc/passwd", // path traversal attempt
        Type:    domain.PersonaElement,
    }
    
    validator := domain.NewElementValidator()
    
    // Act
    result := validator.Validate(elem)
    
    // Assert
    assert.False(t, result.Valid)
    assert.Contains(t, result.Errors, "path traversal detected")
}
```

**Casos de Teste por Função:**

```go
// Table-driven tests para cobertura completa
func TestElementValidator_Validate(t *testing.T) {
    tests := []struct {
        name      string
        element   *domain.Element
        wantValid bool
        wantError string
    }{
        {
            name: "valid persona",
            element: &domain.Element{
                Name: "valid-name",
                Type: domain.PersonaElement,
                Content: "Valid content",
            },
            wantValid: true,
        },
        {
            name: "empty name",
            element: &domain.Element{
                Name: "",
                Type: domain.PersonaElement,
            },
            wantValid: false,
            wantError: "name is required",
        },
        {
            name: "invalid characters in name",
            element: &domain.Element{
                Name: "invalid_name!", // underscore and ! not allowed
                Type: domain.PersonaElement,
            },
            wantValid: false,
            wantError: "invalid characters in name",
        },
        {
            name: "YAML bomb",
            element: &domain.Element{
                Content: generateYAMLBomb(), // expansion ratio > 5:1
                Type: domain.PersonaElement,
            },
            wantValid: false,
            wantError: "potential YAML bomb",
        },
        // ... mais 50+ casos
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := domain.NewElementValidator()
            result := validator.Validate(tt.element)
            
            assert.Equal(t, tt.wantValid, result.Valid)
            if tt.wantError != "" {
                assert.Contains(t, result.Errors, tt.wantError)
            }
        })
    }
}
```

---

### 2. Testes de Integração (15%)

**Objetivo:** Testar interação entre componentes

**Características:**
- I/O limitado (filesystem temporário, mock HTTP)
- Testa fluxos completos use case → repository → filesystem
- Médio tempo de execução (< 100ms)
- Usa fixtures reais

**Exemplo:**
```go
// test/integration/element_lifecycle_test.go
//go:build integration

package integration_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/fsvxavier/mcp-server/internal/usecase"
    "github.com/fsvxavier/mcp-server/internal/infrastructure"
)

func TestElementLifecycle_CreateEditDelete(t *testing.T) {
    // Arrange - Setup temporary directory
    tmpDir, err := os.MkdirTemp("", "mcp-test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)
    
    repo := infrastructure.NewFilesystemElementRepository(tmpDir)
    validator := domain.NewElementValidator()
    indexer := domain.NewSearchIndexer()
    
    createUC := usecase.NewCreateElementUseCase(repo, validator, indexer)
    editUC := usecase.NewEditElementUseCase(repo, validator, indexer)
    deleteUC := usecase.NewDeleteElementUseCase(repo, indexer)
    
    ctx := context.Background()
    
    // Act 1 - Create
    createReq := &usecase.CreateElementRequest{
        Type:        "personas",
        Name:        "test-persona",
        Description: "Test persona",
        Content:     "# Test\n\nContent",
        Author:      "alice",
    }
    
    createResp, err := createUC.Execute(ctx, createReq)
    require.NoError(t, err)
    require.NotEmpty(t, createResp.ID)
    
    // Assert 1 - File exists
    filePath := filepath.Join(tmpDir, "personas", "test-persona.md")
    assert.FileExists(t, filePath)
    
    // Act 2 - Edit
    editReq := &usecase.EditElementRequest{
        ID:      createResp.ID,
        Content: "# Updated\n\nNew content",
    }
    
    _, err = editUC.Execute(ctx, editReq)
    require.NoError(t, err)
    
    // Assert 2 - Content updated
    content, _ := os.ReadFile(filePath)
    assert.Contains(t, string(content), "New content")
    
    // Act 3 - Delete
    deleteReq := &usecase.DeleteElementRequest{ID: createResp.ID}
    err = deleteUC.Execute(ctx, deleteReq)
    require.NoError(t, err)
    
    // Assert 3 - File deleted
    assert.NoFileExists(t, filePath)
}

func TestMCPTransport_Stdio(t *testing.T) {
    // Arrange
    server, err := mcp.NewMCPServer("test-server", "1.0.0")
    require.NoError(t, err)
    
    // Simula stdin/stdout com pipes
    reader, writer := io.Pipe()
    originalStdin := os.Stdin
    originalStdout := os.Stdout
    os.Stdin = reader
    os.Stdout = writer
    defer func() {
        os.Stdin = originalStdin
        os.Stdout = originalStdout
    }()
    
    // Act - Envia request JSON-RPC via stdin
    go func() {
        request := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
        writer.Write([]byte(request + "\n"))
    }()
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    err = server.Start(ctx)
    
    // Assert
    assert.NoError(t, err)
}

func TestMCPTransport_SSE(t *testing.T) {
    // Arrange
    server := setupMCPServerWithSSE(t)
    defer server.Close()
    
    // Act - HTTP request para SSE endpoint
    resp, err := http.Get(server.URL + "/sse")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert
    assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
    assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
    
    // Read SSE events
    scanner := bufio.NewScanner(resp.Body)
    var events []string
    for scanner.Scan() {
        events = append(events, scanner.Text())
        if len(events) >= 3 { // connection + message + close
            break
        }
    }
    
    assert.GreaterOrEqual(t, len(events), 2)
}

func TestMCPTransport_HTTP(t *testing.T) {
    // Arrange
    server := setupMCPServerWithHTTP(t)
    defer server.Close()
    
    // Act - HTTP POST request
    body := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
    resp, err := http.Post(server.URL+"/rpc", "application/json", strings.NewReader(body))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "2.0", result["jsonrpc"])
    assert.Equal(t, float64(1), result["id"])
}

func TestPrivatePersona_UserIsolation(t *testing.T) {
    // Arrange
    tmpDir, _ := os.MkdirTemp("", "mcp-test-*")
    defer os.RemoveAll(tmpDir)
    
    service := setupPrivatePersonaService(t, tmpDir)
    
    // Act - Alice creates private persona
    aliceCtx := auth.ContextWithUser(context.Background(), "alice")
    alicePersona, err := service.CreatePrivatePersona(aliceCtx, "alice", CreatePersonaInput{
        Name: "work-helper",
        Description: "Work assistant",
        Content: "# Work Helper",
        PrivacyLevel: "private",
    })
    require.NoError(t, err)
    
    // Assert - Persona saved in alice's directory
    expectedPath := filepath.Join(tmpDir, "personas", "private-alice", "work-helper.md")
    assert.FileExists(t, expectedPath)
    
    // Act - Bob tries to access Alice's private persona
    bobCtx := auth.ContextWithUser(context.Background(), "bob")
    _, err = service.GetPersona(bobCtx, alicePersona.ID)
    
    // Assert - Access denied
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "access denied")
}

func TestPrivatePersona_ForkWorkflow(t *testing.T) {
    // Arrange
    service := setupPrivatePersonaService(t, tmpDir)
    aliceCtx := auth.ContextWithUser(context.Background(), "alice")
    bobCtx := auth.ContextWithUser(context.Background(), "bob")
    
    // Alice creates and shares persona
    alicePersona, _ := service.CreatePrivatePersona(aliceCtx, "alice", CreatePersonaInput{
        Name: "shared-helper",
        PrivacyLevel: "shared",
    })
    service.SharePersona(aliceCtx, alicePersona.ID, []string{"bob"}, Permissions{Read: true, Fork: true})
    
    // Act - Bob forks Alice's persona
    fork, err := service.ForkPersona(bobCtx, alicePersona.ID, "bob", map[string]interface{}{
        "name": "my-helper",
        "behavioral_traits": []string{"technical", "precise"},
    })
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, "bob", fork.Owner)
    assert.Equal(t, alicePersona.ID, fork.ForkedFrom)
    assert.Equal(t, "private", fork.PrivacyLevel)
    assert.FileExists(t, filepath.Join(tmpDir, "personas", "private-bob", "my-helper.md"))
}

func TestPrivatePersona_BulkImport(t *testing.T) {
    // Arrange
    service := setupPrivatePersonaService(t, tmpDir)
    aliceCtx := auth.ContextWithUser(context.Background(), "alice")
    
    csvData := `name,description,behavioral_traits
dev-helper,Development assistant,"technical,precise"
writer,Content creator,"creative,empathetic"`
    
    // Act
    result, err := service.BulkImportPersonas(aliceCtx, "alice", BulkImportInput{
        Format: "csv",
        Data: csvData,
        PrivacyLevel: "private",
    })
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, 2, result.Imported)
    assert.Equal(t, 0, result.Skipped)
    assert.FileExists(t, filepath.Join(tmpDir, "personas", "private-alice", "dev-helper.md"))
    assert.FileExists(t, filepath.Join(tmpDir, "personas", "private-alice", "writer.md"))
}
```

**Cobertura de Integração:**
- Element lifecycle (create, edit, delete)
- Portfolio sync (local ↔ GitHub)
- Search indexing (inverted index + NLP)
- **MCP SDK Stdio transport**
- **MCP SDK SSE transport**
- **MCP SDK HTTP transport**
- Schema validation end-to-end
- **Private Personas:**
  - User directory isolation
  - Access control enforcement
  - Share/Fork workflows
  - Version control operations
  - Bulk operations (import/export)
  - Advanced search/filtering
    editUC := usecase.NewEditElementUseCase(repo, validator)
    deleteUC := usecase.NewDeleteElementUseCase(repo)
    
    ctx := context.Background()
    
    // Act 1 - Create element
    input := usecase.CreateElementInput{
        Type:        "personas",
        Name:        "test-persona",
        Author:      "test-user",
        Description: "Test persona",
        Content:     "# Test\n\nContent",
    }
    
    created, err := createUC.Execute(ctx, input)
    require.NoError(t, err)
    require.NotNil(t, created)
    
    // Assert 1 - File exists
    expectedPath := filepath.Join(tmpDir, "personas", "test-persona.md")
    assert.FileExists(t, expectedPath)
    
    // Act 2 - Edit element
    editInput := usecase.EditElementInput{
        ID: created.ID,
        Fields: map[string]interface{}{
            "description": "Updated description",
        },
    }
    
    edited, err := editUC.Execute(ctx, editInput)
    require.NoError(t, err)
    
    // Assert 2 - Version incremented
    assert.Equal(t, "1.0.1", edited.Version)
    
    // Act 3 - Delete element
    err = deleteUC.Execute(ctx, created.ID)
    require.NoError(t, err)
    
    // Assert 3 - File removed
    assert.NoFileExists(t, expectedPath)
}

func TestPortfolioSync_GitHubIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping GitHub integration test in short mode")
    }
    
    // This test requires GITHUB_TOKEN env var
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    
    // Test bidirectional sync with real GitHub API
    // Uses mock repository to avoid polluting production data
}
```

**Estrutura de Teste de Integração:**
```
test/
├── integration/
│   ├── element_lifecycle_test.go
│   ├── portfolio_sync_test.go
│   ├── search_index_test.go
│   ├── security_validation_test.go
│   └── fixtures/
│       ├── personas/
│       │   ├── valid-persona.md
│       │   └── invalid-persona.md
│       ├── skills/
│       └── memories/
├── e2e/
└── helpers/
    ├── testutil.go
    └── fixtures.go
```

---

### 3. Testes End-to-End (5%)

**Objetivo:** Testar sistema completo como usuário real

**Características:**
- Testa via MCP protocol (JSON-RPC)
- Usa servidor real rodando em processo
- Simula cliente Claude Desktop
- Mais lentos (< 1s cada)

**Exemplo:**
```go
// test/e2e/mcp_server_test.go
//go:build e2e

package e2e_test

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "os/exec"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMCPServer_ListElements_E2E(t *testing.T) {
    // Arrange - Start MCP server
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "mcp-server")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    
    require.NoError(t, cmd.Start())
    defer cmd.Process.Kill()
    
    // Act - Send JSON-RPC request
    request := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      1,
        "method":  "tools/call",
        "params": map[string]interface{}{
            "name": "list_elements",
            "arguments": map[string]interface{}{
                "type": "personas",
            },
        },
    }
    
    reqData, _ := json.Marshal(request)
    _, err := stdin.Write(append(reqData, '\n'))
    require.NoError(t, err)
    
    // Read response
    var buf bytes.Buffer
    io.CopyN(&buf, stdout, 4096) // read first 4KB
    
    var response map[string]interface{}
    err = json.Unmarshal(buf.Bytes(), &response)
    require.NoError(t, err)
    
    // Assert
    assert.Equal(t, "2.0", response["jsonrpc"])
    assert.Equal(t, float64(1), response["id"])
    assert.NotNil(t, response["result"])
}

func TestMCPServer_CreateElementWorkflow_E2E(t *testing.T) {
    // Test complete workflow:
    // 1. Create element
    // 2. List to verify
    // 3. Activate element
    // 4. Get active elements
    // 5. Deactivate
    // 6. Delete
}
```

---

### 4. Testes de Benchmark (Performance)

**Objetivo:** Garantir performance targets

**Exemplo:**
```go
// internal/domain/element_benchmark_test.go
package domain_test

import (
    "testing"
    "github.com/fsvxavier/mcp-server/internal/domain"
)

func BenchmarkElement_Validate(b *testing.B) {
    elem := &domain.Element{
        Name:    "test-persona",
        Type:    domain.PersonaElement,
        Content: "# Test\n\nContent here",
    }
    
    validator := domain.NewElementValidator()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validator.Validate(elem)
    }
}

func BenchmarkSearchIndexer_Search_1000Elements(b *testing.B) {
    // Setup 1000 elements
    indexer := domain.NewSearchIndexer()
    for i := 0; i < 1000; i++ {
        elem := generateTestElement(i)
        indexer.Index(elem)
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        indexer.Search("debugging")
    }
    // Target: < 10ms per search
}

func BenchmarkPortfolioSync_Bidirectional(b *testing.B) {
    // Target: < 2s for 100 elements
}
```

---

### 5. Testes de Segurança

**Objetivo:** Validar todas as 300+ regras de segurança

**Exemplo:**
```go
// test/security/injection_test.go
package security_test

func TestSecurity_PathTraversal_AllVariations(t *testing.T) {
    attacks := []string{
        "../../etc/passwd",
        "..\\..\\windows\\system32",
        "....//....//etc/passwd",
        "%2e%2e%2f%2e%2e%2f",
        "..;/..;/etc/passwd",
    }
    
    validator := security.NewPathValidator()
    
    for _, attack := range attacks {
        t.Run(attack, func(t *testing.T) {
            err := validator.Validate(attack)
            assert.Error(t, err, "should detect path traversal")
        })
    }
}

func TestSecurity_CommandInjection_AllVariations(t *testing.T) {
    attacks := []string{
        "; rm -rf /",
        "| cat /etc/passwd",
        "`whoami`",
        "$(whoami)",
    }
    
    for _, attack := range attacks {
        // Test against content validation
    }
}

func TestSecurity_YAMLBomb_Detection(t *testing.T) {
    // Generate YAML with expansion ratio > 5:1
    bomb := generateYAMLBomb(10) // 10 levels deep
    
    validator := security.NewYAMLValidator()
    err := validator.Validate(bomb)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "YAML bomb")
}

func TestSecurity_PrototypePollution_Prevention(t *testing.T) {
    maliciousJSON := `{
        "__proto__": {
            "isAdmin": true
        }
    }`
    
    var metadata map[string]interface{}
    err := json.Unmarshal([]byte(maliciousJSON), &metadata)
    require.NoError(t, err)
    
    sanitized := security.SanitizeMetadata(metadata)
    
    assert.NotContains(t, sanitized, "__proto__")
    assert.NotContains(t, sanitized, "constructor")
    assert.NotContains(t, sanitized, "prototype")
}
```

---

## Estrutura de Testes

### Organização de Arquivos

```
mcp-server/
├── internal/
│   ├── domain/
│   │   ├── element.go
│   │   ├── element_test.go        # Testes unitários
│   │   ├── element_benchmark_test.go
│   │   └── validation.go
│   │       └── validation_test.go
│   ├── usecase/
│   │   ├── create_element.go
│   │   └── create_element_test.go
│   └── infrastructure/
│       ├── filesystem_repo.go
│       └── filesystem_repo_test.go
├── test/
│   ├── integration/              # Testes de integração
│   │   ├── element_lifecycle_test.go
│   │   ├── portfolio_sync_test.go
│   │   └── fixtures/
│   ├── e2e/                      # Testes end-to-end
│   │   ├── mcp_server_test.go
│   │   └── claude_integration_test.go
│   ├── security/                 # Testes de segurança
│   │   ├── injection_test.go
│   │   ├── traversal_test.go
│   │   └── yaml_bomb_test.go
│   ├── performance/              # Benchmarks
│   │   └── load_test.go
│   └── helpers/                  # Utilities
│       ├── testutil.go
│       ├── fixtures.go
│       └── mocks/
│           ├── element_repo_mock.go
│           └── github_client_mock.go
```

### Convenções de Nomenclatura

```go
// Testes unitários
func Test<Type>_<Method>_<Scenario>(t *testing.T)

// Exemplos:
func TestElement_Validate_ValidPersona(t *testing.T)
func TestElement_Validate_InvalidName(t *testing.T)
func TestElementValidator_Validate_EmptyContent(t *testing.T)

// Testes de integração
func TestIntegration_<Feature>_<Scenario>(t *testing.T)

// Exemplo:
func TestIntegration_ElementLifecycle_CreateEditDelete(t *testing.T)

// Benchmarks
func Benchmark<Type>_<Method>(b *testing.B)

// Exemplo:
func BenchmarkSearchIndexer_Search_1000Elements(b *testing.B)
```

---

## Ferramentas e Frameworks

### Testing Framework

```go
// Standard library
import "testing"

// Assertions
import "github.com/stretchr/testify/assert"
import "github.com/stretchr/testify/require"

// Mocking
import "github.com/stretchr/testify/mock"

// Table-driven tests
tests := []struct {
    name string
    input string
    want string
}{ /* ... */ }
```

### Mocking

```go
// test/helpers/mocks/element_repo_mock.go
package mocks

type MockElementRepository struct {
    mock.Mock
}

func (m *MockElementRepository) Save(ctx context.Context, elem *domain.Element) error {
    args := m.Called(ctx, elem)
    return args.Error(0)
}

func (m *MockElementRepository) FindByID(ctx context.Context, id string) (*domain.Element, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Element), args.Error(1)
}

// Usage in tests:
func TestCreateElement_Success(t *testing.T) {
    mockRepo := new(mocks.MockElementRepository)
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
    
    uc := usecase.NewCreateElementUseCase(mockRepo, nil, nil)
    _, err := uc.Execute(ctx, input)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Test Fixtures

```go
// test/helpers/fixtures.go
package helpers

func LoadFixture(path string) (*domain.Element, error) {
    data, err := os.ReadFile(filepath.Join("test/fixtures", path))
    if err != nil {
        return nil, err
    }
    
    var elem domain.Element
    if err := yaml.Unmarshal(data, &elem); err != nil {
        return nil, err
    }
    
    return &elem, nil
}

// Usage:
elem, err := helpers.LoadFixture("personas/valid-persona.md")
```

### Coverage Tool

```bash
# Run tests with coverage
go test -v -race -timeout 30s -coverprofile=coverage.out ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Check coverage threshold
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
# Must be >= 98%
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  unit-tests:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.23]
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Run unit tests
        run: |
          go test -v -race -timeout 30s \
            -coverprofile=coverage.out \
            ./internal/... ./pkg/...
      
      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 98" | bc -l) )); then
            echo "Coverage below 98%!"
            exit 1
          fi
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.23
      
      - name: Run integration tests
        run: |
          go test -v -race -timeout 30s \
            -tags=integration \
            ./test/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
      
      - name: Build server
        run: go build -o bin/mcp-server cmd/mcp-server/main.go
      
      - name: Run E2E tests
        run: |
          go test -v -timeout 60s \
            -tags=e2e \
            ./test/e2e/...

  security-tests:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Run security tests
        run: |
          go test -v ./test/security/...
      
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: ./...
```

---

## Métricas de Qualidade

### Targets

| Métrica | Target | Comando |
|---------|--------|---------|
| Cobertura de código | ≥ 98% | `go test -cover ./...` |
| Testes passando | 100% | `go test ./...` |
| Tempo de execução | < 60s | `go test -timeout 60s ./...` |
| Race conditions | 0 | `go test -race ./...` |
| Linting issues | 0 | `golangci-lint run` |

### Comandos de Teste

```bash
# Todos os testes
make test-all

# Apenas unit tests
make test-unit

# Testes de integração
make test-integration

# E2E tests
make test-e2e

# Benchmarks
make bench

# Coverage report
make coverage

# Security tests
make test-security
```

### Makefile

```makefile
.PHONY: test-all test-unit test-integration test-e2e bench coverage

test-all: test-unit test-integration test-e2e

test-unit:
	go test -v -race -timeout 30s -coverprofile=coverage.out ./internal/... ./pkg/...

test-integration:
	go test -v -race -timeout 30s -tags=integration ./test/integration/...

test-e2e:
	go test -v -timeout 60s -tags=e2e ./test/e2e/...

bench:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-security:
	go test -v ./test/security/...
	gosec -quiet ./...

lint:
	golangci-lint run --timeout 5m
```

---

**Próximo Documento:** [Guia de Implementação](./IMPLEMENTATION_GUIDE.md)
