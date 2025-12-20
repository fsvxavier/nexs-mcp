#!/bin/bash
# Example: Create a Code Review Skill using NEXS MCP

echo "Creating Code Review Skill..."

cat << 'EOF'
{
  "tool": "quick_create_skill",
  "arguments": {
    "name": "Code Review",
    "description": "Comprehensive code review with best practices",
    "triggers": [
      "code review",
      "review pr",
      "check code"
    ],
    "procedure": "1. Check code style and formatting\n2. Verify logic and error handling\n3. Look for security issues\n4. Suggest improvements and optimizations\n5. Provide constructive feedback",
    "active": true
  }
}
EOF

echo ""
echo "Skill created! Use 'list_elements.sh' to verify."
