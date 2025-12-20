#!/bin/bash
# Example: List all elements using NEXS MCP

echo "Listing all elements..."

cat << 'EOF'
{
  "tool": "list_elements",
  "arguments": {
    "type": "",
    "active_only": false
  }
}
EOF

echo ""
echo "To filter by type, modify the 'type' argument:"
echo "  - persona"
echo "  - skill"
echo "  - template"
echo "  - agent"
echo "  - memory"
echo "  - ensemble"
