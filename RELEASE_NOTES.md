# Release Notes - NEXS MCP v0.1.0

**Release Date:** December 18, 2025  
**Type:** Initial Release  
**Status:** Production Ready

## 🎉 Overview

First stable release of NEXS MCP Server - A high-performance Model Context Protocol server written in Go.

## ✨ Features

### Core Functionality

- ✅ **Complete MCP Protocol Implementation**
  - JSON-RPC 2.0 support
  - stdio transport
  - Graceful shutdown
  - Error handling

- ✅ **5 Essential Tools**
  - `list_elements` - List and filter elements
  - `get_element` - Retrieve element by ID
  - `create_element` - Create new elements
  - `update_element` - Update existing elements
  - `delete_element` - Remove elements

- ✅ **6 Element Types**
  - Persona
  - Skill
  - Template
  - Agent
  - Memory
  - Ensemble

### Storage

- ✅ **Dual Storage Modes**
  - File-based persistence (YAML)
  - In-memory (for testing)
  
- ✅ **File Organization**
  - Date-based directory structure (YYYY-MM-DD)
  - Type-based subdirectories
  - YAML serialization

### Architecture

- ✅ **Clean Architecture**
  - Domain layer (100% coverage)
  - Infrastructure layer (87.7% coverage)
  - MCP protocol layer (94.0% coverage)
  - Overall: 80.7% test coverage

- ✅ **Thread-Safe**
  - Concurrent operations
  - sync.RWMutex for safety
  - Race detector validated

### Developer Experience

- ✅ **Configuration**
  - Command-line flags
  - Environment variables
  - Sensible defaults

- ✅ **Build & Distribution**
  - Cross-compilation (5 platforms)
  - Docker support
  - Release artifacts

## 📦 Downloads

### Binaries

- [Linux (amd64)](dist/nexs-mcp-0.1.0-linux-amd64.tar.gz)
- [Linux (arm64)](dist/nexs-mcp-0.1.0-linux-arm64.tar.gz)
- [macOS (Intel)](dist/nexs-mcp-0.1.0-darwin-amd64.tar.gz)
- [macOS (Apple Silicon)](dist/nexs-mcp-0.1.0-darwin-arm64.tar.gz)
- [Windows (amd64)](dist/nexs-mcp-0.1.0-windows-amd64.zip)

### Docker

```bash
docker pull ghcr.io/fsvxavier/nexs-mcp:0.1.0
docker pull ghcr.io/fsvxavier/nexs-mcp:latest
```

## 🚀 Quick Start

### Installation

**From binary:**
```bash
# Linux/macOS
tar -xzf nexs-mcp-0.1.0-linux-amd64.tar.gz
chmod +x nexs-mcp-linux-amd64
sudo mv nexs-mcp-linux-amd64 /usr/local/bin/nexs-mcp
```

**From Docker:**
```bash
docker run --rm -it \
  -v $(pwd)/data:/app/data \
  ghcr.io/fsvxavier/nexs-mcp:0.1.0
```

**From source:**
```bash
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp
git checkout v0.1.0
make build
```

### Usage

```bash
# Default (file storage)
nexs-mcp

# Memory storage
nexs-mcp -storage memory

# Custom data directory
nexs-mcp -data-dir /path/to/data
```

### Claude Desktop Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/usr/local/bin/nexs-mcp",
      "args": ["-storage", "file"]
    }
  }
}
```

## 📊 Statistics

- **Lines of Code:** ~2,500
- **Test Cases:** 100+
- **Test Coverage:** 80.7%
- **Binary Size:** ~2.8MB (compressed)
- **Platforms:** 5 (Linux, macOS, Windows - amd64/arm64)

## 🔧 Technical Details

### Dependencies

- Go 1.25+
- gopkg.in/yaml.v3 (YAML serialization)
- github.com/stretchr/testify (testing)

### Performance

- **Startup time:** <100ms
- **Memory usage:** ~10-20MB (idle)
- **Concurrent operations:** Thread-safe
- **Storage:** File-based (YAML) or in-memory

### Compatibility

- ✅ MCP Protocol: 2024-11-05
- ✅ JSON-RPC: 2.0
- ✅ Go: 1.25+
- ✅ Docker: Multi-stage build
- ✅ OS: Linux, macOS, Windows

## 📚 Documentation

- [README](README.md) - Project overview
- [Tools Reference](docs/TOOLS.md) - Complete tool documentation
- [Examples](examples/) - Usage examples
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues
- [Architecture](docs/plano/ARCHITECTURE.md) - System design
- [Claude Desktop Setup](examples/integration/claude_desktop_setup.md)

## 🔒 Security

- ✅ Input validation
- ✅ Safe YAML parsing
- ✅ No external network calls
- ✅ Secure file operations
- ✅ Non-root Docker user

## ⚠️ Known Limitations

1. **Single-instance only** - No distributed mode yet
2. **No database backend** - File or memory storage only
3. **No authentication** - Intended for local use with Claude Desktop
4. **Element limits** - Recommended max 1000 elements per type
5. **Unicode in YAML** - Some special characters may need escaping

## 🛣️ Roadmap

See [ROADMAP.md](docs/next_steps/ROADMAP.md) for future plans.

**v0.2.0 (Planned - Q1 2026):**
- GitHub synchronization
- Advanced search (NLP)
- More element types
- REST API
- WebSocket support

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](LICENSE) file.

## 🙏 Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) - Protocol specification
- [DollhouseMCP](https://github.com/DollhouseMCP/mcp-server) - Inspiration
- Go community for excellent tools and libraries

## 📞 Support

- **Issues:** https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions:** https://github.com/fsvxavier/nexs-mcp/discussions
- **Documentation:** https://github.com/fsvxavier/nexs-mcp/tree/main/docs

---

**Full Changelog:** https://github.com/fsvxavier/nexs-mcp/compare/...v0.1.0

**Verified with:**
- Go 1.25.0
- Docker 24.0.7
- Claude Desktop 1.0.0
