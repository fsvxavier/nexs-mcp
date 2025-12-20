# NPM Distribution Guide

**Version**: 0.11.0  
**Package**: `@nexs-mcp/server`  
**Status**: ✅ Complete (M0.9)

## Overview

NEXS MCP Server is distributed as an NPM package for easy installation and integration with Claude Desktop and other MCP clients. The package wraps pre-compiled Go binaries with Node.js scripts for cross-platform compatibility.

## Architecture

### Package Structure

```
@nexs-mcp/server/
├── package.json              # NPM package configuration
├── index.js                  # Entry point (exports binary path)
├── bin/
│   ├── nexs-mcp.js          # Binary wrapper script
│   ├── nexs-mcp-darwin-amd64     # macOS Intel binary
│   ├── nexs-mcp-darwin-arm64     # macOS Apple Silicon binary
│   ├── nexs-mcp-linux-amd64      # Linux x64 binary
│   ├── nexs-mcp-linux-arm64      # Linux ARM64 binary
│   ├── nexs-mcp-windows-amd64.exe # Windows x64 binary
│   └── nexs-mcp-windows-arm64.exe # Windows ARM64 binary
├── scripts/
│   ├── install-binary.js    # Post-install setup
│   └── test.js              # Installation verification
└── README.npm.md            # NPM-specific documentation
```

### Binary Wrapper (`bin/nexs-mcp.js`)

The wrapper script:
1. Detects the current platform and architecture
2. Selects the appropriate binary
3. Spawns the Go binary with all arguments
4. Pipes stdio (stdin, stdout, stderr)
5. Forwards exit codes

**Platform Detection**:
```javascript
const binaryMap = {
  'darwin-x64': 'nexs-mcp-darwin-amd64',
  'darwin-arm64': 'nexs-mcp-darwin-arm64',
  'linux-x64': 'nexs-mcp-linux-amd64',
  'linux-arm64': 'nexs-mcp-linux-arm64',
  'win32-x64': 'nexs-mcp-windows-amd64.exe',
  'win32-arm64': 'nexs-mcp-windows-arm64.exe'
};
```

### Post-Install Script (`scripts/install-binary.js`)

Runs after `npm install` to:
1. Verify the appropriate binary exists
2. Set executable permissions (Unix systems)
3. Display installation success message

### Test Script (`scripts/test.js`)

Verifies installation with 4 tests:
1. Binary exists in expected location
2. Binary is executable
3. Help command works
4. MCP server can be invoked

## Installation Methods

### 1. Global Installation

```bash
npm install -g @nexs-mcp/server
nexs-mcp --help
```

**Use case**: System-wide availability

### 2. Local Installation

```bash
npm install @nexs-mcp/server
npx nexs-mcp --help
```

**Use case**: Project-specific dependency

### 3. NPX (No Installation)

```bash
npx @nexs-mcp/server --help
```

**Use case**: One-off execution, CI/CD

### 4. Claude Desktop Integration

**macOS Configuration** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "npx",
      "args": ["@nexs-mcp/server"],
      "env": {
        "NEXS_DATA_DIR": "/Users/username/nexs/elements"
      }
    }
  }
}
```

**Linux Configuration** (`~/.config/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "npx",
      "args": ["@nexs-mcp/server"],
      "env": {
        "NEXS_DATA_DIR": "/home/username/nexs/elements"
      }
    }
  }
}
```

**Windows Configuration** (`%APPDATA%\Claude\claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "npx.cmd",
      "args": ["@nexs-mcp/server"],
      "env": {
        "NEXS_DATA_DIR": "C:\\Users\\username\\nexs\\elements"
      }
    }
  }
}
```

## Building Binaries

### Prerequisites

- Go 1.21+
- Make (optional, for Makefile commands)

### Build All Platforms

```bash
# Using Makefile
make build-all

# Manual build for all platforms
GOOS=darwin GOARCH=amd64 go build -o bin/nexs-mcp-darwin-amd64 ./cmd/nexs-mcp
GOOS=darwin GOARCH=arm64 go build -o bin/nexs-mcp-darwin-arm64 ./cmd/nexs-mcp
GOOS=linux GOARCH=amd64 go build -o bin/nexs-mcp-linux-amd64 ./cmd/nexs-mcp
GOOS=linux GOARCH=arm64 go build -o bin/nexs-mcp-linux-arm64 ./cmd/nexs-mcp
GOOS=windows GOARCH=amd64 go build -o bin/nexs-mcp-windows-amd64.exe ./cmd/nexs-mcp
GOOS=windows GOARCH=arm64 go build -o bin/nexs-mcp-windows-arm64.exe ./cmd/nexs-mcp
```

### Build Current Platform Only

```bash
# Using Makefile
make build

# Manual build
go build -o bin/nexs-mcp ./cmd/nexs-mcp
```

### Binary Sizes

Typical binary sizes (uncompressed):
- Darwin (macOS): ~12-14 MB
- Linux: ~12-14 MB  
- Windows: ~12-14 MB

Total package size with all binaries: ~70-85 MB

## Testing Local Installation

### 1. Test Package Locally

```bash
# Install from local directory
npm install .

# Test wrapper
node bin/nexs-mcp.js --help

# Run test suite
npm test
```

### 2. Test with NPX

```bash
# Link package locally
npm link

# Test via npx
npx @nexs-mcp/server --help
```

### 3. Integration Test

```bash
# Create test element
echo 'name: test-agent
type: agent
description: Test agent for NPM installation' > /tmp/test-agent.yaml

# Start server with test data
NEXS_DATA_DIR=/tmp npx @nexs-mcp/server
```

## Publishing to NPM

### Prerequisites

1. NPM account: https://www.npmjs.com/signup
2. Login: `npm login`
3. Scope access: `@nexs-mcp` organization

### Publishing Steps

```bash
# 1. Build all binaries
make build-all

# 2. Verify package contents
npm pack --dry-run

# 3. Test installation locally
npm install .
npm test

# 4. Publish to NPM
npm publish --access public

# 5. Verify publication
npm view @nexs-mcp/server
```

### Version Management

```bash
# Patch version (0.11.0 → 0.11.1)
npm version patch

# Minor version (0.11.0 → 0.12.0)
npm version minor

# Major version (0.11.0 → 1.0.0)
npm version major
```

## Package Configuration

### package.json Key Fields

```json
{
  "name": "@nexs-mcp/server",
  "version": "0.11.0",
  "description": "MCP server for AI element management",
  "main": "index.js",
  "bin": {
    "nexs-mcp": "./bin/nexs-mcp.js"
  },
  "scripts": {
    "postinstall": "node scripts/install-binary.js",
    "test": "node scripts/test.js"
  },
  "os": ["darwin", "linux", "win32"],
  "cpu": ["x64", "arm64"]
}
```

### .npmignore

Excludes from package:
- Source code (`cmd/`, `internal/`, `pkg/`)
- Test files (`*_test.go`, `test/`)
- Development files (`.git`, `.vscode`, etc.)
- Build tools (`Makefile`, `go.mod`, etc.)

Includes in package:
- Binaries (`bin/nexs-mcp-*`)
- Wrapper scripts (`bin/nexs-mcp.js`, `scripts/*.js`)
- Package files (`package.json`, `index.js`)
- Documentation (`README.npm.md`, `LICENSE`, `CHANGELOG.md`)

## Programmatic Usage

### Import Package

```javascript
const nexsMcp = require('@nexs-mcp/server');

console.log('Binary path:', nexsMcp.binaryPath);
console.log('Platform:', nexsMcp.platform);
console.log('Architecture:', nexsMcp.arch);
```

### Spawn Server Programmatically

```javascript
const { spawn } = require('child_process');
const { binaryPath } = require('@nexs-mcp/server');

const server = spawn(binaryPath, [], {
  env: {
    ...process.env,
    NEXS_DATA_DIR: '/path/to/elements'
  }
});

server.stdout.on('data', (data) => {
  console.log('Server:', data.toString());
});

server.on('exit', (code) => {
  console.log('Server exited with code', code);
});
```

## Troubleshooting

### Binary Not Found

**Error**: `Binary not found: /path/to/nexs-mcp-xxx`

**Solution**:
1. Run `npm install` to trigger post-install script
2. Verify platform: `node -p "process.platform + '-' + process.arch"`
3. Check bin/ directory: `ls node_modules/@nexs-mcp/server/bin/`

### Permission Denied (Unix)

**Error**: `EACCES: permission denied`

**Solution**:
```bash
chmod +x node_modules/@nexs-mcp/server/bin/nexs-mcp-*
```

### Unsupported Platform

**Error**: `Unsupported platform: xxx-yyy`

**Solution**: Build binary manually for your platform:
```bash
go build -o custom-binary ./cmd/nexs-mcp
# Copy to: node_modules/@nexs-mcp/server/bin/
```

### Package Size Too Large

**Issue**: Package with all binaries is 70-85 MB

**Solutions**:
1. **Platform-specific packages** (future):
   - `@nexs-mcp/server-darwin-arm64`
   - `@nexs-mcp/server-linux-amd64`
   - etc.

2. **Download on install** (future):
   - Store binaries on GitHub Releases
   - Download appropriate binary during post-install

3. **Compression** (current):
   - Binaries are already UPX compressed

## Environment Variables

### NEXS_DATA_DIR
- **Description**: Base directory for NEXS elements
- **Default**: `~/.nexs/elements`
- **Example**: `/home/user/nexs/elements`

### NEXS_LOG_LEVEL
- **Description**: Logging verbosity
- **Values**: `debug`, `info`, `warn`, `error`
- **Default**: `info`

### NEXS_LOG_FORMAT
- **Description**: Log output format
- **Values**: `json`, `text`
- **Default**: `json`

### GITHUB_TOKEN
- **Description**: GitHub personal access token (for portfolio search)
- **Required**: No (optional feature)
- **Format**: `ghp_xxxxxxxxxxxxx`

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test NPM Package

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [16, 18, 20]
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node }}
      
      - name: Install package
        run: npm install .
      
      - name: Run tests
        run: npm test
      
      - name: Test CLI
        run: npx nexs-mcp --help
```

## Performance Considerations

### Startup Time
- **Cold start**: ~100-200ms (binary spawn)
- **Warm start**: ~50-100ms (cached binary)

### Memory Usage
- **Idle**: ~20-30 MB
- **Active**: ~50-100 MB (depends on element count)

### Binary Size Optimization
- Built with Go compiler optimizations
- Stripped debug symbols (`-ldflags="-s -w"`)
- UPX compression (optional, reduces size by ~50%)

## Future Enhancements

### M0.10: Advanced Distribution
- [ ] Platform-specific NPM packages
- [ ] Binary download on install (reduce package size)
- [ ] Auto-update mechanism
- [ ] Homebrew formula (macOS)
- [ ] Snapcraft package (Linux)
- [ ] Chocolatey package (Windows)

### M0.12: Enhanced NPM Integration
- [ ] TypeScript definitions for programmatic use
- [ ] Node.js API wrapper
- [ ] Event emitters for server lifecycle
- [ ] Promise-based tool invocation
- [ ] Stream-based element operations

## Related Documentation

- [README.npm.md](../README.npm.md) - User-facing NPM documentation
- [package.json](../package.json) - NPM package configuration
- [.npmignore](../.npmignore) - Package exclusion rules
- [User Guide](USER_GUIDE.md) - General usage documentation

## Changelog

### v0.11.0 (2024-12-20)
- ✅ Initial NPM distribution support (M0.9)
- ✅ Cross-platform binary wrapper
- ✅ Post-install binary setup
- ✅ Installation verification tests
- ✅ Claude Desktop integration guide
- ✅ Programmatic API (index.js)

---

**Status**: Production Ready  
**Milestone**: M0.9 Complete  
**Last Updated**: 2024-12-20
