# Changelog

All notable changes to NEXS MCP Server will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0-dev] - 2025-12-19

### Added - M0.6 Analytics & Convenience (5/7 Tasks Complete)

#### Analytics Tools
- **get_usage_stats** tool - Usage analytics with period filtering
  - Period options: last_hour, last_24h, last_7_days, last_30_days, all
  - Metrics: total ops, success rate, operations by tool, errors by tool
  - Top 10 most used tools with operation counts
  - Top 10 slowest operations with durations
  - Recent errors with timestamps and details
  - Active users list and daily breakdown
  - JSON persistence in ~/.nexs-mcp/metrics
  - Auto-save every 5 minutes with circular buffer (10k metrics)

- **get_performance_dashboard** tool - Performance monitoring with latency percentiles
  - Latency percentiles: p50 (median), p95, p99
  - Slow operations identification (>p95 latency)
  - Fast operations tracking (<p50 latency)
  - Per-operation statistics (count, avg, max, min duration)
  - Period-based filtering (hour/day/week/month/all)
  - Top 10 slow and fast operations
  - JSON persistence in ~/.nexs-mcp/performance
  - Thread-safe circular buffer (10k operations)

#### Convenience Tools
- **duplicate_element** tool - Element duplication with metadata preservation
  - Optional new name (defaults to "Copy of {original}")
  - Automatic ID generation with timestamp suffix
  - Access control integration (read permission check)
  - Preserves tags, description, version, author
  - Returns new element with complete metadata

- **list_elements active_only filter** - Enhanced list_elements tool
  - New `active_only` boolean parameter
  - Takes priority over explicit `is_active` parameter
  - Resolves get_active_elements gap
  - Backward compatible with existing usage

### Technical Implementation

#### New Packages
- `internal/application/statistics.go` (335 LOC) - Usage analytics engine
- `internal/application/statistics_test.go` (265 LOC) - 5 comprehensive tests
- `internal/logger/metrics.go` (318 LOC) - Performance metrics tracker
- `internal/logger/metrics_test.go` (225 LOC) - 8 comprehensive tests

#### New MCP Handlers
- `internal/mcp/analytics_tools.go` (75 LOC) - get_usage_stats handler
- `internal/mcp/performance_tools.go` (73 LOC) - get_performance_dashboard handler

#### Enhanced Files
- `internal/mcp/tools.go` (+138 LOC) - New input/output structs, duplicate handler
- `internal/mcp/server.go` (+21 LOC) - Metrics initialization and tool registration

### Test Results
- 13 new tests added (all passing)
- Statistics tests: 5/5 PASS (period filtering, save/load, aggregation)
- Performance tests: 8/8 PASS (percentiles, slow ops, circular buffer)
- Total test count: 182+ tests
- 100% pass rate maintained

### Metrics & Analytics

#### Usage Statistics Features
- Circular buffer: 10,000 max metrics with auto-eviction
- Period aggregation: Hour/day/week/month/all
- Success rate calculation with percentages
- Most used tools ranking (top 10)
- Slowest operations tracking (top 10)
- Error tracking with recent errors list
- Active users identification
- Daily operation breakdown
- Auto-save every 5 minutes

#### Performance Metrics Features
- Percentile calculation: p50, p95, p99 with linear interpolation
- Slow operation alerts: >p95 threshold
- Fast operation tracking: <p50 baseline
- Per-operation stats: count, avg, max, min duration
- Period filtering: last_hour to all-time
- JSON persistence: ~/.nexs-mcp/performance/performance_metrics.json
- Thread-safe: sync.RWMutex for concurrent access
- Circular buffer: 10,000 operations max

### Gap Resolution
- ✅ get_active_elements gap - Resolved via list_elements active_only filter
- ✅ duplicate_element gap - Full implementation with access control
- ✅ get_usage_stats gap - Comprehensive analytics tool
- ⏳ submit_to_collection gap - Still pending (M0.7 planned)

### Tool Count Update
- Before M0.6: 44 tools
- After M0.6: 47 tools (+3)
- Total story points: 13/18 complete (72%)
- Gaps resolved: 2/4 (50%)

## [0.5.0-dev] - 2025-12-19

### Added - M0.5 Production Readiness (8/9 Tasks Complete)

#### Backup & Restore System
- **backup_portfolio** tool - Create tar.gz backups with SHA-256 checksums
- **restore_portfolio** tool - Restore portfolios with validation and rollback
- **activate_element** tool - Shortcut to activate elements
- **deactivate_element** tool - Shortcut to deactivate elements
- Compression support (none, fast, best)
- Atomic restore operations with automatic rollback on failure
- Merge strategies (skip, overwrite, merge) for conflict resolution
- Pre-calculated checksums for integrity verification

#### Memory Management Tools
- **search_memory** tool - Semantic search with relevance scoring (content: 5pts/word, name: 25pts/word)
- **summarize_memories** tool - Statistics and top authors analysis
- **update_memory** tool - Partial updates with automatic hash recalculation
- **delete_memory** tool - Delete specific memories
- **clear_memories** tool - Bulk delete with date filters
- Date-based sorting for equal relevance scores

#### Structured Logging
- slog-based logging package with JSON/text format support
- Context extraction for request_id, user, operation, tool
- Configurable log levels (debug, info, warn, error)
- **list_logs** tool - Query logs with 7 filter criteria
- LogBuffer with circular 1000-entry storage
- Thread-safe buffered logging handler
- Command-line flags: `--log-level`, `--log-format`

#### User Identity Management
- UserSession global singleton with thread-safe operations
- **get_current_user** tool - Returns username, auth status, metadata
- **set_user_context** tool - Set user identity with metadata
- **clear_user_context** tool - Clear session (requires confirmation)
- Metadata support for custom user attributes

#### GitHub Authentication
- OAuth2 device flow integration
- **check_github_auth** tool - Verify token status and get username
- **refresh_github_token** tool - Auto-refresh expired tokens (24h threshold)
- **init_github_auth** tool - Initiate device flow authentication
- Device code storage in MCPServer for polling
- Integration with GitHubOAuthClient and GitHubClient

### Improved

#### Test Coverage
- Logger package: 24.5% → 92.1% (+67.6%)
- Added 30 comprehensive tests in `buffer_test.go`
- Created COVERAGE_REPORT.md with gap analysis
- Overall project coverage: 72.2% (excluding main)
- All tests passing (100% pass rate)
- Total test count: 169+ tests

#### Documentation
- Updated README.md with 44 MCP tools documentation
- Created comprehensive tool categorization
- Added M0.5 milestone tracking
- Updated badges (coverage, version, tool count)

### Technical Details

#### New Packages
- `internal/backup` - Backup and restore services (56.3% coverage)
- `internal/logger` - Structured logging (92.1% coverage)
- Buffer implementation with circular queue
- Filtering support for 7 criteria

#### New Files Created (M0.5)
- `internal/backup/backup.go` (318 LOC)
- `internal/backup/restore.go` (401 LOC)
- `internal/backup/backup_test.go` (388 LOC)
- `internal/mcp/backup_tools.go` (280 LOC)
- `internal/mcp/backup_tools_test.go` (389 LOC)
- `internal/mcp/memory_tools.go` (489 LOC)
- `internal/mcp/memory_tools_test.go` (543 LOC)
- `internal/logger/logger.go` (164 LOC)
- `internal/logger/buffer.go` (310 LOC)
- `internal/logger/logger_test.go` (344 LOC)
- `internal/logger/buffer_test.go` (600+ LOC)
- `internal/mcp/log_tools.go` (109 LOC)
- `internal/mcp/log_tools_test.go` (219 LOC)
- `internal/mcp/user_tools.go` (202 LOC)
- `internal/mcp/user_tools_test.go` (267 LOC)
- `internal/mcp/github_auth_tools.go` (208 LOC)
- `internal/mcp/github_auth_tools_test.go` (148 LOC)
- `COVERAGE_REPORT.md` - Test coverage analysis
- `CHANGELOG.md` - This file

#### Modified Files
- `internal/mcp/server.go` - Registered 16 new tools (28 → 44)
- `internal/config/config.go` - Added LogLevel, LogFormat fields
- `cmd/nexs-mcp/main.go` - Migrated to slog, updated tool count

### Commits (M0.5)
- `feat: implement backup and restore system with tar.gz compression`
- `feat: implement backup MCP tools (backup_portfolio, restore_portfolio, activate, deactivate)`
- `feat: implement memory management tools with relevance scoring`
- `feat: implement structured logging with slog and buffered handler`
- `feat: implement list_logs MCP tool with filtering`
- `feat: implement user identity and GitHub auth tools`
- `test: improve logger package coverage to 92.1%`

### Performance
- Circular buffer: O(1) add, O(n) query
- Relevance scoring: O(n*m) where n=memories, m=words
- Backup compression: Configurable levels (none/fast/best)
- Thread-safe operations with sync.RWMutex

### Breaking Changes
None - All changes are additive

---

## [0.4.0] - 2025-12-18

### Added - M0.4 Collection System

#### Collection Management (10 tools)
- **browse_collections** - Discover collections from multiple sources
- **install_collection** - Install from github://, file://, https:// URIs
- **uninstall_collection** - Remove installed collections
- **list_installed_collections** - List all installed collections
- **get_collection_info** - Get collection metadata
- **export_collection** - Export to tar.gz archive
- **update_collection** - Update specific collection
- **update_all_collections** - Batch update all collections
- **check_collection_updates** - Check for available updates
- **publish_collection** - Publish to GitHub repository

#### GitHub Integration (5 tools)
- **github_auth_start** - OAuth2 device flow
- **github_auth_status** - Token validation
- **github_list_repos** - Repository listing
- **github_sync_push** - Push elements to GitHub
- **github_sync_pull** - Pull elements from GitHub

#### Infrastructure
- Collection Registry with source management
- GitHub, Local, and HTTP source providers
- Collection Installer with validation
- Manifest validation (YAML schema)
- Version tracking and dependency management

---

## [0.2.0] - 2025-12-18

### Added - M0.2 Element System Complete

#### Element Types (6 types)
- Persona - AI behavior customization
- Skill - Reusable capabilities
- Template - Content templates
- Agent - Goal-oriented workflows
- Memory - Content storage with hashing
- Ensemble - Multi-agent orchestration

#### Type-Specific Tools (6 tools)
- **create_persona** - Persona creation with traits
- **create_skill** - Skill creation with triggers
- **create_template** - Template creation with variables
- **create_agent** - Agent creation with goals
- **create_memory** - Memory creation with hashing
- **create_ensemble** - Ensemble creation for coordination

#### Documentation
- Complete element documentation (~800 lines)
- Usage examples for all element types
- Integration patterns
- Best practices

---

## [0.1.0] - 2025-12-15

### Added - Initial Release

#### Core Features
- Clean Architecture implementation
- Official MCP SDK v1.1.0 integration
- Stdio transport support
- File-based YAML storage
- In-memory storage option

#### Generic CRUD Tools (5 tools)
- **list_elements** - List with filtering
- **get_element** - Get by ID
- **create_element** - Generic creation
- **update_element** - Update existing
- **delete_element** - Delete by ID

#### Infrastructure
- Element repository (file + memory)
- Configuration management
- Thread-safe operations
- Graceful shutdown
- Error handling

#### Testing
- Unit tests for all layers
- Integration tests
- 80%+ test coverage

---

## Release Notes

### Version Strategy
- **v0.x.x** - Pre-release development
- **v1.0.0** - First stable release (planned Q1 2026)

### Milestone Timeline
- **M0.1** - Foundation (Dec 2025) ✅
- **M0.2** - Element Types (Dec 2025) ✅
- **M0.4** - Collection System (Dec 2025) ✅
- **M0.5** - Production Readiness (Dec 2025) 🔄 88% Complete
- **M0.6** - Advanced Features (Q1 2026) 📋 Planned
- **M1.0** - Stable Release (Q1 2026) 📋 Planned

### Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

### License
MIT License - See [LICENSE](LICENSE) for details.
