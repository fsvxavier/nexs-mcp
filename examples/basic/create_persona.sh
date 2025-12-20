#!/bin/bash
# Example: Create a Technical Writer Persona using NEXS MCP

echo "Creating Technical Writer Persona..."

# This example demonstrates how to create a persona via MCP
# In practice, this would be sent via the MCP protocol to the server

cat << 'EOF'
{
  "tool": "quick_create_persona",
  "arguments": {
    "name": "Technical Writer",
    "description": "Expert in writing clear and concise technical documentation",
    "expertise": [
      "technical writing",
      "API documentation",
      "user guides",
      "markdown"
    ],
    "traits": [
      "clear",
      "concise",
      "detail-oriented",
      "user-focused"
    ],
    "communication_style": "professional and accessible",
    "active": true
  }
}
EOF

echo ""
echo "Persona created! Use 'list_elements.sh' to verify."
