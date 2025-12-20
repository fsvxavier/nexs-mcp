# Contributing to NEXS MCP

**Thank you for your interest in contributing to NEXS MCP!**

NEXS MCP is an open-source Model Context Protocol (MCP) server built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp v1.1.0`). We welcome contributions from the community to make this project better.

This document provides guidelines and instructions for contributing to the project.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Documentation Requirements](#documentation-requirements)
- [Community and Communication](#community-and-communication)

---

## Code of Conduct

### Our Pledge

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone, regardless of age, body size, visible or invisible disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Examples of behavior that contributes to a positive environment:**

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Examples of unacceptable behavior:**

- The use of sexualized language or imagery
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team. All complaints will be reviewed and investigated promptly and fairly.

---

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include as many details as possible:

**Bug Report Template:**

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. Use tool '...'
3. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Environment:**
- OS: [e.g., macOS 14.0, Ubuntu 22.04]
- Go version: [e.g., 1.21.5]
- NEXS MCP version: [e.g., 0.1.0]
- MCP SDK version: [e.g., v1.1.0]

**Additional context**
Add any other context about the problem here. Include error logs if available.
```

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear description** of the enhancement
- **Use case** - Why is this enhancement needed?
- **Proposed solution** - How would you implement it?
- **Alternatives considered** - What other approaches did you think about?

**Enhancement Request Template:**

```markdown
**Is your feature request related to a problem?**
A clear description of the problem. Ex. I'm frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear description of any alternative solutions or features.

**Additional context**
Add any other context, mockups, or examples about the feature request.
```

### Contributing Code

1. **Fork the repository**
2. **Create a feature branch** from `main`
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Submit a pull request**

### Contributing Documentation

Documentation improvements are highly valued:

- Fix typos, grammar, or unclear explanations
- Add examples and tutorials
- Improve API documentation
- Translate documentation (future)

### Participating in Discussions

- Answer questions from other users
- Share your use cases and experiences
- Provide feedback on proposals and RFCs
- Help review pull requests

---

## Getting Started

### Prerequisites

Before you begin, ensure you have:

- **Go 1.21 or later** - [Download](https://golang.org/dl/)
- **Git** - [Download](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Unix systems
- **golangci-lint** (optional) - For linting: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Initial Setup

1. **Fork the repository** on GitHub

2. **Clone your fork:**

```bash
git clone https://github.com/YOUR_USERNAME/nexs-mcp.git
cd nexs-mcp
```

3. **Add upstream remote:**

```bash
git remote add upstream https://github.com/fsvxavier/nexs-mcp.git
git fetch upstream
```

4. **Install dependencies:**

```bash
go mod download
```

5. **Verify installation:**

```bash
make build
make test
```

### Building the Project

```bash
# Build the binary
make build

# Run the server
make run

# Build for all platforms
make build-all
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detector
make test-race

# Generate coverage report
make test-coverage
```

For more detailed information, see [Development Setup](docs/development/SETUP.md) and [Testing Guide](docs/development/TESTING.md).

---

## Development Workflow

### Branching Strategy

We follow a simplified Git Flow model:

- `main` - Production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates
- `refactor/*` - Code refactoring
- `test/*` - Test improvements

### Creating a Branch

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/add-new-tool

# Or a bugfix branch
git checkout -b fix/memory-leak-in-cache
```

### Making Changes

1. **Write clean, focused commits** - Each commit should represent a single logical change
2. **Keep commits small** - Easier to review and revert if needed
3. **Test your changes** - Ensure all tests pass before committing
4. **Follow coding standards** - See section below

### Syncing with Upstream

Regularly sync your branch with upstream to avoid conflicts:

```bash
git checkout main
git pull upstream main
git checkout feature/your-feature
git rebase main
```

### Before Submitting

Run the pre-submission checklist:

```bash
# Format code
make fmt

# Run linter
make lint

# Run all tests
make test-race

# Check coverage
make test-coverage

# Build successfully
make build
```

---

## Coding Standards

NEXS MCP follows standard Go conventions with additional project-specific guidelines.

### Go Style Guide

We follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) with these highlights:

#### Project Structure

```
nexs-mcp/
├── cmd/               # Application entry points
│   └── nexs-mcp/     # Main server application
├── internal/         # Private application code
│   ├── domain/       # Business logic (pure Go, no dependencies)
│   ├── application/  # Use cases and orchestration
│   ├── infrastructure/ # External concerns (storage, MCP, etc.)
│   └── mcp/          # MCP server implementation
├── pkg/              # Public libraries (if any)
├── docs/             # Documentation
├── examples/         # Usage examples
└── test/             # Integration tests
```

#### Naming Conventions

**Packages:**

```go
// ✅ Good - lowercase, singular, descriptive
package domain
package mcp
package indexing

// ❌ Bad - uppercase, plural, underscores
package Domain
package mcpHandlers
package memory_management
```

**Interfaces:**

```go
// ✅ Good - noun or verb+er
type Repository interface { ... }
type Validator interface { ... }
type ElementCreator interface { ... }

// ❌ Bad - suffixed with "Interface"
type RepositoryInterface interface { ... }
```

**Functions and Methods:**

```go
// ✅ Good - verb-based, clear intent
func CreatePersona(id string, data PersonaData) error
func ValidateElement(elem Element) error
func (s *Server) RegisterTools() error

// ❌ Bad - ambiguous, abbreviated
func New(id string, data PersonaData) error
func Validate(elem Element) error
func (s *Server) RegTools() error
```

**Variables:**

```go
// ✅ Good - descriptive, camelCase
var defaultTimeout = 30 * time.Second
var ensembleResults []EnsembleResult
var personaCache map[string]*Persona

// ❌ Bad - unclear, abbreviated
var dt = 30 * time.Second
var res []EnsembleResult
var pc map[string]*Persona
```

**Constants:**

```go
// ✅ Good - PascalCase or camelCase for internal
const (
    DefaultCacheSize = 1000
    MaxRetryAttempts = 3
    minBatchSize = 10  // internal const
)

// ❌ Bad - SCREAMING_CASE (not idiomatic in Go)
const (
    DEFAULT_CACHE_SIZE = 1000
    MAX_RETRY_ATTEMPTS = 3
)
```

#### Code Organization

**File Structure:**

```go
// Every file should follow this order:
// 1. Package declaration
// 2. Imports (standard library, external, internal)
// 3. Constants
// 4. Variables
// 5. Type definitions
// 6. Constructors
// 7. Methods (grouped by receiver)
// 8. Functions

package domain

import (
    "context"
    "time"

    "github.com/google/uuid"

    "github.com/fsvxavier/nexs-mcp/internal/validation"
)

const (
    DefaultMaxAge = 24 * time.Hour
)

type Element struct {
    ID        string
    Type      ElementType
    CreatedAt time.Time
}

func NewElement(typ ElementType) (*Element, error) {
    // Constructor logic
}

func (e *Element) Validate() error {
    // Method logic
}
```

#### Error Handling

**Use descriptive errors:**

```go
// ✅ Good - wrapped errors with context
if err := s.repo.Save(element); err != nil {
    return fmt.Errorf("failed to save element %s: %w", element.ID, err)
}

// ✅ Good - custom error types for domain errors
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ❌ Bad - generic errors without context
if err := s.repo.Save(element); err != nil {
    return err
}
```

**Check errors immediately:**

```go
// ✅ Good
file, err := os.Open("config.yaml")
if err != nil {
    return err
}
defer file.Close()

// ❌ Bad - deferred error checking
file, err := os.Open("config.yaml")
defer file.Close()
if err != nil {
    return err
}
```

#### Comments and Documentation

**Package documentation:**

```go
// Package domain implements the core business logic for NEXS MCP.
// It provides six types of AI elements: Personas, Skills, Templates,
// Agents, Memories, and Ensembles.
//
// Domain models are pure Go with no external dependencies, ensuring
// business logic remains isolated from infrastructure concerns.
package domain
```

**Function documentation:**

```go
// CreatePersona creates a new Persona element with the provided configuration.
// It validates the input, generates a unique ID, and persists the persona.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - name: Unique identifier for the persona
//   - profile: The persona's characteristics and behaviors
//
// Returns the created Persona or an error if validation fails or persistence fails.
func CreatePersona(ctx context.Context, name string, profile Profile) (*Persona, error) {
    // Implementation
}
```

**Inline comments:**

```go
// ✅ Good - explain "why", not "what"
// Use exponential backoff to avoid overwhelming the API during transient failures
time.Sleep(backoffDuration)

// ❌ Bad - stating the obvious
// Sleep for backoff duration
time.Sleep(backoffDuration)
```

#### MCP SDK Usage

**Always use the official MCP Go SDK:**

```go
import (
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ✅ Good - Use SDK types and methods
server := mcp.NewServer(serverInfo)
server.RegisterTool(toolDefinition, handler)

// Register tools with proper error handling
if err := s.server.RegisterTool(toolDef, handler); err != nil {
    return fmt.Errorf("failed to register tool %s: %w", toolDef.Name, err)
}
```

**Tool Registration Pattern:**

```go
// Follow this pattern for registering MCP tools
func (s *Server) registerElementTools() error {
    tools := []struct {
        definition mcp.Tool
        handler    mcp.ToolHandler
    }{
        {
            definition: mcp.Tool{
                Name:        "create_persona",
                Description: "Create a new persona element",
                InputSchema: createPersonaSchema,
            },
            handler: s.handleCreatePersona,
        },
        // More tools...
    }

    for _, tool := range tools {
        if err := s.server.RegisterTool(tool.definition, tool.handler); err != nil {
            return fmt.Errorf("failed to register %s: %w", tool.definition.Name, err)
        }
    }
    return nil
}
```

#### Testing Conventions

**Table-driven tests:**

```go
func TestValidateElement(t *testing.T) {
    tests := []struct {
        name    string
        element Element
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid element",
            element: Element{
                ID:   "test-1",
                Type: PersonaType,
                Name: "Assistant",
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            element: Element{
                Type: PersonaType,
                Name: "Assistant",
            },
            wantErr: true,
            errMsg:  "ID is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.element.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Mock interfaces:**

```go
// Use testify/mock for complex mocking
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, elem Element) error {
    args := m.Called(ctx, elem)
    return args.Error(0)
}

// In tests
mockRepo := new(MockRepository)
mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
```

---

## Commit Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/) for clear, structured commit history.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, missing semicolons, etc.)
- `refactor` - Code refactoring without functional changes
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks (dependencies, build config, etc.)
- `ci` - CI/CD pipeline changes

### Scopes

Common scopes in NEXS MCP:

- `domain` - Domain layer changes
- `application` - Application layer changes
- `infrastructure` - Infrastructure layer changes
- `mcp` - MCP server implementation
- `collection` - Collection management
- `indexing` - Indexing and search
- `api` - API changes
- `cli` - CLI changes
- `docs` - Documentation
- `test` - Test infrastructure

### Examples

**Feature:**
```
feat(mcp): add portfolio management tools

Implement three new MCP tools for portfolio management:
- create_portfolio: Create and manage portfolios
- list_portfolios: List all available portfolios
- update_portfolio: Update portfolio configuration

Uses official MCP Go SDK for tool registration.

Closes #42
```

**Bug fix:**
```
fix(domain): prevent duplicate IDs in ensemble members

Element IDs were not being validated for uniqueness when adding
members to ensembles, causing runtime panics during aggregation.

Added validation in AddMember method with comprehensive tests.

Fixes #156
```

**Documentation:**
```
docs(api): update tool documentation with examples

- Add practical examples for each tool
- Include error handling patterns
- Document rate limits and constraints
- Clarify MCP SDK version requirements
```

**Refactoring:**
```
refactor(infrastructure): extract storage interface

Move file-based storage logic into a separate interface to enable
alternative implementations (in-memory, database, etc.).

No functional changes, all tests pass.
```

**Breaking change:**
```
feat(api)!: change element ID generation to UUIDv7

BREAKING CHANGE: Element IDs now use UUIDv7 instead of random strings.
Existing elements need to be migrated.

Migration guide added in docs/migration/v2.md
```

### Commit Guidelines

1. **Write in imperative mood** - "Add feature" not "Added feature"
2. **Keep subject under 72 characters**
3. **Separate subject from body** with a blank line
4. **Wrap body at 72 characters**
5. **Explain what and why**, not how
6. **Reference issues and PRs** in footer

---

## Pull Request Process

### Before Creating a PR

1. ✅ **Sync with upstream main**
2. ✅ **All tests pass** (`make test-race`)
3. ✅ **Code is formatted** (`make fmt`)
4. ✅ **No linter warnings** (`make lint`)
5. ✅ **Coverage maintained or improved**
6. ✅ **Documentation updated**
7. ✅ **Commit messages follow conventions**

### Creating a Pull Request

1. **Push your branch** to your fork:

```bash
git push origin feature/your-feature
```

2. **Create PR** on GitHub from your fork

3. **Fill out the PR template:**

```markdown
## Description
Brief description of the changes.

## Motivation and Context
Why is this change required? What problem does it solve?
Related issue: #123

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests you ran and their results.

## Checklist
- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have updated the documentation accordingly
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## MCP SDK Version
- [ ] Uses official MCP Go SDK v1.1.0+
- [ ] Follows MCP specification requirements
```

### PR Review Process

1. **Automated checks run** - CI pipeline must pass
2. **Code review** - At least one maintainer reviews
3. **Changes requested** - Address feedback and push updates
4. **Approval** - Once approved, PR can be merged
5. **Merge** - Maintainer merges using squash or rebase

### Review Guidelines

**For Authors:**

- Respond to all comments
- Update PR based on feedback
- Keep PR focused and small (< 400 lines preferred)
- Don't force-push after review starts (breaks review context)

**For Reviewers:**

- Be respectful and constructive
- Explain the "why" behind suggestions
- Distinguish between blocking issues and nitpicks
- Approve when ready, even if minor nitpicks remain
- Focus on:
  - Correctness
  - Performance
  - Security
  - Maintainability
  - Test coverage

### Merge Criteria

A PR can be merged when:

- ✅ All CI checks pass
- ✅ At least one approval from a maintainer
- ✅ No unresolved conversations
- ✅ Branch is up to date with main
- ✅ Coverage does not decrease significantly

---

## Testing Requirements

### Coverage Requirements

- **Overall coverage:** 80%+ target
- **New code:** 90%+ coverage for new features
- **Domain layer:** 95%+ coverage (critical business logic)
- **Application layer:** 85%+ coverage
- **Infrastructure layer:** 70%+ coverage (more integration tests)

### Test Types

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test component interactions
3. **MCP Tests** - Test MCP tool registration and execution
4. **Performance Tests** - For critical paths

### Writing Tests

**Every PR should include:**

- Unit tests for new functions/methods
- Table-driven tests for multiple scenarios
- Error case coverage
- Edge case coverage

**Example:**

```go
func TestCreatePersona(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := NewPersonaService(mockRepo)
        persona, err := svc.Create(ctx, "assistant", validProfile)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, persona.ID)
        assert.Equal(t, "assistant", persona.Name)
    })

    t.Run("validation error", func(t *testing.T) {
        svc := NewPersonaService(mockRepo)
        _, err := svc.Create(ctx, "", invalidProfile)
        
        assert.Error(t, err)
        assert.IsType(t, &ValidationError{}, err)
    })

    t.Run("repository error", func(t *testing.T) {
        mockRepo := new(MockRepository)
        mockRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))
        
        svc := NewPersonaService(mockRepo)
        _, err := svc.Create(ctx, "assistant", validProfile)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "db error")
    })
}
```

For detailed testing guidelines, see [Testing Guide](docs/development/TESTING.md).

---

## Documentation Requirements

### When to Update Documentation

Documentation should be updated when:

- Adding new features or tools
- Changing existing APIs
- Fixing bugs that affect documented behavior
- Adding configuration options
- Changing installation or setup procedures

### Documentation Types

1. **Code Comments** - Document exported functions, types, and packages
2. **API Documentation** - Tool definitions, parameters, responses
3. **User Guides** - How-to guides and tutorials
4. **Architecture Documentation** - Design decisions and architecture
5. **CHANGELOG** - Track changes between versions

### Documentation Standards

**API Documentation:**

```go
// CreatePersona creates a new persona element with the specified configuration.
//
// A persona defines an AI agent's character, behavior patterns, and communication style.
// Each persona must have a unique name within the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - name: Unique identifier for the persona (alphanumeric, underscores, hyphens)
//   - profile: Persona configuration including traits, behaviors, and constraints
//
// Returns:
//   - *Persona: The created persona with generated ID and metadata
//   - error: ValidationError if input is invalid, or persistence error
//
// Example:
//   profile := PersonaProfile{
//       Traits:     []string{"helpful", "concise"},
//       Tone:       "professional",
//       Expertise:  []string{"golang", "architecture"},
//   }
//   persona, err := svc.CreatePersona(ctx, "code-reviewer", profile)
//
// MCP Tool: create_persona
func (s *Service) CreatePersona(ctx context.Context, name string, profile PersonaProfile) (*Persona, error) {
    // Implementation
}
```

**Markdown Documentation:**

- Use clear headings and structure
- Include code examples
- Add table of contents for long documents
- Use tables for reference material
- Include diagrams where helpful

---

## Community and Communication

### Where to Get Help

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - General questions and discussions
- **Documentation** - Check docs/ directory first
- **Examples** - See examples/ directory for usage patterns

### Asking Questions

When asking questions:

1. **Search first** - Check existing issues and discussions
2. **Be specific** - Provide context and details
3. **Include versions** - Go version, NEXS MCP version, OS
4. **Share code** - Minimal reproducible examples
5. **Show effort** - What have you tried?

### Reporting Security Issues

**Do not report security vulnerabilities through public GitHub issues.**

Instead, please email security concerns to the maintainers. Include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

---

## Additional Resources

### Documentation

- [Architecture Overview](docs/architecture/OVERVIEW.md)
- [Development Setup](docs/development/SETUP.md)
- [Testing Guide](docs/development/TESTING.md)
- [Release Process](docs/development/RELEASE.md)
- [MCP Tools API](docs/api/MCP_TOOLS.md)
- [MCP Resources API](docs/api/MCP_RESOURCES.md)

### External Resources

- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Effective Go](https://golang.org/doc/effective_go)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

## Recognition

Contributors are recognized in:

- **CHANGELOG.md** - For each release
- **GitHub Contributors Page** - Automatic
- **README.md** - Top contributors section

---

**Thank you for contributing to NEXS MCP! Your efforts help make AI system management better for everyone.**

For questions about contributing, please open a discussion on GitHub.
