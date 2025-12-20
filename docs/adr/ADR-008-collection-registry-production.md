# ADR-008: Collection Registry Production Implementation

**Status:** Proposed  
**Date:** 2025-12-19  
**Authors:** NEXS-MCP Development Team  
**Milestone:** M0.8 - Collection Registry Production  
**Story Points:** 13  
**Priority:** P0-Critical

---

## Context

Following M0.7 completion (MCP Resources Protocol), NEXS-MCP has foundational collection support via:
- `internal/collection/sources/` (GitHub, local filesystem sources)
- `internal/collection/registry.go` (multi-source aggregation)
- `internal/collection/manifest.go` (comprehensive manifest structure)
- `internal/collection/validator.go` (basic validation)
- `internal/collection/installer.go` (installation with dependency resolution)

However, the current implementation lacks production-grade features required for secure, reliable collection distribution:

**Current Gaps:**
1. **Validation:** Basic checks only (~20 rules), no comprehensive schema/security validation
2. **Security:** No signature verification, checksum validation, or malicious code detection
3. **Performance:** No caching or indexing in registry (repeated network calls)
4. **Publishing:** Manual process, no automated PR workflow or GitHub integration
5. **Discovery:** Limited search/filtering, no metadata indexing
6. **Testing:** Minimal integration tests for multi-source scenarios

**Business Requirements:**
- Enable safe, community-driven collection publishing (like npm, pypi)
- Prevent malicious collections (code injection, path traversal attacks)
- Ensure consistent quality (comprehensive validation)
- Optimize performance (caching, indexing)
- Streamline contribution (automated PR workflow)

**Technical Constraints:**
- Must use official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`)
- Maintain zero regressions (all 2,331 tests passing)
- Follow Clean Architecture patterns
- Security-first approach (default deny, explicit allow)

---

## Decision

Implement **Production-Grade Collection Registry** with five core systems:

### 1. Enhanced Validation System (`internal/collection/validator.go`)

**100+ Validation Rules** across categories:

**Schema Validation (30 rules):**
- Required fields presence (name, version, author, description, elements)
- Field type correctness (strings, arrays, objects)
- Field format validation (semver, email, URL, URIs)
- Field length constraints (name: 3-64, description: 10-500)
- Enum validation (category, element types, hook types)

**Security Validation (25 rules):**
- Path traversal detection (`..`, absolute paths, symlinks)
- Shell injection prevention (backticks, `$()`, `&&`, `||`, `;`, `|`)
- Malicious patterns (eval, exec, rm -rf, curl | bash)
- File permissions (no setuid, no 777)
- Dependency depth limits (max 10 levels)

**Dependency Validation (15 rules):**
- URI format (github://, file://, https://)
- Version constraint syntax (semver ranges: ^, ~, >, <, =)
- Circular dependency detection
- Dependency existence verification
- Version conflict resolution

**Element Validation (20 rules):**
- Path validation (glob patterns, file existence)
- Element type matching (file extension vs type)
- Duplicate detection (same path)
- Size limits (< 1MB per file, < 10MB total)
- Content validation (valid YAML/JSON/Markdown)

**Hook Validation (10 rules):**
- Hook type validation (command, validate, backup, confirm)
- Command safety (whitelisted commands only)
- Hook execution time limits
- Platform compatibility (Linux/macOS/Windows)

**Implementation:**
```go
// internal/collection/validator.go enhancements
type ValidationError struct {
    Field    string   `json:"field"`
    Rule     string   `json:"rule"`
    Message  string   `json:"message"`
    Severity string   `json:"severity"` // error, warning
    Path     string   `json:"path"`     // JSON path to field
}

type ValidationResult struct {
    Valid    bool                `json:"valid"`
    Errors   []*ValidationError  `json:"errors"`
    Warnings []*ValidationError  `json:"warnings"`
    Stats    map[string]int      `json:"stats"`
}

func (v *Validator) ValidateComprehensive(manifest *Manifest) *ValidationResult {
    // Run all 100+ rules
    // Return structured errors with fix suggestions
}
```

---

### 2. Security Validation System (`internal/collection/security.go`)

**New Package:** `internal/collection/security/`

**Components:**

**A. Signature Verification:**
```go
// security/signature.go
type SignatureVerifier interface {
    Verify(manifestPath string, signaturePath string, publicKey string) error
}

// Support GPG and SSH signatures
type GPGVerifier struct { }
type SSHVerifier struct { }
```

**B. Checksum Validation:**
```go
// security/checksum.go
type ChecksumValidator struct {
    algorithm string // sha256, sha512
}

func (c *ChecksumValidator) Validate(tarballPath string, expectedChecksum string) error {
    // Compute hash, compare
}
```

**C. Trusted Sources:**
```go
// security/sources.go
type TrustedSourceRegistry struct {
    sources map[string]*TrustedSource
}

type TrustedSource struct {
    Name      string
    Pattern   string // github.com/org/*
    PublicKey string
    Required  bool   // if true, unsigned collections rejected
}
```

**D. Malicious Code Detection:**
```go
// security/scanner.go
type CodeScanner struct {
    rules []*ScanRule
}

type ScanRule struct {
    Name        string
    Pattern     *regexp.Regexp
    Severity    string // critical, high, medium, low
    Description string
}

var maliciousPatterns = []ScanRule{
    {Name: "eval-injection", Pattern: regexp.MustCompile(`eval\s*\(`), Severity: "critical"},
    {Name: "remote-exec", Pattern: regexp.MustCompile(`curl.*\|.*bash`), Severity: "critical"},
    // 50+ patterns
}
```

**Configuration:**
```go
// config/config.go additions
type SecurityConfig struct {
    RequireSignatures bool
    TrustedSources    []string
    AllowUnsigned     bool
    ScanEnabled       bool
    ScanThreshold     string // critical, high, medium
}
```

---

### 3. Registry Caching & Indexing (`internal/collection/registry.go`)

**A. In-Memory Cache:**
```go
type CachedCollection struct {
    Collection *sources.Collection
    ExpiresAt  time.Time
    AccessCount int
    LastAccess  time.Time
}

type RegistryCache struct {
    collections map[string]*CachedCollection // URI -> cached
    metadata    map[string][]*sources.CollectionMetadata // source -> metadata
    mu          sync.RWMutex
    ttl         time.Duration // default: 15min
}

func (r *Registry) GetCached(ctx context.Context, uri string) (*sources.Collection, error) {
    // Check cache first, fallback to network
}
```

**B. Metadata Index:**
```go
type MetadataIndex struct {
    byAuthor   map[string][]*sources.CollectionMetadata
    byCategory map[string][]*sources.CollectionMetadata
    byTag      map[string][]*sources.CollectionMetadata
    byKeyword  map[string][]*sources.CollectionMetadata
    full       []*sources.CollectionMetadata
    mu         sync.RWMutex
}

func (i *MetadataIndex) Search(query string, filters *SearchFilters) []*sources.CollectionMetadata {
    // Full-text search + filtering
}
```

**C. Dependency Graph Cache:**
```go
type DependencyGraph struct {
    nodes map[string]*DependencyNode
    mu    sync.RWMutex
}

type DependencyNode struct {
    URI          string
    Dependencies []*DependencyNode
    Dependents   []*DependencyNode
    Depth        int
}

func (g *DependencyGraph) Resolve(uri string) ([]*DependencyNode, error) {
    // Topological sort, cycle detection
}
```

**Performance Targets:**
- Cold browse: ~500ms (network-bound)
- Cached browse: <10ms (95th percentile)
- Collection get (cold): ~1s (network-bound)
- Collection get (cached): <5ms
- Dependency resolution: <100ms

---

### 4. Publishing System (`internal/mcp/publishing_tools.go`)

**New MCP Tool: `publish_collection`**

**Workflow:**
1. **Prepare:** Validate manifest, generate checksums, create tarball
2. **Fork:** Auto-fork target registry repo (e.g., nexs-mcp/collections)
3. **Clone:** Clone fork locally
4. **Commit:** Add collection files, create commit
5. **Push:** Push to fork
6. **PR:** Create GitHub PR with review template
7. **Release:** Create GitHub release with assets (optional)

**Implementation:**
```go
// internal/mcp/publishing_tools.go
func init() {
    publishCollectionTool := mcp.Tool{
        Name: "publish_collection",
        Description: "Publish a collection to NEXS-MCP registry",
        InputSchema: schema{
            "type": "object",
            "properties": {
                "manifest_path": {"type": "string", "description": "Path to collection.yaml"},
                "registry": {"type": "string", "default": "github.com/fsvxavier/nexs-mcp-collections"},
                "create_release": {"type": "boolean", "default": false},
                "dry_run": {"type": "boolean", "default": false},
            },
            "required": ["manifest_path"],
        },
    }
}

func handlePublishCollection(params map[string]interface{}) (*toolResult, error) {
    // 1. Load and validate manifest (100+ rules)
    // 2. Security scan
    // 3. Generate checksums (SHA-256)
    // 4. Create tarball
    // 5. Fork registry repo (via GitHub API)
    // 6. Clone fork
    // 7. Copy files, commit
    // 8. Push to fork
    // 9. Create PR (with template)
    // 10. Create release (if requested)
}
```

**GitHub Integration:**
```go
// internal/infrastructure/github_publisher.go
type GitHubPublisher struct {
    client *github.Client
    token  string
}

func (p *GitHubPublisher) ForkRepository(owner, repo string) (string, error) {
    // Fork via GitHub API
}

func (p *GitHubPublisher) CreatePullRequest(params *PRParams) (*github.PullRequest, error) {
    // Create PR with review template
}

func (p *GitHubPublisher) CreateRelease(params *ReleaseParams) (*github.RepositoryRelease, error) {
    // Create release, upload assets
}
```

**PR Template:**
```markdown
## Collection Submission: {author}/{name}

**Version:** {version}
**Author:** {author}
**Category:** {category}

### Description
{description}

### Checklist
- [ ] Manifest valid (100+ rules passed)
- [ ] Security scan passed
- [ ] All elements tested
- [ ] Dependencies resolved
- [ ] Documentation complete
- [ ] Examples provided

### Stats
- Elements: {total_elements}
- Personas: {personas}
- Skills: {skills}
- Templates: {templates}

### Links
- Repository: {repository}
- Documentation: {documentation}
- Homepage: {homepage}
```

---

### 5. Enhanced Discovery (`internal/mcp/collection_tools.go`)

**Update `list_collections` tool:**

**Current Output:**
```
Available Collections:
- author/name@version
```

**Enhanced Output:**
```markdown
## Available Collections (127 found)

### 🌟 Featured
📦 ai-devops/aws-automation@2.1.0 ⭐ 1.2k
   AWS infrastructure automation with Terraform and CloudFormation
   Tags: aws, devops, terraform
   Source: github.com/ai-devops/collections
   Elements: 15 skills, 8 templates, 3 personas

### 🎨 Creative Writing (Category: creative)
📦 storyteller/narrative-tools@1.5.2 ⭐ 856
   Professional narrative structuring and character development
   Tags: writing, story, creative
   Source: local
   Elements: 12 skills, 5 templates, 2 personas

### Search & Filters
- search_collections(query="aws terraform")
- filter_collections(category="devops", tags=["aws"])
- sort_collections(by="stars|downloads|updated")
```

**New Tool: `search_collections`**
```go
searchCollectionsTool := mcp.Tool{
    Name: "search_collections",
    Description: "Search collections with advanced filtering",
    InputSchema: schema{
        "type": "object",
        "properties": {
            "query": {"type": "string"},
            "category": {"type": "string"},
            "tags": {"type": "array", "items": {"type": "string"}},
            "author": {"type": "string"},
            "min_stars": {"type": "number"},
            "sort_by": {"type": "string", "enum": ["relevance", "stars", "downloads", "updated"]},
            "limit": {"type": "number", "default": 20},
        },
    },
}
```

---

## Architecture

### Component Diagram
```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Server Layer                         │
├─────────────────────────────────────────────────────────────┤
│  publish_collection │ list_collections │ search_collections  │
│  install_collection │ update_collection │ uninstall          │
└───────────────┬──────────────────────┬──────────────────────┘
                │                      │
                v                      v
┌───────────────────────────┐  ┌──────────────────────────────┐
│   Collection Manager      │  │    Collection Registry       │
├───────────────────────────┤  ├──────────────────────────────┤
│ • Install/Update/Remove   │  │ • Multi-source aggregation   │
│ • Dependency resolution   │  │ • In-memory cache (15min)    │
│ • Hook execution          │  │ • Metadata indexing          │
│ • Rollback support        │  │ • Search/filtering           │
└───────────┬───────────────┘  └────┬─────────────────────────┘
            │                       │
            v                       v
┌───────────────────────────────────────────────────────────────┐
│                      Validator Layer                          │
├───────────────────────────────────────────────────────────────┤
│  Schema (30) │ Security (25) │ Dependency (15) │ Element (20) │
└───────────────────────────────────────────────────────────────┘
                                │
                                v
┌───────────────────────────────────────────────────────────────┐
│                     Security Layer                            │
├───────────────────────────────────────────────────────────────┤
│  Signature Verify │ Checksum │ Trusted Sources │ Code Scan    │
└───────────────────────────────────────────────────────────────┘
                                │
                                v
┌───────────────────────────────────────────────────────────────┐
│                   Collection Sources                          │
├───────────────────────────────────────────────────────────────┤
│      GitHub Source   │   Local Source   │   HTTP Source       │
└───────────────────────────────────────────────────────────────┘
```

### Data Flow: Publishing

```
Developer → publish_collection
    ↓
1. Load manifest.yaml
    ↓
2. Validate (100+ rules) ──→ ValidationResult
    ├─ Schema validation
    ├─ Security validation
    ├─ Dependency validation
    └─ Element validation
    ↓
3. Security Scan ──→ ScanResult
    ├─ Code patterns
    ├─ Checksum generation
    └─ Signature creation (optional)
    ↓
4. Package ──→ collection.tar.gz + checksums.txt
    ↓
5. GitHub Workflow
    ├─ Fork registry repo
    ├─ Clone fork
    ├─ Add files (manifest, elements, checksums)
    ├─ Commit with message
    ├─ Push to fork
    └─ Create PR (with template)
    ↓
6. Return PR URL ──→ User
```

### Data Flow: Installation

```
User → install_collection("github://author/name@1.0.0")
    ↓
1. Registry.Get(uri)
    ├─ Check cache (hit: <5ms, miss: network call)
    └─ Source.Get(uri) ──→ Collection
    ↓
2. Validator.ValidateComprehensive(manifest)
    └─ 100+ rules ──→ ValidationResult
    ↓
3. Security.Validate(collection)
    ├─ Signature verification (if required)
    ├─ Checksum validation
    ├─ Code scanning
    └─ Trusted source check
    ↓
4. Installer.ResolveDependencies(manifest)
    ├─ Build dependency graph
    ├─ Detect cycles
    ├─ Topological sort
    └─ Recursive install
    ↓
5. Installer.AtomicInstall(collection)
    ├─ Extract to temp directory
    ├─ Validate elements
    ├─ Execute pre-install hooks
    ├─ Move to install location
    ├─ Execute post-install hooks
    └─ Update installation state
    ↓
6. Return InstallationRecord ──→ User
```

---

## Alternatives Considered

### Alternative 1: External Registry Service

**Approach:** Host centralized registry service (like npmjs.com)

**Pros:**
- Centralized management
- Rich web UI
- Better analytics
- Spam prevention

**Cons:**
- Infrastructure cost
- Single point of failure
- Not aligned with MCP philosophy (local-first)
- Requires backend maintenance

**Decision:** ❌ Rejected - Stick to GitHub-based approach (decentralized, free, Git-native)

---

### Alternative 2: Plugin-Based Validation

**Approach:** Allow custom validation plugins

**Pros:**
- Extensible validation
- Community-driven rules
- Language-specific validators

**Cons:**
- Security risk (malicious validators)
- Complexity overhead
- Performance impact

**Decision:** ❌ Deferred to M1.4 - Start with comprehensive built-in rules (100+), add plugin support later if needed

---

### Alternative 3: Blockchain-Based Registry

**Approach:** Use blockchain for tamper-proof registry

**Pros:**
- Immutable history
- Decentralized trust
- Cryptographic verification

**Cons:**
- Extreme complexity
- Performance issues
- No real benefit over Git + GPG

**Decision:** ❌ Rejected - Overkill, Git already provides sufficient integrity

---

### Alternative 4: No Publishing Tool (Manual PR)

**Approach:** Users manually create PRs to registry repo

**Pros:**
- Simpler implementation
- More control for reviewers

**Cons:**
- High friction for contributors
- Error-prone (manual validation)
- Slow adoption

**Decision:** ❌ Rejected - Automation is critical for community growth

---

## Consequences

### Positive

✅ **Security:** Multi-layered validation prevents malicious collections
✅ **Performance:** Caching reduces network calls by ~95% (15min TTL)
✅ **Quality:** 100+ validation rules ensure consistency
✅ **Adoption:** Automated publishing reduces friction to <5 minutes
✅ **Discoverability:** Enhanced search/filtering, metadata indexing
✅ **Reliability:** Atomic installation with rollback, dependency resolution
✅ **Community:** GitHub-native workflow, familiar to developers

### Negative

⚠️ **Complexity:** 100+ validation rules require extensive testing
⚠️ **Maintenance:** Security patterns need regular updates
⚠️ **GitHub Dependency:** Requires GitHub token for publishing
⚠️ **Cache Invalidation:** 15min TTL may serve stale data

### Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| False positives (security scan) | High | Medium | Whitelist safe patterns, severity levels |
| GitHub rate limits | Medium | Low | Cache aggressively, backoff/retry |
| Malicious collection bypasses | Critical | Low | Multi-layered defense, community review |
| Performance degradation (100+ rules) | Medium | Low | Parallel validation, early exit |

---

## Implementation Notes

### Week 1: Validation & Security (8 story points)

**Day 1-2: Enhanced Validator**
- Extend `internal/collection/validator.go`
- Add 100+ rules across 5 categories
- Structured error reporting (`ValidationError`, `ValidationResult`)
- Unit tests for each rule (~150 test cases)

**Day 3-4: Security System**
- Create `internal/collection/security/` package
- Implement signature verification (GPG + SSH)
- Checksum validation (SHA-256)
- Malicious code scanner (50+ patterns)
- Trusted source registry
- Integration tests

**Day 5: Registry Enhancements**
- In-memory cache with TTL (15min default)
- Metadata indexing (author, category, tags, keywords)
- Search implementation
- Performance tests (verify <10ms cached)

### Week 2: Publishing & Integration (5 story points)

**Day 1-2: GitHub Publisher**
- Create `internal/infrastructure/github_publisher.go`
- Fork, clone, commit, push, PR workflow
- Release creation
- Integration tests (mock GitHub API)

**Day 3: Publishing Tool**
- Create `internal/mcp/publishing_tools.go`
- `publish_collection` MCP tool
- Dry-run mode
- Progress reporting
- Error handling

**Day 4: Enhanced Discovery**
- Update `list_collections` (rich formatting)
- Create `search_collections` tool
- Metadata display improvements
- Integration tests

**Day 5: Documentation & Testing**
- `docs/collections/REGISTRY.md` (registry usage)
- `docs/collections/PUBLISHING.md` (publishing guide)
- `docs/collections/SECURITY.md` (security model)
- Integration tests (full workflows)
- Update NEXT_STEPS.md

---

## Success Criteria

### Functional Requirements

- ✅ Manifest validation catches 100+ error types
- ✅ Security scan detects 50+ malicious patterns
- ✅ publish_collection creates valid PR in <2 minutes
- ✅ Registry caching reduces network calls by >90%
- ✅ Dependency resolution handles circular deps, version conflicts
- ✅ All existing tests passing (0 regressions)

### Non-Functional Requirements

- ✅ Performance: Cached browse <10ms (p95), cached get <5ms
- ✅ Security: Zero known bypasses for validation/scanning
- ✅ Reliability: Atomic installation with rollback
- ✅ Usability: Publishing takes <5 minutes end-to-end
- ✅ Maintainability: Comprehensive tests (>90% coverage)

### Documentation

- ✅ `docs/collections/REGISTRY.md` (usage guide)
- ✅ `docs/collections/PUBLISHING.md` (step-by-step publishing)
- ✅ `docs/collections/SECURITY.md` (security model)
- ✅ ADR-008 (this document) complete
- ✅ API documentation (GoDoc comments)

---

## Future Enhancements (Post-M0.8)

### M1.4: Collection Plugins (4 weeks)
- Custom validation plugins
- Language-specific validators (Python, TypeScript, etc.)
- Community-contributed rules

### M1.5: Analytics & Insights (2 weeks)
- Download tracking
- Usage statistics
- Popularity metrics
- Trending collections

### M1.6: Advanced Security (3 weeks)
- Sandboxed hook execution
- Runtime monitoring
- Vulnerability scanning
- Automated security audits

### M2.0: Web UI (6 weeks)
- Collection browser
- Publishing wizard
- Analytics dashboard
- Review interface

---

## References

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [DollhouseMCP Collections](https://github.com/DollhouseMCP/mcp-server)
- [npm Registry Architecture](https://docs.npmjs.com/)
- [PyPI Security Model](https://warehouse.pypa.io/)
- [ADR-007: MCP Resources Implementation](./ADR-007-mcp-resources-implementation.md)
- [comparing.md](../../comparing.md)
- [NEXT_STEPS.md](../../NEXT_STEPS.md)

---

## Approval

**Proposed:** 2025-12-19  
**Reviewed:**  
**Approved:**  
**Implemented:**  

**Reviewers:**
- [ ] Lead Architect
- [ ] Security Team
- [ ] Backend Team
- [ ] DevOps Team

**Status:** ✅ Ready for Implementation
