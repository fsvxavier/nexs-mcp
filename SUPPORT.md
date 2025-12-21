# Support

Welcome to NEXS-MCP support! This document will help you get the assistance you need.

## Getting Help

### Documentation

Before asking for help, please check our comprehensive documentation:

- **[Getting Started Guide](docs/user-guide/GETTING_STARTED.md)** - Installation and initial setup
- **[Quick Start Guide](docs/user-guide/QUICK_START.md)** - 10 practical tutorials
- **[Troubleshooting Guide](docs/user-guide/TROUBLESHOOTING.md)** - Common issues and solutions
- **[API Reference](docs/api/)** - MCP tools, resources, and CLI documentation
- **[Element Documentation](docs/elements/)** - Persona, Skill, Template, Agent, Memory, Ensemble guides
- **[Architecture Documentation](docs/architecture/)** - System design and architecture
- **[Development Documentation](docs/development/)** - Contributing and development guides

### Quick Links

- 🏠 [Main README](README.md) - Project overview and features
- 📦 [Installation Guide](README.md#installation) - Multiple installation methods
- 🔧 [Configuration](docs/api/CLI.md) - CLI flags and configuration
- 🐛 [Known Issues](https://github.com/fsvxavier/nexs-mcp/issues) - Check existing issues
- 💡 [Examples](examples/) - Code examples and usage patterns

## Support Channels

### 1. GitHub Issues

For bug reports and feature requests:

**[Create an Issue](https://github.com/fsvxavier/nexs-mcp/issues/new/choose)**

We have templates for:
- 🐛 Bug Reports
- ✨ Feature Requests
- ❓ Questions
- 📦 Element Submissions

Please use the appropriate template and provide as much detail as possible.

### 2. GitHub Discussions

For general questions, ideas, and community discussion:

**[Start a Discussion](https://github.com/fsvxavier/nexs-mcp/discussions)**

Categories:
- **General** - General discussion about NEXS-MCP
- **Ideas** - Share and discuss ideas for improvements
- **Q&A** - Ask questions and get answers from the community
- **Show and Tell** - Share your elements and use cases

### 3. Troubleshooting

If you're experiencing issues, follow these steps:

1. **Check the [Troubleshooting Guide](docs/user-guide/TROUBLESHOOTING.md)**
   - Common installation issues
   - Configuration problems
   - Runtime errors
   - Performance issues

2. **Enable debug logging**
   ```bash
   nexs-mcp --log-level debug
   ```

3. **Verify your installation**
   ```bash
   nexs-mcp --version
   nexs-mcp health-check
   ```

4. **Check existing issues**
   - [All Issues](https://github.com/fsvxavier/nexs-mcp/issues)
   - [Bug Reports](https://github.com/fsvxavier/nexs-mcp/labels/bug)

## What to Include in Support Requests

When asking for help, please include:

### System Information
- Operating System and version
- NEXS-MCP version (`nexs-mcp --version`)
- Installation method (Go, Docker, NPM, Homebrew, binary)
- Go version (if running from source)

### Problem Description
- Clear description of the issue
- Expected behavior
- Actual behavior
- Error messages (full output, not screenshots)

### Reproduction Steps
1. Detailed steps to reproduce the issue
2. Configuration files (remove sensitive data)
3. Relevant log output with `--log-level debug`
4. MCP tool calls that cause the issue

### Example
```markdown
**Environment:**
- OS: macOS 14.1 (Apple Silicon)
- NEXS-MCP: v1.0.0
- Installation: Homebrew

**Issue:**
Getting "connection refused" when trying to sync portfolio with GitHub.

**Steps to Reproduce:**
1. Run `nexs-mcp` with Claude Desktop
2. Call MCP tool `github_sync_push`
3. See error: "connection refused"

**Logs:**
```shell
2025-12-20T10:30:45.123Z DEBUG Attempting GitHub sync...
2025-12-20T10:30:45.456Z ERROR connection refused
```

**Configuration:**
```json
{
  "log_level": "debug",
  "data_dir": "~/.nexs-mcp/data"
}
```
```

## Response Times

We aim to respond to:
- **Critical bugs**: Within 48 hours
- **Bug reports**: Within 1 week
- **Feature requests**: Within 2 weeks
- **Questions**: Within 1 week

*Note: These are goals, not guarantees. Response times may vary based on complexity and maintainer availability.*

## Self-Service Resources

### Common Issues

#### Installation Issues
- [Installation Guide](README.md#installation)
- [Troubleshooting: Installation](docs/user-guide/TROUBLESHOOTING.md#installation-issues)

#### Configuration Issues
- [CLI Reference](docs/api/CLI.md)
- [Configuration Examples](docs/api/CLI.md#configuration-examples)

#### GitHub Integration Issues
- [GitHub OAuth Setup](docs/user-guide/GETTING_STARTED.md#github-integration)
- [Troubleshooting: GitHub](docs/user-guide/TROUBLESHOOTING.md#github-integration)

#### Claude Desktop Integration
- [Integration Guide](examples/integration/claude_desktop_setup.md)
- [Configuration Example](examples/integration/claude_desktop_config.json)

#### Element Issues
- [Element Documentation](docs/elements/README.md)
- [Creating Elements](docs/development/ADDING_ELEMENT_TYPE.md)
- [Element Validation](docs/api/MCP_TOOLS.md#validation-tools)

### Video Tutorials

*Coming soon: Video tutorials and walkthroughs*

### Blog Posts

*Coming soon: Technical blog posts and deep dives*

## Community

Join our community:

- ⭐ [Star us on GitHub](https://github.com/fsvxavier/nexs-mcp)
- 👁️ [Watch for updates](https://github.com/fsvxavier/nexs-mcp/subscription)
- 🍴 [Fork and contribute](CONTRIBUTING.md)
- 💬 [Join discussions](https://github.com/fsvxavier/nexs-mcp/discussions)

## Contributing

Want to help improve NEXS-MCP?

- 📖 Read our [Contributing Guide](CONTRIBUTING.md)
- 🐛 [Report bugs](https://github.com/fsvxavier/nexs-mcp/issues/new?template=bug_report.yml)
- 💡 [Suggest features](https://github.com/fsvxavier/nexs-mcp/issues/new?template=feature_request.yml)
- 📦 [Submit elements](https://github.com/fsvxavier/nexs-mcp/issues/new?template=element_submission.yml)
- 🔧 [Submit pull requests](https://github.com/fsvxavier/nexs-mcp/pulls)

## Security Issues

For security vulnerabilities, please **do not** use public GitHub issues.

See our [Security Policy](SECURITY.md) for how to report security issues responsibly.

## Commercial Support

Currently, we do not offer commercial support. For enterprise needs or custom
development, please contact [fsvxavier@gmail.com](mailto:fsvxavier@gmail.com).

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md).
By participating in this project you agree to abide by its terms.

## Feedback

We value your feedback! Let us know:
- What's working well
- What could be improved
- What features you'd like to see
- How you're using NEXS-MCP

Share your feedback in [GitHub Discussions](https://github.com/fsvxavier/nexs-mcp/discussions).

---

**Thank you for using NEXS-MCP!** 🚀

If you find this project helpful, please consider:
- ⭐ Starring the repository
- 📢 Sharing with others
- 🤝 Contributing to the project
- 📦 Submitting your elements to the collection

---

**Last Updated:** December 20, 2025
