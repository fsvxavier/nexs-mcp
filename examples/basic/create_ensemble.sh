#!/bin/bash
# Example: Create a Documentation Team Ensemble using NEXS MCP

echo "Creating Documentation Team Ensemble..."

cat << 'EOF'
{
  "tool": "quick_create_ensemble",
  "arguments": {
    "name": "Documentation Team",
    "description": "Multi-agent ensemble for generating comprehensive documentation",
    "members": [
      "persona:technical-writer",
      "skill:code-review"
    ],
    "execution_mode": "sequential",
    "aggregation_strategy": "merge",
    "active": true
  }
}
EOF

echo ""
echo "Ensemble created! This ensemble will:"
echo "1. Use the Technical Writer persona for documentation"
echo "2. Apply Code Review skill for quality"
echo "3. Execute sequentially (one after another)"
echo "4. Merge results from all members"
