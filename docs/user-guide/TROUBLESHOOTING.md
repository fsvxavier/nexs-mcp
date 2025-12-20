# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with NEXS-MCP.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Connection Problems](#connection-problems)
- [Element Operations](#element-operations)
- [GitHub Integration](#github-integration)
- [Performance Issues](#performance-issues)
- [Storage and Data](#storage-and-data)
- [Logging and Debugging](#logging-and-debugging)
- [FAQ](#faq)

---

## Installation Issues

### Problem: `go install` fails with "not found"

**Symptoms:**
```bash
$ go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest
go: github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest: not found
```

**Solutions:**

1. **Check Go version:**
   ```bash
   go version  # Should be 1.21 or later
   ```

2. **Update Go if needed:**
   ```bash
   # Download from https://go.dev/dl/
   # Or use your package manager:
   brew upgrade go  # macOS
   sudo apt upgrade golang-go  # Ubuntu
   ```

3. **Clear Go cache:**
   ```bash
   go clean -modcache
   go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest
   ```

4. **Check network/proxy:**
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest
   ```

### Problem: Binary not in PATH

**Symptoms:**
```bash
$ nexs-mcp
bash: nexs-mcp: command not found
```

**Solution:**

Add Go bin directory to PATH:

```bash
# Find where Go installs binaries
go env GOPATH

# Add to shell profile (.bashrc, .zshrc, etc.)
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

### Problem: Permission denied

**Symptoms:**
```bash
$ nexs-mcp
bash: /usr/local/bin/nexs-mcp: Permission denied
```

**Solution:**
```bash
chmod +x $(which nexs-mcp)
```

---

## Connection Problems

### Problem: Claude can't find NEXS-MCP server

**Symptoms:**
- Claude says "MCP server not available"
- No NEXS-MCP tools appear in Claude

**Solutions:**

1. **Check configuration file location:**
   
   **macOS:**
   ```bash
   ls -la ~/Library/Application\ Support/Claude/claude_desktop_config.json
   ```
   
   **Linux:**
   ```bash
   ls -la ~/.config/Claude/claude_desktop_config.json
   ```

2. **Verify configuration syntax:**
   ```bash
   # macOS
   cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | jq .
   
   # Linux
   cat ~/.config/Claude/claude_desktop_config.json | jq .
   ```
   
   Should not show JSON errors.

3. **Test server standalone:**
   ```bash
   # Run directly
   nexs-mcp
   
   # Should output:
   # NEXS MCP Server v1.0.0
   # Registered 55 tools
   # Server ready. Listening on stdio...
   ```

4. **Check Claude logs:**
   - Claude Desktop > View > Developer > Developer Tools
   - Look for connection errors
   - Search for "nexs-mcp" in console

5. **Restart Claude:**
   - Completely quit Claude Desktop (not just close window)
   - Relaunch
   - Wait 10 seconds for servers to initialize

### Problem: Server starts but tools don't work

**Symptoms:**
- Server shows as connected
- Tools appear but fail when called

**Solutions:**

1. **Check data directory permissions:**
   ```bash
   ls -la data/elements/
   # Should show directories for each element type
   
   # Fix permissions if needed:
   chmod -R 755 data/
   ```

2. **Enable debug logging:**
   
   Update `claude_desktop_config.json`:
   ```json
   {
     "mcpServers": {
       "nexs-mcp": {
         "command": "nexs-mcp",
         "args": ["-log-level", "debug"]
       }
     }
   }
   ```

3. **Check for zombie processes:**
   ```bash
   ps aux | grep nexs-mcp
   # Kill any stale processes:
   pkill nexs-mcp
   ```

---

## Element Operations

### Problem: "Element not found"

**Symptoms:**
```
Error: element persona-xyz-001 not found
```

**Solutions:**

1. **List all elements:**
   ```
   In Claude: "List all my elements"
   ```

2. **Search for element:**
   ```
   "Search for XYZ"
   "Find elements with name XYZ"
   ```

3. **Check file directly:**
   ```bash
   ls data/elements/personas/
   cat data/elements/personas/persona-*.yaml
   ```

4. **Verify element ID:**
   - Element IDs are case-sensitive
   - Format: `{type}-{name}-{number}`
   - Example: `persona-tech-expert-001`

### Problem: "Invalid element" error

**Symptoms:**
```
Error: validation failed for element
```

**Solutions:**

1. **Validate the element:**
   ```
   "Validate element [element-id]"
   ```

2. **Check required fields:**
   
   Every element needs:
   ```yaml
   id: unique-id
   type: persona|skill|template|agent|memory|ensemble
   name: Element Name
   description: A description
   ```

3. **Inspect YAML syntax:**
   ```bash
   yamllint data/elements/*/your-element.yaml
   ```

4. **Fix manually:**
   ```bash
   # Edit the file
   nano data/elements/personas/persona-xyz-001.yaml
   
   # Verify with validation
   "Validate element persona-xyz-001"
   ```

### Problem: Cannot create element with same name

**Symptoms:**
```
Error: element with similar name already exists
```

**Solutions:**

1. **List existing elements:**
   ```
   "List all [type]"
   ```

2. **Use different name:**
   ```
   "Create persona Tech Expert 2"
   ```

3. **Delete old element first:**
   ```
   "Delete element [old-element-id]"
   "Create persona Tech Expert"
   ```

4. **Use update instead:**
   ```
   "Update persona Tech Expert to add expertise in Rust"
   ```

### Problem: Template rendering fails

**Symptoms:**
```
Error: template rendering failed
Missing variable: [variable-name]
```

**Solutions:**

1. **Check template variables:**
   ```
   "Get template [template-id]"
   # Lists all required variables
   ```

2. **Provide all variables:**
   ```
   "Instantiate template with:
   - var1: value1
   - var2: value2
   - var3: value3"
   ```

3. **Validate template syntax:**
   ```
   "Validate template [template-id]"
   ```

4. **Test with sample data:**
   ```
   "Validate template [template-id] with test data"
   ```

---

## GitHub Integration

### Problem: GitHub authentication fails

**Symptoms:**
```
Error: authentication failed
Error: device code expired
```

**Solutions:**

1. **Check authentication status:**
   ```
   "Check GitHub authentication status"
   ```

2. **Re-authenticate:**
   ```
   "Start GitHub authentication"
   # Follow the prompts
   # Complete within 15 minutes
   ```

3. **Verify token:**
   ```bash
   # Token stored in:
   ~/.nexs-mcp/auth/github_token.enc
   
   # Check permissions:
   ls -la ~/.nexs-mcp/auth/
   # Should be 600 (rw-------)
   ```

4. **Manual token (advanced):**
   ```bash
   # Generate at https://github.com/settings/tokens
   # Scopes needed: repo, read:user
   export GITHUB_TOKEN=ghp_your_token_here
   ```

### Problem: Cannot push to GitHub

**Symptoms:**
```
Error: repository not found
Error: permission denied
```

**Solutions:**

1. **Verify repository exists:**
   ```
   "List my GitHub repositories"
   ```

2. **Check permissions:**
   - Repository must be owned by you or you must have write access
   - For organizations: Enable OAuth app access

3. **Create repository:**
   ```bash
   # Via GitHub CLI
   gh repo create my-nexs-portfolio --private
   
   # Or in Claude:
   "Create GitHub repository my-nexs-portfolio"
   ```

4. **Check authentication:**
   ```
   "Check GitHub authentication status"
   ```

### Problem: Sync conflicts

**Symptoms:**
```
Warning: conflicts detected during sync
Conflicting elements: [list]
```

**Solutions:**

1. **Review conflicts:**
   ```
   "Show sync conflicts"
   ```

2. **Choose resolution strategy:**
   ```
   "Resolve conflicts using local-wins"
   "Resolve conflicts using remote-wins"
   "Resolve conflicts using newest-wins"
   ```

3. **Manual resolution:**
   - Pull conflicted elements
   - Review changes
   - Keep desired version
   - Push again

---

## Performance Issues

### Problem: Server is slow

**Symptoms:**
- Operations take >5 seconds
- Claude response delays
- Timeouts

**Solutions:**

1. **Check system resources:**
   ```bash
   # CPU usage
   top | grep nexs-mcp
   
   # Memory usage
   ps aux | grep nexs-mcp
   ```

2. **Reduce portfolio size:**
   ```
   # Delete old memories
   "Delete memories older than 90 days"
   
   # Archive inactive elements
   "Backup portfolio"
   "Delete inactive elements"
   ```

3. **Clear cache:**
   ```
   "Clear collection cache"
   ```

4. **Use in-memory storage for testing:**
   ```bash
   nexs-mcp -storage memory
   ```

5. **Enable performance monitoring:**
   ```
   "Show performance dashboard"
   ```

### Problem: High memory usage

**Symptoms:**
```bash
$ ps aux | grep nexs-mcp
nexs-mcp  1234  5.2  15.0 ...  # >15% memory
```

**Solutions:**

1. **Restart server:**
   - Quit Claude Desktop
   - Wait 10 seconds
   - Relaunch

2. **Reduce in-memory index:**
   - Index size grows with elements
   - Consider backup + restore to clean

3. **Switch to file storage:**
   ```json
   {
     "mcpServers": {
       "nexs-mcp": {
         "command": "nexs-mcp",
         "args": ["-storage", "file"]
       }
     }
   }
   ```

---

## Storage and Data

### Problem: Data directory not created

**Symptoms:**
```
Error: failed to create data directory
Error: permission denied: data/elements
```

**Solutions:**

1. **Check permissions:**
   ```bash
   ls -la $(pwd)
   # Current directory must be writable
   ```

2. **Create manually:**
   ```bash
   mkdir -p data/elements/{agents,personas,skills,templates,memories,ensembles}
   chmod -R 755 data/
   ```

3. **Use custom location:**
   ```bash
   nexs-mcp -data-dir ~/my-nexs-data
   ```

4. **Check disk space:**
   ```bash
   df -h .
   ```

### Problem: Corrupted element files

**Symptoms:**
```
Error: failed to parse element
Error: invalid YAML
```

**Solutions:**

1. **Validate YAML:**
   ```bash
   yamllint data/elements/*/element-xyz.yaml
   ```

2. **Restore from backup:**
   ```
   "Restore portfolio from backup [filename]"
   ```

3. **Fix manually:**
   ```bash
   # Make backup
   cp element-xyz.yaml element-xyz.yaml.bak
   
   # Edit with proper YAML editor
   nano element-xyz.yaml
   
   # Validate
   yamllint element-xyz.yaml
   ```

4. **Rebuild from git:**
   ```bash
   git pull origin main
   ```

### Problem: Lost elements after restart

**Symptoms:**
- Elements disappear after restarting
- Portfolio is empty

**Causes & Solutions:**

1. **Using in-memory storage:**
   ```
   # Check server config
   "List all elements"
   
   # Switch to file storage
   Update claude_desktop_config.json to use -storage file
   ```

2. **Wrong data directory:**
   ```bash
   # Find actual directory
   find ~ -name "persona-*.yaml" 2>/dev/null
   
   # Update config to point there
   ```

3. **Deleted accidentally:**
   ```
   # Restore from backup
   "Restore portfolio from backup"
   
   # Or pull from GitHub
   "Pull latest changes from GitHub"
   ```

---

## Logging and Debugging

### Enable Debug Logging

**In claude_desktop_config.json:**
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": [
        "-log-level", "debug",
        "-log-format", "json"
      ]
    }
  }
}
```

### View Logs

**Claude Desktop Logs:**
1. Claude Desktop > View > Developer > Developer Tools
2. Go to Console tab
3. Filter for "nexs-mcp"

**Server Logs (if running standalone):**
```bash
nexs-mcp -log-level debug 2>&1 | tee nexs-mcp.log
```

**Query Structured Logs:**
```
"Show logs from today"
"Show error logs"
"Show logs for create_persona tool"
"Show logs with keyword 'authentication'"
```

### Common Log Patterns

**Authentication issues:**
```
ERROR: token validation failed
ERROR: GitHub API rate limit exceeded
```

**Permission issues:**
```
ERROR: failed to write element
ERROR: permission denied
```

**Validation issues:**
```
WARN: element validation failed
ERROR: required field missing
```

### Debug Checklist

When reporting issues, include:

1. **Version:**
   ```bash
   nexs-mcp --version
   ```

2. **Configuration:**
   ```bash
   cat claude_desktop_config.json
   ```

3. **Logs:**
   ```
   "Show logs from today filtered by errors"
   ```

4. **System info:**
   ```bash
   uname -a
   go version
   ```

5. **Reproduction steps:**
   - What you tried
   - Expected result
   - Actual result

---

## FAQ

### Q: How many elements can I have?

**A:** No hard limit. Tested with 10,000+ elements. Performance may degrade above 50,000.

### Q: Can I edit elements manually?

**A:** Yes! Elements are YAML files. Edit with any text editor. Validate after:
```
"Validate element [element-id]"
```

### Q: Does NEXS-MCP work offline?

**A:** Yes, except:
- GitHub integration (requires internet)
- Collection installation (requires internet)
- Everything else works offline

### Q: Can multiple Claude instances share a portfolio?

**A:** No. File locking prevents corruption. Use GitHub sync for multi-user scenarios.

### Q: How do I migrate from in-memory to file storage?

**A:**
1. Create backup while using memory storage
2. Switch to file storage
3. Restore backup
```
"Create backup"
# Update config to file storage
"Restore from backup"
```

### Q: Can I use NEXS-MCP without Claude?

**A:** Yes! Run standalone:
```bash
# Start server
nexs-mcp

# In another terminal, send MCP commands via stdio
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | nexs-mcp
```

### Q: How do I reset everything?

**A:**
```bash
# Backup first!
"Create backup"

# Delete all elements
rm -rf data/elements/*

# Or start fresh
rm -rf data/
nexs-mcp
```

### Q: What's the difference between Skill and Agent?

**A:**
- **Skill**: Passive capability (checklist, procedure)
- **Agent**: Active executor (autonomous, goal-driven)

### Q: Can I nest ensembles?

**A:** Yes! An ensemble can include agents that are themselves ensembles.

### Q: How do I contribute elements to collections?

**A:**
```
"Publish collection to GitHub"
```
Creates PR to nexs-mcp-registry repository.

---

## Getting Help

### Community Support

1. **GitHub Discussions:** https://github.com/fsvxavier/nexs-mcp/discussions
   - Ask questions
   - Share experiences
   - Feature requests

2. **GitHub Issues:** https://github.com/fsvxavier/nexs-mcp/issues
   - Report bugs
   - Track known issues
   - Submit feature requests

### Documentation

- [Getting Started](./GETTING_STARTED.md)
- [Quick Start](./QUICK_START.md)
- [Element Types](../elements/README.md)
- [MCP Tools Reference](../api/MCP_TOOLS.md)

### Debug Commands

```
# System info
"Show usage statistics"
"Display performance dashboard"
"Get capability index stats"

# Health checks
"List all elements"
"Check GitHub authentication status"
"Show logs from today"

# Validation
"Validate element [id]"
"Validate template [id]"
```

---

## Error Code Reference

### E001: Element Not Found
**Cause:** Element ID doesn't exist  
**Fix:** Use `list_elements` to find correct ID

### E002: Validation Failed
**Cause:** Element doesn't meet schema requirements  
**Fix:** Use `validate_element` to see specific issues

### E003: Permission Denied
**Cause:** Insufficient file system permissions  
**Fix:** Check directory permissions with `ls -la`

### E004: GitHub Auth Failed
**Cause:** Token invalid or expired  
**Fix:** Re-authenticate with `github_auth_start`

### E005: Template Error
**Cause:** Invalid template syntax or missing variables  
**Fix:** Use `validate_template` and check variables

### E006: Storage Error
**Cause:** Cannot read/write to storage  
**Fix:** Check disk space and permissions

### E007: Network Error
**Cause:** Cannot reach GitHub API  
**Fix:** Check internet connection and firewall

---

## Emergency Recovery

### Complete Reset

```bash
# 1. Backup everything
"Create backup of portfolio"

# 2. Note your GitHub connection
"Check GitHub authentication status"

# 3. Stop Claude
# Quit Claude Desktop

# 4. Clean data
rm -rf data/
rm -rf ~/.nexs-mcp/

# 5. Restart
# Launch Claude Desktop

# 6. Restore
"Restore from backup [filename]"
```

### Recover from Git

```bash
# 1. Clone your portfolio repo
git clone https://github.com/username/portfolio.git

# 2. Copy to NEXS data directory
cp -r portfolio/* data/elements/

# 3. Restart Claude Desktop
```

---

**Still stuck?** Open an issue: https://github.com/fsvxavier/nexs-mcp/issues/new
