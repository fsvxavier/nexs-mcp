# NEXS-MCP Documentation

**Version:** v1.4.0
**Last Updated:** January 4, 2026
**Status:** Production Ready - Sprint 18 Complete

Welcome to the NEXS-MCP documentation! This guide will help you find the information you need.

## 📊 Project Overview

**NEXS-MCP** is a production-ready Model Context Protocol (MCP) server for managing AI workflows with advanced memory consolidation, semantic search, and multi-agent orchestration.

### Key Statistics
- **💪 94 MCP Tools** across 13 categories
- **🏗️ 21 Application Services** (4 new in Sprint 14)
- **✅ 295 Tests** (100% passing, 0 race conditions)
- **📈 76.4% Test Coverage** (application layer)
- **📦 82,075 Lines** (40,240 production + 41,835 tests)
- **🚀 6 Element Types** (Persona, Skill, Memory, Template, Agent, Ensemble)
- **🔍 4 Embedding Providers** (OpenAI, Transformers, ONNX, Sentence-Transformers)

## 📚 Documentation Structure

### 🚀 User Documentation

Perfect for getting started and learning how to use NEXS-MCP.

- **[Getting Started](./user-guide/GETTING_STARTED.md)** ⭐ **Start here!**
  - What is NEXS-MCP?
  - Installation (5 methods)
  - First run and Claude Desktop integration
  - Creating your first elements
  - Understanding element types
  - Common workflows

- **[Quick Start Guide](./user-guide/QUICK_START.md)**
  - 10 hands-on tutorials (2-5 min each)
  - Create personas, skills, templates
  - Search and filter
  - Backup/restore workflows
  - Memory management
  - GitHub integration
  - Collections
  - Ensembles
  - Analytics

- **[Troubleshooting](./user-guide/TROUBLESHOOTING.md)**
  - Installation issues
  - Connection problems
  - Element operations
  - GitHub integration
  - Performance issues
  - Storage and data
  - FAQ (15+ questions)
  - Error code reference
  - Emergency recovery

### 📖 Element Types

Deep dive into each of the 6 element types NEXS-MCP supports.

- **[Elements Overview](./elements/README.md)**
  - Introduction to all element types
  - When to use each type
  - Element lifecycle

- **[Personas](./elements/PERSONA.md)**
  - Behavioral characteristics
  - Expertise areas
  - Communication styles
  - Use cases and examples

- **[Skills](./elements/SKILL.md)**
  - Reusable capabilities
  - Triggers and procedures
  - Dependencies
  - Categories

- **[Templates](./elements/TEMPLATE.md)**
  - Content patterns
  - Variable substitution
  - Handlebars syntax
  - Template helpers (20+)

- **[Agents](./elements/AGENT.md)**
  - Autonomous execution
  - Goals and actions
  - Decision trees
  - Max iterations

- **[Memories](./elements/MEMORY.md)**
  - Context preservation
  - Memory types (episodic, semantic, procedural)
  - Content hashing
  - Search and retrieval
  - **New:** Advanced consolidation features (Sprint 14)

- **[Ensembles](./elements/ENSEMBLE.md)**
  - Multi-agent orchestration
  - Execution modes
  - Aggregation strategies
  - Voting and consensus

### ⏱️ Infrastructure Features

- **[Token Optimization System](./analysis/TOKEN_OPTIMIZATION_GAPS.md)** ⭐ **New in v1.3.0**
  - 8 integrated optimization services
  - 81-95% token reduction across all operations
  - Prompt compression (35% reduction)
  - Streaming handler (chunked delivery)
  - Semantic deduplication (92%+ similarity)
  - Automatic summarization (70% compression)
  - Context window manager (smart truncation)
  - Adaptive cache (L1/L2 with 1h-7d TTL)
  - Batch processing (10x throughput)
  - Response compression (70-75% reduction)
  - Configuration guide and monitoring

- **[Memory Consolidation](./architecture/APPLICATION.md#memory-consolidation)** ⭐ **New in Sprint 14**
  - HNSW-based duplicate detection
  - DBSCAN + K-means clustering
  - Knowledge graph extraction (NLP entities & relationships)
  - Hybrid search (HNSW + linear fallback)
  - Quality-based retention policies
  - 10 MCP tools for consolidation workflows

- **[Working Memory System](./api/WORKING_MEMORY_TOOLS.md)** ⭐ **New in v1.3.0**
  - Context-aware conversation tracking
  - Conversation lifecycle management
  - Working memory operations (15 tools)
  - Semantic search across conversations
  - Integration with token optimization

- **[Background Task Scheduler](./api/TASK_SCHEDULER.md)** ✨ **New in v1.2.0**
  - Cron-like scheduling (wildcards, ranges, steps, lists)
  - Priority-based execution (Low/Medium/High)
  - Task dependencies with validation
  - Persistent storage with atomic writes
  - Auto-retry mechanisms
  - Examples: cleanup, decay recalculation, backup tasks

- **[Temporal Features & Time Travel](./api/TEMPORAL_FEATURES.md)** ✨ **New in v1.2.0**
  - Version history tracking
  - Confidence decay algorithms
  - Time travel queries
  - Critical relationship preservation
  - [User Guide: Time Travel](./user-guide/TIME_TRAVEL.md)
  - Multi-agent orchestration
  - Execution modes (sequential, parallel, hybrid)
  - Aggregation strategies
  - Use cases

### 🏗️ Architecture & Design

Understand how NEXS-MCP is built.

- **[Architecture Overview](./architecture/OVERVIEW.md)** ⭐ **Essential reading!**
  - Clean architecture principles
  - MCP SDK integration
  - System components
  - Data flow and layer responsibilities
  - Performance characteristics

- **[Domain Layer](./architecture/DOMAIN.md)**
  - Business logic and entities
  - Element types and validation
  - Repository patterns
  - Domain events

- **[Application Layer](./architecture/APPLICATION.md)**
  - Use cases and orchestration
  - Ensemble execution
  - Portfolio management
  - Statistics and monitoring

- **[Infrastructure Layer](./architecture/INFRASTRUCTURE.md)**
  - Storage implementations
  - File-based repository
  - Caching strategies
  - External integrations

- **[MCP Implementation](./architecture/MCP.md)**
  - Official MCP Go SDK usage
  - Tool registration patterns
  - Resource management
  - Protocol compliance

#### Architecture Decision Records (ADRs)

- **[ADR-001: Hybrid Collection Architecture](./adr/ADR-001-hybrid-collection-architecture.md)**
- **[ADR-007: MCP Resources Implementation](./adr/ADR-007-mcp-resources-implementation.md)**
- **[ADR-008: Collection Registry Production](./adr/ADR-008-collection-registry-production.md)**
- **[ADR-009: Element Template System](./adr/ADR-009-element-template-system.md)**
- **[ADR-010: Missing Element Tools](./adr/ADR-010-missing-element-tools.md)**

### 🔧 API & Tools

Reference documentation for developers and power users.

- **[MCP Tools API](./api/MCP_TOOLS.md)** ⭐
  - Complete tool reference (104 tools - updated Sprint 14)
  - Element management tools (26 tools)
  - Memory operations (9 tools)
  - Working memory (15 tools) ⭐ **v1.3.0**
  - **Memory consolidation (10 tools)** ⭐ **New in Sprint 14**
  - Token Optimization (8 tools) ⭐ **v1.3.0**
  - Relationships (5 tools)
  - Temporal/Versioning (4 tools) ✨ **v1.2.0**
  - Quality scoring (3 tools)
  - GitHub integration (11 tools)
  - Search & discovery (7 tools)
  - Ensemble operations (2 tools)
  - Backup and restore (2 tools)
  - Logging & analytics (3 tools)
  - User context (3 tools)
  - Template management (4 tools)

- **[MCP Resources API](./api/MCP_RESOURCES.md)**
  - Resource URIs and schemas
  - Element resources
  - Portfolio resources
  - Collection resources
  - Content formats

- **[CLI Reference](./api/CLI.md)**
  - Command-line interface
  - Usage examples
  - Configuration options

- **[VSCode Settings Reference](./VSCODE_SETTINGS_REFERENCE.md)** ⭐ **New!**
  - Complete configuration guide
  - All environment variables documented
  - Production-ready settings
  - Development configurations
  - Future features (ONNX/Vector Search)
  - Troubleshooting guide

- **[MCP Resources (Legacy)](./mcp/RESOURCES.md)**
  - Capability index
  - Resource URIs
  - Content formats

### 👥 Contributing & Development

Guides for contributors and developers.

- **[Contributing Guide](../CONTRIBUTING.md)** ⭐ **Start here to contribute!**
  - Code of conduct
  - How to contribute
  - Coding standards
  - Commit conventions
  - PR process
  - Testing requirements

- **[Development Setup](./development/SETUP.md)**
  - Prerequisites and installation
  - Building from source
  - Running locally
  - IDE setup (VS Code, GoLand)
  - Debug configuration
  - Common issues

- **[Testing Guide](./development/TESTING.md)**
  - Test organization
  - Unit testing
  - Integration testing
  - MCP protocol testing
  - Writing tests
  - Coverage requirements
  - Mocking strategies

- **[Release Process](./development/RELEASE.md)**
  - Version bumping
  - Changelog generation
  - Creating releases
  - Publishing to registries
  - Release checklist
  - Rollback procedures

- **[Template System](./templates/TEMPLATES.md)**
  - Built-in templates
  - Creating custom templates
  - Template helpers

### 🚢 Deployment

Guides for deploying NEXS-MCP in different environments.

- **[Docker Deployment](./deployment/DOCKER.md)** ⭐
  - Quick start with Docker
  - Docker Compose setup
  - Volume management
  - Environment variables
  - Production deployment (Swarm, Kubernetes)
  - Security best practices
  - Troubleshooting

### 📦 Collections

Learn about the NEXS-MCP collection ecosystem.

- **[Collection Publishing Guide](./collections/PUBLISHING.md)**
  - Creating collections
  - Manifest format
  - Publishing to registry

- **[Collection Registry](./collections/REGISTRY.md)**
  - Finding collections
  - Installing collections
  - Managing installed collections

- **[Security Guidelines](./collections/SECURITY.md)**
  - Collection validation
  - Security scanning
  - Best practices

### 🔍 Indexing & Search

Advanced search and discovery features.

- **[Enhanced Index](./indexing/ENHANCED_INDEX.md)**
  - TF-IDF indexing
  - Semantic search
  - Similarity detection

- **[M0.10 Summary](./indexing/M0.10_SUMMARY.md)**
  - Indexing implementation details
  - Performance characteristics

### 📋 Planning & Roadmap

Project planning and development tracking.

- **[Next Steps](../NEXT_STEPS.md)** - Current priorities and roadmap
- **[README (Next Steps)](./next_steps/01_README.md)** - Planning overview
- **[Immediate Next Steps](./next_steps/02_IMMEDIATE_NEXT_STEPS.md)**
- **[Roadmap](./next_steps/03_ROADMAP.md)**
- **[Milestones](./next_steps/04_MILESTONES.md)**
- **[Backlog](./next_steps/05_BACKLOG.md)**
- **[Risks and Mitigations](./next_steps/06_RISKS_AND_MITIGATIONS.md)**
- **[Metrics and KPIs](./next_steps/07_METRICS_AND_KPIS.md)**

### 🏗️ Project Planning (Legacy)

Historical planning documents.

- **[README](./plano/01_README.md)**
- **[Executive Summary](./plano/02_EXECUTIVE_SUMMARY.md)**
- **[Architecture](./plano/03_ARCHITECTURE.md)**
- **[Tools Spec](./plano/04_TOOLS_SPEC.md)**
- **[Testing Plan](./plano/05_TESTING_PLAN.md)**

---

## 🎯 Quick Navigation

### I want to...

**...get started quickly**
→ [Getting Started](./user-guide/GETTING_STARTED.md) → [Quick Start](./user-guide/QUICK_START.md)

**...integrate with Claude Desktop**
→ [Getting Started: Claude Desktop Integration](./user-guide/GETTING_STARTED.md#claude-desktop-integration)

**...understand element types**
→ [Elements Overview](./elements/README.md)

**...create my first persona**
→ [Quick Start: Tutorial 1](./user-guide/QUICK_START.md#tutorial-1-create-your-first-persona-2-minutes)

**...use templates**
→ [Quick Start: Tutorial 3](./user-guide/QUICK_START.md#tutorial-3-generate-documents-with-templates-3-minutes)

**...backup my portfolio**
→ [Quick Start: Tutorial 5](./user-guide/QUICK_START.md#tutorial-5-backup-your-portfolio-1-minute)

**...sync with GitHub**
→ [Quick Start: Tutorial 7](./user-guide/QUICK_START.md#tutorial-7-github-integration-5-minutes)

**...install collections**
→ [Quick Start: Tutorial 8](./user-guide/QUICK_START.md#tutorial-8-collections-4-minutes)

**...deploy with Docker**
→ [Docker Deployment Guide](./deployment/DOCKER.md)

**...troubleshoot issues**
→ [Troubleshooting Guide](./user-guide/TROUBLESHOOTING.md)

**...find API reference**
→ Coming soon! (check [MCP Resources](./mcp/RESOURCES.md) for now)

**...contribute**
→ [Contributing Guide](../CONTRIBUTING.md) → [Development Setup](./development/SETUP.md)

**...set up development environment**
→ [Development Setup](./development/SETUP.md)

**...write tests**
→ [Testing Guide](./development/TESTING.md)

**...create a release**
→ [Release Process](./development/RELEASE.md)

**...understand the architecture**
→ [Architecture Overview](./architecture/OVERVIEW.md)

---

## 📊 Documentation Stats

- **User Guides:** 3 documents (2,000+ lines)
- **Element Docs:** 7 documents
- **Architecture Docs:** 5 comprehensive guides (5,000+ lines)
- **API Reference:** 3 complete references (4,000+ lines)
- **Contributing Guides:** 4 detailed guides (2,800+ lines)
- **ADRs:** 5 architecture decisions
- **Deployment Guides:** 1 comprehensive guide (600+ lines)
- **Total:** 70+ documentation files (15,000+ lines)

---

## 🆘 Need Help?

1. **Check the docs:** Use the navigation above
2. **Search:** Use Ctrl+F / Cmd+F to search this page
3. **Troubleshooting:** [Common issues and solutions](./user-guide/TROUBLESHOOTING.md)
4. **FAQ:** [Frequently asked questions](./user-guide/TROUBLESHOOTING.md#faq)
5. **Community:** [GitHub Discussions](https://github.com/fsvxavier/nexs-mcp/discussions)
6. **Bugs:** [GitHub Issues](https://github.com/fsvxavier/nexs-mcp/issues)

---

## 🤝 Contributing

We welcome contributions to documentation! If you find errors, have suggestions, or want to add examples:

1. Fork the repository
2. Make your changes
3. Submit a pull request
4. Include clear descriptions of your changes

Areas we'd love help with:
- More hands-on tutorials
- Real-world use case examples
- Translations
- Video tutorials (link from docs)
- API reference documentation

---

## 📝 Documentation Standards

When contributing:

- Use clear, concise language
- Include code examples
- Add expected outputs where relevant
- Link to related documentation
- Keep formatting consistent
- Test all commands and examples

---

**Last Updated:** December 24, 2025
**Version:** 1.3.0
**Status:** ✅ Comprehensive user documentation complete
