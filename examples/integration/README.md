# Integration Examples

Examples for integrating NEXS MCP with various tools and platforms.

## Claude Desktop Integration

The most common integration is with Claude Desktop. See `claude_desktop_config.json` for configuration.

### Setup

1. Locate your Claude Desktop config file:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

2. Add NEXS MCP configuration (see `claude_desktop_config.json`)

3. Restart Claude Desktop

4. Verify NEXS MCP tools are available in Claude

### Configuration Options

- **command**: Path to nexs-mcp binary
- **args**: Command-line arguments (optional)
- **env**: Environment variables for configuration

See `claude_desktop_config.json` for a complete example.

## Other Integrations

Coming soon:
- Python client example
- REST API wrapper
- VS Code extension
