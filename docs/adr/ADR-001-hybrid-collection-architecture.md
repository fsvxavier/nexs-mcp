# ADR-001: Hybrid Collection Architecture (GitHub + Local)

**Status:** Accepted  
**Date:** 2025-12-18  
**Authors:** NEXS MCP Team  
**Decision Owners:** Architecture Team

---

## Context

NEXS MCP requires a system for discovering, installing, and managing collections of elements (Personas, Skills, Templates, Agents, Memories, Ensembles). Collections enable users to:

- Share curated sets of related elements
- Install pre-built element packages
- Discover community-contributed content
- Manage dependencies between collections
- Version and update element sets

The original DollhouseMCP project uses a centralized registry at `dollhousemcp.com/collections`. We need to decide on the collection distribution architecture for NEXS MCP.

### Requirements

**Functional:**
- FR1: Users must be able to discover collections
- FR2: Users must be able to install collections from multiple sources
- FR3: Collections must support versioning and updates
- FR4: Collections must be shareable (export/import)
- FR5: Dependency resolution between collections
- FR6: Offline functionality for local development

**Non-Functional:**
- NFR1: No dependency on external centralized services
- NFR2: Leverage existing infrastructure (GitHub OAuth already implemented)
- NFR3: Support both online and offline workflows
- NFR4: Extensible to add new sources in the future
- NFR5: Simple for users to publish their own collections
- NFR6: Free to use (no hosting costs for registry)

### Constraints

- GitHub Integration (OAuth2 + API client) already implemented in M0.3
- File-based repository pattern already established
- Must align with Clean Architecture principles
- Go 1.25 standard library capabilities
- MCP protocol limitations (stdio transport, no long-running connections)

---

## Decision

We will implement a **Hybrid Collection Architecture** that supports multiple collection sources without dependency on centralized services.

### Primary Sources

1. **GitHub Collections** (Primary - recommended)
   - Collections are GitHub repositories with standardized structure
   - Discovery via GitHub Search API with topic tags (`nexs-collection`)
   - Versioning via Git tags (semver)
   - Installation via existing GitHubClient (OAuth2 already implemented)
   - Free hosting and bandwidth via GitHub
   - Built-in code review, issues, and community features

2. **Local Collections** (Secondary - offline support)
   - Collections as local directories with `collection.yaml` manifest
   - Discovery via filesystem scanning
   - Import/export as `.tar.gz` or `.zip` archives
   - Completely offline workflow
   - Full privacy (no external dependencies)

3. **Extensible Source Interface** (Future-proof)
   - `CollectionSource` interface allows adding new sources
   - Potential future sources: HTTP registries, S3 buckets, IPFS, etc.
   - Configuration via `~/.nexs-mcp/sources.yaml`

### URI Scheme

Collections are referenced via URIs:

```
github://owner/repo[@version]          # GitHub repository
file:///path/to/collection              # Local directory
https://example.com/collection.tar.gz  # HTTP download
```

### Architecture Components

```
internal/collection/
  ├── manifest.go          # Collection manifest schema (YAML)
  ├── registry.go          # Multi-source registry coordinator
  ├── sources/
  │   ├── interface.go     # CollectionSource interface
  │   ├── github.go        # GitHub implementation (reuses GitHubClient)
  │   ├── local.go         # Local filesystem implementation
  │   └── http.go          # HTTP/HTTPS download (future)
  ├── installer.go         # Installation workflow (atomic operations)
  ├── validator.go         # Manifest and element validation
  ├── dependency.go        # Dependency resolution
  └── manager.go           # Update/export/publish operations
```

### Manifest Format

Collections use a `collection.yaml` manifest:

```yaml
name: "DevOps Toolkit"
version: "2.1.0"
author: "username"
description: "Collection description"
tags: ["devops", "kubernetes"]
dependencies:
  - "github://nexs-mcp/base-skills@^1.2.0"
elements:
  - personas/*.yaml
  - skills/*.yaml
```

See `docs/collection-manifest-example.yaml` for complete example.

---

## Consequences

### Positive

✅ **Decentralized Architecture**
- No single point of failure
- No hosting costs or infrastructure maintenance
- Community can self-host and share freely
- Resistant to service shutdowns

✅ **Leverage Existing Infrastructure**
- Reuses GitHub OAuth2 implementation (M0.3)
- Reuses GitHubClient and sync logic
- Familiar Git-based workflow for developers
- GitHub provides free hosting, CDN, and bandwidth

✅ **Offline Support**
- Local collections work completely offline
- No network dependency for local development
- Private collections stay private (no publishing required)

✅ **Extensibility**
- `CollectionSource` interface allows new sources
- Can add HTTP registries, cloud storage, IPFS later
- Not locked into any single technology

✅ **Simplicity for Publishers**
- Anyone can publish: create GitHub repo + add topic tag
- No registration, approval, or accounts required
- Standard Git workflow (branches, PRs, tags)
- Free code review and issue tracking

✅ **Familiar Tooling**
- Git tags for versioning (standard semver)
- GitHub Releases for changelogs
- GitHub Topics for categorization
- GitHub Search for discovery

### Negative

❌ **Discovery Limitations**
- GitHub Search API has rate limits (60 req/hour unauthenticated)
- No centralized popularity metrics (stars/downloads proxy)
- Requires GitHub account for private collections
- Search relevance depends on GitHub's algorithm

❌ **Version Management Complexity**
- Must parse Git tags for version resolution
- No guaranteed semver compliance (trust authors)
- Dependency resolution more complex than centralized registry
- Potential for circular dependencies (must detect)

❌ **Initial Setup Friction**
- Users need GitHub authentication for GitHub collections
- Publishers need basic Git knowledge
- No curated "official" collection list by default

❌ **Quality Control**
- No centralized review or validation
- Malicious collections possible (same as npm, PyPI)
- Users must trust collection authors
- No built-in security scanning (future: GitHub security features)

### Neutral

⚖️ **GitHub Dependency**
- Pro: Already implemented, widely used, reliable
- Con: Creates dependency on GitHub as a platform
- Mitigation: Local collections provide alternative

⚖️ **Manifest Complexity**
- Pro: Comprehensive features (hooks, dependencies, metadata)
- Con: More complex than simple file list
- Mitigation: Start simple, optional advanced features

---

## Alternatives Considered

### Alternative 1: Centralized Registry (dollhousemcp.com approach)

**Pros:**
- Single source of truth
- Curated, reviewed collections
- Centralized analytics and metrics
- Consistent discovery experience

**Cons:**
- Requires hosting infrastructure ($$$)
- Maintenance burden (servers, database, API)
- Single point of failure
- Gatekeeping (approval process)
- Risk of shutdown/abandonment
- **Rejected**: Goes against NFR1 (no centralized dependencies)

### Alternative 2: GitHub-Only (No Local Support)

**Pros:**
- Simpler implementation (one source)
- Consistent experience
- Leverage full GitHub feature set

**Cons:**
- No offline support (violates NFR3)
- Requires GitHub account for all users
- No privacy for local-only collections
- **Rejected**: Violates NFR3 (offline functionality)

### Alternative 3: Local-Only (No GitHub Integration)

**Pros:**
- Complete privacy and offline support
- No external dependencies
- Simple implementation

**Cons:**
- Difficult discovery and sharing
- Manual distribution (email, sneakernet)
- No version management
- Doesn't leverage M0.3 GitHub integration
- **Rejected**: Violates NFR2 (leverage existing infra)

### Alternative 4: NPM/PyPI-style Registry

**Pros:**
- Proven model (npm, PyPI, RubyGems)
- Centralized package management
- Standardized tooling

**Cons:**
- Requires hosting and maintenance
- Overkill for MCP use case
- Hosting costs
- **Rejected**: Violates NFR1 and NFR6

### Alternative 5: IPFS/Decentralized Storage

**Pros:**
- Truly decentralized
- Censorship-resistant
- No hosting costs

**Cons:**
- Complex setup for users
- Immature tooling (Go IPFS libraries)
- Slow/unreliable for large collections
- Poor discovery mechanisms
- **Rejected**: Too experimental, poor UX

---

## Implementation Plan

### Phase 1: Core Infrastructure (M0.4 - Week 1)
- [ ] `collection.yaml` manifest schema
- [ ] `CollectionSource` interface
- [ ] Multi-source registry coordinator
- [ ] Manifest parser and validator

### Phase 2: GitHub Source (M0.4 - Week 1)
- [ ] GitHub Topics API integration
- [ ] Git clone/checkout via existing GitHubClient
- [ ] Version resolution from Git tags
- [ ] Tests with mock GitHub API

### Phase 3: Local Source (M0.4 - Week 1)
- [ ] Filesystem scanner
- [ ] Tar.gz/zip import/export
- [ ] Local version management
- [ ] Tests with temp directories

### Phase 4: Installation & Dependencies (M0.4 - Week 2)
- [ ] Atomic installation workflow
- [ ] Dependency resolution (DAG traversal)
- [ ] Rollback on failure
- [ ] Conflict detection

### Phase 5: Management Tools (M0.4 - Week 2)
- [ ] Update checking (Git fetch)
- [ ] Export collection (create tar.gz)
- [ ] Publish helper (local → GitHub)
- [ ] MCP tools implementation

### Phase 6: Documentation & Examples (M0.4 - Week 2)
- [ ] Collection authoring guide
- [ ] Publishing tutorial
- [ ] Example collections
- [ ] Best practices

---

## Metrics and Success Criteria

### Technical Metrics
- Collection installation time: < 30s for typical collection (50 elements)
- Discovery search time: < 2s for GitHub search
- Local scan time: < 1s for 100 collections
- Dependency resolution: supports 5+ levels deep
- Test coverage: > 90% for collection package

### User Metrics
- Time to publish first collection: < 30 minutes (including docs)
- Published collections in first 3 months: > 20
- Collection installation success rate: > 95%
- User satisfaction: > 4.0/5.0 (survey)

### Community Metrics
- Official example collections: 10+ (devops, data-science, creative, etc.)
- Community collections: 50+ by month 6
- GitHub stars on popular collections: > 100
- Active collection maintainers: > 30

---

## References

- [GitHub Search API](https://docs.github.com/en/rest/search)
- [GitHub Topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics)
- [Semantic Versioning](https://semver.org/)
- [npm package.json spec](https://docs.npmjs.com/cli/v10/configuring-npm/package-json) (inspiration)
- [Cargo.toml manifest](https://doc.rust-lang.org/cargo/reference/manifest.html) (inspiration)
- M0.3 GitHub Integration (`internal/infrastructure/github_*.go`)
- Clean Architecture principles (Uncle Bob)

---

## Changelog

| Date | Author | Change |
|------|--------|--------|
| 2025-12-18 | NEXS MCP Team | Initial ADR creation |
| 2025-12-18 | NEXS MCP Team | Added detailed manifest example |
| 2025-12-18 | NEXS MCP Team | Added implementation plan and metrics |

---

## Approval

- [x] **Architect:** Approved - leverages existing GitHub infrastructure
- [x] **Tech Lead:** Approved - extensible design, no vendor lock-in
- [x] **Product Owner:** Approved - meets user needs, low maintenance burden
- [x] **Security:** Approved - with recommendation to add collection signature verification in future

---

## Related ADRs

- ADR-002: Element Repository Pattern (file-based storage rationale)
- ADR-003: GitHub Integration Architecture (OAuth2 + sync implementation)
- ADR-004: Collection Dependency Resolution (future - detailed dependency algorithm)

---

**Next Steps:**

1. Review and refine `collection.yaml` schema with team
2. Create proof-of-concept GitHub collection
3. Begin M0.4 implementation (estimated 2 weeks)
4. Update NEXT_STEPS.md with M0.4 tasks (✅ complete)
