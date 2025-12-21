# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The NEXS-MCP team takes security bugs seriously. We appreciate your efforts to
responsibly disclose your findings, and will make every effort to acknowledge
your contributions.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**[fsvxavier@gmail.com](mailto:fsvxavier@gmail.com)**

Include the following information in your report:

- Type of vulnerability (e.g., injection, XSS, authentication bypass, etc.)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

After you submit a report, we will:

1. **Acknowledge receipt** within 48 hours
2. **Provide an initial assessment** within 5 business days
3. **Keep you informed** of the progress toward fixing the vulnerability
4. **Notify you** when the vulnerability is fixed

### Our Commitment

We commit to:

- Work with you to understand and resolve the issue quickly
- Keep you informed throughout the process
- Credit you for the discovery (unless you prefer to remain anonymous)
- Coordinate public disclosure timing with you

## Security Best Practices

When using NEXS-MCP, we recommend:

### Token and Credential Management

- **Never commit tokens or credentials** to version control
- Store GitHub OAuth tokens encrypted (NEXS-MCP does this automatically)
- Use environment variables or secure configuration files for sensitive data
- Regularly rotate access tokens
- Use the minimum required permissions for tokens

### File Permissions

- Ensure `~/.nexs-mcp/auth/` has restricted permissions (0700)
- Token files should be readable only by the owner (0600)
- Protect configuration files containing sensitive data

### Network Security

- Use HTTPS for all GitHub API communications (enforced by default)
- Verify SSL certificates (enabled by default)
- Be cautious when adding external collection sources

### Docker Security

If running NEXS-MCP in Docker:

- Run as non-root user (default in our Docker image)
- Use read-only volumes where possible
- Enable security options (no-new-privileges, etc.)
- Keep the Docker image updated

### Element Security

When creating or using elements:

- **Review elements** before importing from external collections
- Be cautious with elements that:
  - Request sensitive information
  - Execute system commands
  - Access external APIs
  - Modify critical files
- Validate element sources
- Check element permissions and capabilities

### GitHub Integration Security

- Use OAuth tokens instead of personal access tokens when possible
- Review repository permissions before syncing
- Be cautious when forking repositories
- Verify PR submissions before creating them
- Don't sync portfolios containing sensitive information to public repositories

## Known Security Considerations

### Token Storage

NEXS-MCP encrypts GitHub OAuth tokens using AES-256-GCM with:
- PBKDF2 key derivation (100,000 iterations)
- Machine-specific salt
- Secure random nonces

However, this protects against unauthorized file access, not against:
- Malicious processes running under your user account
- Root/admin access to your machine
- Memory dumps while NEXS-MCP is running

### Element Execution

Elements (especially Skills and Agents) can execute arbitrary logic. While we
validate element structure, we cannot verify the safety of element behavior.
Always review elements from untrusted sources.

### MCP Protocol

The MCP protocol is designed for AI assistant integration. When integrated with
Claude Desktop or similar tools:
- The AI can invoke any MCP tool
- Consider the security implications of exposed tools
- Be cautious with tools that modify data or execute commands

## Vulnerability Disclosure Policy

When we receive a security report:

1. **Confirmation**: We confirm the vulnerability and determine its severity
2. **Patching**: We develop and test a fix
3. **Release**: We release a patched version
4. **Disclosure**: We publish a security advisory

### Public Disclosure Timeline

- **Critical vulnerabilities**: Fix released within 7 days, public disclosure after 30 days
- **High severity**: Fix released within 30 days, public disclosure after 60 days
- **Medium/Low severity**: Fix included in next regular release

We may adjust this timeline based on:
- Complexity of the fix
- Testing requirements
- Coordination with downstream users

## Security Updates

Subscribe to security updates:

- Watch this repository for security advisories
- Check [GitHub Security Advisories](https://github.com/fsvxavier/nexs-mcp/security/advisories)
- Follow release notes for security patches

## Bug Bounty Program

We currently do not have a bug bounty program. However, we deeply appreciate
security researchers who help us keep NEXS-MCP secure, and we will:

- Publicly acknowledge your contribution (with your permission)
- Give you credit in release notes
- Consider featuring your work in project communications

## Legal

We request that you:

- Give us reasonable time to address issues before public disclosure
- Make a good faith effort to avoid privacy violations, data destruction, and service interruption
- Do not access or modify data that doesn't belong to you
- Contact us immediately if you inadvertently access sensitive data

We will not pursue legal action against researchers who:

- Follow this policy
- Report vulnerabilities responsibly
- Don't exploit vulnerabilities beyond what's necessary to demonstrate the issue

## Questions?

If you have questions about this security policy, please contact:
[fsvxavier@gmail.com](mailto:fsvxavier@gmail.com)

---

**Last Updated:** December 20, 2025
