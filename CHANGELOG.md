# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2025-12-23

### Added
- **Memory Quality System (Sprint 8):**
  - ONNX Quality Scorer with 2 production models (MS MARCO + Paraphrase-Multilingual)
  - Multi-tier fallback system: ONNX → Groq API → Gemini API → Implicit Signals
  - Quality-based retention policies (High: 365d, Medium: 180d, Low: 90d)
  - 3 new MCP tools: `score_memory_quality`, `get_retention_policy`, `get_retention_stats`
  - Memory retention service with automatic archival
  - 91 MCP tools total (88 + 3 quality tools)
- **ONNX Models:**
  - MS MARCO MiniLM-L-6-v2 (default): 61.64ms latency, 9 languages
  - Paraphrase-Multilingual-MiniLM-L12-v2 (configurable): 109.41ms latency, 11 languages with CJK support
  - Automatic CJK skip for MS MARCO (Japanese/Chinese)
  - Comprehensive benchmarks: speed, concurrency, effectiveness, text-length
- **Documentation:**
  - BENCHMARK_RESULTS.md - Complete performance analysis
  - ONNX_QUALITY_AUDIT.md - Technical audit (80% conforme)
  - ONNX_MODEL_CONFIGURATION.md - User configuration guide
  - QUALITY_USAGE_ANALYSIS.md - Internal usage analysis (100% conforme)

### Changed
- Updated DefaultConfig() to use MS MARCO as default model
- Removed all Distiluse model references (DistiluseV1, DistiluseV2 discontinued)
- Enhanced test helpers to support both production models

### Performance
- MS MARCO: 61.64ms average (9 languages, non-CJK)
- Paraphrase-Multilingual: 109.41ms average (11 languages, full coverage)
- Zero cost, full privacy, offline-capable scoring
- 100% test passing rate

## [1.0.1] - 2025-12-20

### Added
- GitHub community infrastructure:
  - Issue templates (bug report, feature request, question, element submission)
  - Pull request template with comprehensive checklist
  - Community files (CODE_OF_CONDUCT.md, SECURITY.md, SUPPORT.md)
- Comprehensive benchmark suite:
  - 12 performance benchmarks covering CRUD, search, validation, concurrency
  - Automated comparison script (benchmark/compare.sh)
  - Detailed documentation and results analysis
- Template validator enhancements:
  - Variable type validation (string, number, boolean, array, object)
  - Handlebars block helper validation ({{#if}}/{{/if}})
  - Unbalanced delimiter detection

### Fixed
- Template validator now properly validates variable types
- Template validator detects unclosed Handlebars blocks
- Template validator detects unbalanced delimiters (}} without {{)
- TestTokenizeAndCount test data corrected

### Changed
- CI: Updated golangci-lint to v2.7.1 for consistency with local development

### Performance
- Element Create: ~115µs
- Element Read: ~195ns
- Element Update: ~111µs
- Element Delete: ~20µs
- Element List: ~9µs
- Search by Type: ~9µs
- Search by Tags: ~2µs
- Validation: ~274ns
- Startup Time: ~1.1ms

All performance targets met ✅

## [1.0.0] - 2025-12-19

### Added
- Initial release with core MCP server functionality
- Element management (Agent, Persona, Skill, Ensemble, Memory, Template)
- Template system with Handlebars support
- Collection management with registry
- GitHub integration for portfolio sync
- Distribution via NPM, Docker, and Homebrew
- Enhanced indexing with TF-IDF search
- Backup and restore functionality
- Access control and security features

[1.3.0]: https://github.com/fsvxavier/nexs-mcp/compare/v1.0.1...v1.3.0
[1.0.1]: https://github.com/fsvxavier/nexs-mcp/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/fsvxavier/nexs-mcp/releases/tag/v1.0.0
