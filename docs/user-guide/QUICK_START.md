# Quick Start Guide

Get productive with NEXS-MCP in 5 minutes! This guide provides hands-on tutorials for the most common tasks.

## Prerequisites

- NEXS-MCP installed and running
- Claude Desktop configured (see [Getting Started](./GETTING_STARTED.md))

## Tutorial 1: Create Your First Persona (2 minutes)

### Goal
Create a "Senior Developer" persona that you can use for code reviews and technical discussions.

### Steps

**In Claude Desktop, say:**
```
Create a persona called "Senior Developer" who is experienced in Python, Go, and cloud architecture. 
They should be analytical, detail-oriented, and have a mentoring communication style.
```

**What happens:**
- Claude uses `quick_create_persona` tool
- NEXS-MCP creates the element
- File saved to `data/elements/personas/persona-senior-developer-001.yaml`

**Verify it worked:**
```
List all my personas
```

**Expected output:**
```
Found 1 persona:
- Senior Developer (persona-senior-developer-001)
  Expert in: Python, Go, Cloud Architecture
  Traits: Analytical, Detail-oriented
  Active: Yes
```

### Use Your Persona

```
As Senior Developer, review this Python code:
[paste your code]
```

Claude will adopt the Senior Developer persona's expertise and communication style.

---

## Tutorial 2: Create a Code Review Skill (3 minutes)

### Goal
Build a reusable skill for performing code reviews with a consistent checklist.

### Steps

**Create the skill:**
```
Create a skill called "Code Review Checklist" that includes these steps:
1. Check code formatting and style
2. Look for security vulnerabilities
3. Verify error handling
4. Check for performance issues
5. Validate test coverage

Category: code-quality
Dependencies: linting tools
```

**What Claude does:**
- Uses `create_skill` tool
- Structures the procedures
- Saves to `data/elements/skills/skill-code-review-checklist-001.yaml`

**Use the skill:**
```
Apply the Code Review Checklist skill to this code:
[paste code]
```

**Expected behavior:**
Claude will systematically work through each step of your checklist.

---

## Tutorial 3: Generate Documents with Templates (3 minutes)

### Goal
Create a bug report template that you can reuse with different values.

### Steps

**Create the template:**
```
Create a template called "Bug Report" with these variables:
- bug_id: The bug number
- title: Short description
- severity: critical, high, medium, low
- steps: How to reproduce
- expected: What should happen
- actual: What actually happens
```

**Instantiate the template:**
```
Use the Bug Report template with:
- bug_id: BUG-123
- title: Login fails with special characters
- severity: high
- steps: 1. Go to login page 2. Enter username with @ symbol 3. Click login
- expected: User logs in successfully
- actual: Error "Invalid username format"
```

**Result:**
```markdown
# Bug Report: BUG-123

## Login fails with special characters

**Severity:** HIGH

### Steps to Reproduce
1. Go to login page
2. Enter username with @ symbol
3. Click login

### Expected Behavior
User logs in successfully

### Actual Behavior
Error "Invalid username format"
```

**Save as element (optional):**
```
Save that bug report as a memory for future reference
```

---

## Tutorial 4: Search and Filter Elements (2 minutes)

### Goal
Find specific elements in your growing portfolio.

### By Type

```
Show me all my personas
Show me all my skills
List all templates
```

### By Tag

```
Find all elements tagged with "security"
Search for elements about "API"
```

### By Status

```
List all active elements
Show inactive personas
```

### Full-Text Search

```
Search my portfolio for "authentication"
Find elements mentioning "database"
```

**Pro Tip:** Use semantic search for better results:
```
Find capabilities related to code quality
```

---

## Tutorial 5: Backup Your Portfolio (1 minute)

### Create Backup

```
Create a backup of my portfolio
```

**What happens:**
- NEXS-MCP creates `nexs-portfolio-backup-YYYYMMDD-HHMMSS.tar.gz`
- Includes all elements
- Adds SHA-256 checksums
- Compresses for storage

**Verify:**
```bash
# In terminal
ls -lh nexs-portfolio-backup-*.tar.gz
```

### Restore Backup

```
Restore my portfolio from backup nexs-portfolio-backup-20251220-103000.tar.gz
```

**Options:**
- `merge_strategy: skip` - Keep existing elements
- `merge_strategy: overwrite` - Replace existing elements
- `merge_strategy: rename` - Create new IDs for conflicts

---

## Tutorial 6: Memory Management (3 minutes)

### Auto-Save Conversations

NEXS-MCP can automatically save conversation context:

```
Enable auto-save for our conversation
```

Now every significant exchange is saved as a memory.

### Manual Memory Creation

```
Remember that we deployed version 2.1.0 to production with these features:
- New user dashboard
- Performance improvements
- Bug fixes for login
```

### Search Memories

```
What did we discuss about deployments?
Find memories from last week
Search memories about "performance"
```

### Memory Summary

```
Summarize all my memories
Show memory statistics
```

---

## Tutorial 7: GitHub Integration (5 minutes)

### Initial Setup

**Start authentication:**
```
Start GitHub authentication
```

**Follow the prompts:**
1. Visit the URL provided
2. Enter the device code
3. Authorize the application

**Verify:**
```
Check GitHub authentication status
```

### Push to GitHub

```
Sync my portfolio to GitHub repository "my-nexs-portfolio"
```

**What happens:**
- Creates repository if it doesn't exist
- Commits all elements
- Pushes to `main` branch

### Pull from GitHub

```
Pull latest changes from my GitHub portfolio
```

**Handles conflicts:**
- Shows conflicts if found
- Prompts for resolution strategy
- Applies your choice

---

## Tutorial 8: Collections (4 minutes)

### Browse Collections

```
Show me available collections
```

### Search Collections

```
Search for collections about "DevOps"
Find collections with security personas
```

### Install Collection

```
Install collection from github://fsvxavier/nexs-collection-starter
```

**What installs:**
- 5 starter personas
- 10 common skills
- 5 useful templates
- Installation report

### Use Collection Elements

```
List elements from the starter collection
Activate the "DevOps Engineer" persona from starter collection
```

---

## Tutorial 9: Ensembles for Complex Tasks (5 minutes)

### Create an Ensemble

```
Create an ensemble called "Code Review Team" with:
- security-expert persona
- performance-analyst persona
- style-checker persona

Execution mode: parallel
Aggregation: consensus
```

### Execute Ensemble

```
Execute the Code Review Team ensemble on this code:
[paste code]
```

**What happens:**
1. All three personas analyze the code simultaneously
2. Each provides their perspective
3. Results are aggregated by consensus
4. Final report combines all insights

### Check Status

```
Get status of Code Review Team ensemble
```

---

## Tutorial 10: Analytics and Monitoring (3 minutes)

### Usage Statistics

```
Show usage statistics for the last 30 days
```

**See:**
- Most used tools
- Success rates
- Call frequency
- Top 10 tools

### Performance Dashboard

```
Display performance dashboard
```

**See:**
- P50, P95, P99 latencies
- Slow operations
- Tool performance
- System health

### Query Logs

```
Show logs from today filtered by errors
List logs for the backup_portfolio tool
```

---

## Common Patterns

### Pattern 1: Role-Based Work

```
1. "Create a persona for Data Scientist"
2. "Activate Data Scientist"
3. "Analyze this dataset: [data]"
4. "Generate a report using the Analysis Report template"
5. "Save the analysis as a memory"
```

### Pattern 2: Document Generation

```
1. "Create template for Meeting Notes"
2. "Instantiate Meeting Notes with today's standup data"
3. "Save as memory with tag 'standup'"
4. "Search for all standup memories from this week"
```

### Pattern 3: Multi-Agent Workflow

```
1. "Create ensemble with 3 specialized personas"
2. "Execute ensemble on complex task"
3. "Review aggregated results"
4. "Save insights as memories"
```

### Pattern 4: Portfolio Management

```
1. "List all elements"
2. "Deactivate unused personas"
3. "Delete old memories from 90 days ago"
4. "Create backup"
5. "Push to GitHub"
```

---

## Quick Reference Commands

### Creation
- `Create a persona called "X"`
- `Create a skill for "Y"`
- `Create a template with variables A, B, C`
- `Create an agent to do "Z"`

### Management
- `List all [type]`
- `Activate element [id/name]`
- `Deactivate element [id/name]`
- `Delete element [id/name]`
- `Update [element] to change [field]`

### Search
- `Search for [query]`
- `Find elements tagged [tag]`
- `Show active elements`
- `List elements of type [type]`

### Backup
- `Create backup`
- `Restore from backup [filename]`

### GitHub
- `Start GitHub auth`
- `Push to GitHub`
- `Pull from GitHub`

### Collections
- `Browse collections`
- `Install collection [uri]`
- `List installed collections`

### Analytics
- `Show usage stats`
- `Display performance dashboard`
- `Show logs`

---

## Pro Tips

### 1. Use Quick Create for Speed

```
# Instead of specifying everything:
Quick create persona "DevOps Engineer"

# Instead of full skill definition:
Quick create skill "Database Backup"
```

### 2. Tag Everything

```
Create a persona tagged with "backend", "python", "api"
```

Later:
```
Find all elements tagged "backend"
```

### 3. Leverage Memory Search

```
# Specific search
Search memories for "deployment" from last week

# Semantic search
Find memories related to production issues
```

### 4. Use Templates for Consistency

Create templates for:
- Meeting notes
- Bug reports
- Code review checklists
- Project documentation
- API specifications

### 5. Build Ensembles for Quality

```
Create ensembles for:
- Code reviews (security + performance + style)
- Design reviews (UX + technical + business)
- Content reviews (grammar + tone + accuracy)
```

---

## Troubleshooting Common Issues

### Issue: "Element not found"
```
# Check spelling
List all elements

# Use partial match
Search for "dev"
```

### Issue: "Invalid template"
```
# Validate first
Validate template [template-id]

# Check syntax
Get template [template-id]
```

### Issue: "GitHub auth failed"
```
# Check status
Check GitHub authentication status

# Re-authenticate
Start GitHub authentication again
```

---

## Next Steps

1. **[Element Types Guide](../elements/README.md)** - Deep dive into each type
2. **[MCP Tools Reference](../api/MCP_TOOLS.md)** - Complete tool documentation
3. **[Troubleshooting](./TROUBLESHOOTING.md)** - Detailed problem solving
4. **[Advanced Examples](../../examples/)** - Real-world use cases

---

**Questions?** Check the [FAQ](./TROUBLESHOOTING.md#faq) or ask in [Discussions](https://github.com/fsvxavier/nexs-mcp/discussions)!
