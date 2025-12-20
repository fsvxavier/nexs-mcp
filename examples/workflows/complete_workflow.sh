#!/bin/bash
# Complete NEXS MCP Workflow Example
# This demonstrates a full workflow from element creation to ensemble execution

set -e

echo "========================================="
echo "NEXS MCP Complete Workflow Example"
echo "========================================="
echo ""

# Step 1: Create a Persona
echo "Step 1: Creating Technical Writer Persona..."
cat << 'EOF'
{
  "tool": "quick_create_persona",
  "arguments": {
    "name": "Technical Writer",
    "description": "Expert in clear technical documentation",
    "expertise": ["documentation", "API design"],
    "traits": ["clear", "concise"]
  }
}
EOF
echo ""

# Step 2: Create a Skill
echo "Step 2: Creating Code Review Skill..."
cat << 'EOF'
{
  "tool": "quick_create_skill",
  "arguments": {
    "name": "Code Review",
    "description": "Review code for best practices",
    "triggers": ["code review", "review pr"],
    "procedure": "1. Check style\n2. Verify logic\n3. Suggest improvements"
  }
}
EOF
echo ""

# Step 3: Create a Template
echo "Step 3: Creating Documentation Template..."
cat << 'EOF'
{
  "tool": "quick_create_template",
  "arguments": {
    "name": "API Documentation",
    "description": "Template for API endpoint documentation",
    "content": "# {{endpoint}}\n\n## Description\n{{description}}\n\n## Request\n```\n{{request}}\n```\n\n## Response\n```\n{{response}}\n```",
    "variables": ["endpoint", "description", "request", "response"]
  }
}
EOF
echo ""

# Step 4: Create an Ensemble
echo "Step 4: Creating Documentation Ensemble..."
cat << 'EOF'
{
  "tool": "quick_create_ensemble",
  "arguments": {
    "name": "Documentation Team",
    "description": "Multi-agent documentation generation",
    "members": ["persona:technical-writer", "skill:code-review"],
    "execution_mode": "sequential",
    "aggregation_strategy": "merge"
  }
}
EOF
echo ""

# Step 5: List All Elements
echo "Step 5: Listing all created elements..."
cat << 'EOF'
{
  "tool": "list_elements",
  "arguments": {}
}
EOF
echo ""

# Step 6: Backup Portfolio
echo "Step 6: Creating portfolio backup..."
cat << 'EOF'
{
  "tool": "backup_portfolio",
  "arguments": {
    "output_path": "./backup-example.tar.gz",
    "compression": "best",
    "include_inactive": false
  }
}
EOF
echo ""

# Step 7: Get Usage Statistics
echo "Step 7: Viewing usage statistics..."
cat << 'EOF'
{
  "tool": "get_usage_stats",
  "arguments": {
    "period": "7d",
    "include_top_n": 10
  }
}
EOF
echo ""

echo "========================================="
echo "Workflow Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Verify elements were created: list_elements"
echo "2. Execute the ensemble: execute_ensemble"
echo "3. Sync with GitHub: github_sync_push"
echo "4. View performance: get_performance_dashboard"
