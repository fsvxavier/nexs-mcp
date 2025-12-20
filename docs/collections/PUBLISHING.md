# Collection Publishing Guide

Complete guide for publishing NEXS MCP collections to GitHub with automated workflows.

## Overview

The publishing system provides:
- **Automated validation**: 100+ manifest rules
- **Security scanning**: 50+ malicious pattern detection
- **GitHub automation**: Fork, clone, commit, PR creation
- **Release management**: Tarball creation, checksums, signatures
- **Dry-run mode**: Test before publishing

## Quick Start

### Using the MCP Tool

```bash
# Basic publish
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN"
}'

# With release creation
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN",
  "create_release": true
}'

# Dry-run (test without publishing)
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN",
  "dry_run": true
}'
```

## Publishing Workflow

### 7-Step Process

```
1. Load Manifest
   ├─ Read collection.yaml
   └─ Parse and validate structure

2. Validate Manifest (100+ rules)
   ├─ Schema validation (30 rules)
   ├─ Security validation (25 rules)
   ├─ Dependency validation (15 rules)
   ├─ Element validation (20 rules)
   └─ Hook validation (10 rules)

3. Security Scan (50+ patterns)
   ├─ Code injection detection
   ├─ Malicious command detection
   ├─ Path traversal detection
   └─ Credential leak detection

4. Create Tarball
   ├─ Package collection files
   ├─ Generate SHA-256 checksum
   └─ Generate SHA-512 checksum

5. Fork Repository
   ├─ Create GitHub fork
   └─ Wait for fork completion

6. Commit & Push
   ├─ Clone forked repository
   ├─ Create feature branch
   ├─ Add collection files
   ├─ Commit changes
   └─ Push to fork

7. Create Pull Request
   ├─ Generate PR description
   ├─ Include metadata
   ├─ Add checksums
   └─ Submit PR to registry
```

## Step-by-Step Guide

### 1. Prepare Your Collection

Create a `collection.yaml` manifest:

```yaml
name: my-awesome-collection
version: 1.0.0
author: john@example.com
description: A collection of AI assistants
category: ai-ml
license: MIT

repository: https://github.com/user/my-collection
homepage: https://example.com/collections/my-awesome

tags:
  - ai
  - nlp
  - automation

elements:
  - path: personas/*.yaml
    type: persona
  - path: skills/*.yaml
    type: skill
  - path: templates/*.yaml
    type: template

dependencies:
  - uri: github.com/nexs-mcp/core-skills
    version: ^1.0.0

hooks:
  post_install:
    - type: command
      command: echo "Collection installed successfully"
```

### 2. Test Locally

```bash
# Validate manifest
nexs-mcp validate collection.yaml

# Test installation locally
nexs-mcp install file://./path/to/collection

# Run security scan
nexs-mcp scan ./collection
```

### 3. Publish (Dry-Run)

Test the publishing process without actually creating a PR:

```bash
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN",
  "dry_run": true
}'
```

Expected output:
```json
{
  "status": "dry_run_success",
  "message": "✅ Dry-run completed successfully",
  "tarball_path": "/tmp/my-awesome-collection-1.0.0.tar.gz",
  "checksums": {
    "sha256": "abc123...",
    "sha512": "def456..."
  },
  "validation": {
    "valid": true,
    "errors": 0,
    "warnings": 0
  },
  "security_scan": {
    "clean": true,
    "findings": 0
  }
}
```

### 4. Publish (Production)

If dry-run succeeds, publish to the registry:

```bash
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN"
}'
```

Expected output:
```json
{
  "status": "success",
  "message": "✅ Collection published successfully",
  "pr_url": "https://github.com/nexs-mcp/registry/pull/123",
  "pr_number": 123,
  "tarball_path": "/tmp/my-awesome-collection-1.0.0.tar.gz",
  "checksums": {
    "sha256": "abc123...",
    "sha512": "def456..."
  }
}
```

### 5. Monitor PR

Your PR will be automatically created with:
- Collection metadata
- File checksums
- Validation results
- Security scan results

Maintainers will review and merge your contribution.

## Pull Request Template

The system generates a detailed PR description:

```markdown
# Add my-awesome-collection v1.0.0

## Collection Information
- **Author**: john@example.com
- **Category**: ai-ml
- **License**: MIT
- **Homepage**: https://example.com/collections/my-awesome

## Description
A collection of AI assistants

## Tags
ai, nlp, automation

## Elements
- 5 personas
- 3 skills
- 2 templates

## Dependencies
- github.com/nexs-mcp/core-skills ^1.0.0

## Checksums
- **SHA-256**: abc123...
- **SHA-512**: def456...

## Validation
✅ All validation checks passed (100+ rules)

## Security Scan
✅ No security issues detected (50+ patterns)

---
*This PR was automatically generated by the NEXS MCP publishing system.*
```

## Configuration Options

### MCP Tool Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `manifest_path` | string | ✅ | - | Path to collection.yaml |
| `registry` | string | ✅ | - | Target registry (e.g., github.com/nexs-mcp/registry) |
| `github_token` | string | ✅ | - | GitHub personal access token |
| `create_release` | boolean | ❌ | false | Create GitHub release |
| `dry_run` | boolean | ❌ | false | Test without publishing |
| `skip_security_scan` | boolean | ❌ | false | Skip security validation |

### GitHub Token Permissions

Required scopes:
- `repo` (full repository access)
- `workflow` (if using GitHub Actions)

Create token: https://github.com/settings/tokens/new

## Validation Rules

### Schema Validation (30 rules)
- Required fields (name, version, author, description)
- Version format (semver)
- Email format (author, maintainers)
- URL format (repository, homepage)
- Category values
- License SPDX identifiers

### Security Validation (25 rules)
- Path traversal prevention
- Command injection detection
- Shell expansion prevention
- Hardcoded credentials detection
- Unsafe file permissions

### Dependency Validation (15 rules)
- URI format
- Version constraints
- Circular dependency detection
- Dependency resolution

### Element Validation (20 rules)
- Path validation
- Type validation
- File existence
- Glob pattern support

### Hook Validation (10 rules)
- Command safety
- Type validation
- Required tool checks

## Security Scanning

### 50+ Pattern Categories

**Critical (15 patterns):**
- `eval` injection
- `exec` injection
- `rm -rf` commands
- `curl | bash` execution
- Fork bombs

**High (20 patterns):**
- Netcat listeners
- `chmod 777` permissions
- SQL injection
- Unsafe deserialization
- Privilege escalation

**Medium (10 patterns):**
- Base64 decode operations
- Hardcoded credentials
- Debug statements
- Unsafe temp files

**Low (5 patterns):**
- Console logging
- Development artifacts

## Release Management

### With Release Creation

```bash
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN",
  "create_release": true
}'
```

Creates:
- Git tag: `v1.0.0`
- GitHub release with tarball
- SHA-256 and SHA-512 checksums
- Release notes from changelog

### Tarball Contents

```
my-awesome-collection-1.0.0/
├── collection.yaml
├── SHA256SUMS
├── SHA512SUMS
├── personas/
│   ├── assistant-1.yaml
│   └── assistant-2.yaml
├── skills/
│   └── automation.yaml
└── templates/
    └── default.yaml
```

## Error Handling

### Common Errors

**Invalid Manifest:**
```json
{
  "status": "validation_failed",
  "validation_errors": [
    {
      "field": "version",
      "rule": "format",
      "message": "invalid semver format: v1.0",
      "fix": "Use format: 1.0.0"
    }
  ]
}
```

**Security Issues:**
```json
{
  "status": "security_failed",
  "security_findings": [
    {
      "severity": "critical",
      "file": "scripts/install.sh",
      "line": 10,
      "rule": "eval-injection",
      "description": "eval() can execute arbitrary code"
    }
  ]
}
```

**GitHub Errors:**
```json
{
  "status": "github_error",
  "message": "Failed to create PR: rate limit exceeded",
  "retry_after": 3600
}
```

## Best Practices

1. **Always dry-run first**: Test before publishing
2. **Keep manifests simple**: Minimal dependencies
3. **Version properly**: Follow semantic versioning
4. **Document well**: Clear descriptions and tags
5. **Test locally**: Install and test before publishing
6. **Security first**: Never skip security scans
7. **License clearly**: Use SPDX identifiers
8. **Tag appropriately**: Help users discover your collection

## Troubleshooting

### Validation Fails

1. Check manifest syntax (YAML)
2. Verify all required fields
3. Validate version format (semver)
4. Fix file paths (no traversal)
5. Review hooks (no dangerous commands)

### Security Scan Fails

1. Review flagged files
2. Remove eval/exec calls
3. Check file permissions
4. Avoid hardcoded secrets
5. Use safe commands

### PR Creation Fails

1. Verify GitHub token permissions
2. Check rate limits
3. Ensure fork doesn't exist
4. Validate registry repository

### Fork Issues

- Wait 5 seconds after fork creation
- Check repository visibility
- Verify token has fork permissions

## Advanced Usage

### Custom Registry

```bash
# Publish to custom registry
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/myorg/my-registry",
  "github_token": "$GITHUB_TOKEN"
}'
```

### Skip Security Scan (Not Recommended)

```bash
# Only for trusted internal collections
mcp call publish_collection '{
  "manifest_path": "./collection.yaml",
  "registry": "github.com/nexs-mcp/registry",
  "github_token": "$GITHUB_TOKEN",
  "skip_security_scan": true
}'
```

### Programmatic Publishing

```go
import (
    "github.com/fsvxavier/nexs-mcp/internal/infrastructure"
    "github.com/fsvxavier/nexs-mcp/internal/collection"
)

publisher := infrastructure.NewGitHubPublisher(token)

opts := &infrastructure.PublishOptions{
    RegistryOwner: "nexs-mcp",
    RegistryRepo:  "registry",
    BranchName:    "add-my-collection-v1.0.0",
    CommitMessage: "Add my-collection v1.0.0",
    PRTitle:       "Add my-collection v1.0.0",
    PRBody:        prTemplate,
    CreateRelease: true,
}

pr, err := publisher.PublishCollection(ctx, manifest, tarballPath, opts)
```

## See Also

- [Registry Guide](REGISTRY.md)
- [Security Validation](SECURITY.md)
- [Collection Manifest Spec](../collection-manifest-example.yaml)
- [GitHub Publisher API](../../internal/infrastructure/github_publisher.go)
