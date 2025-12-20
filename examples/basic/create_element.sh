#!/bin/bash

# Example: Create a new persona element using NEXS MCP

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Creating Persona Element ===${NC}\n"

# Create element request
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "create_element",
    "arguments": {
      "type": "persona",
      "name": "Senior Software Engineer",
      "description": "Experienced Go developer specializing in distributed systems and microservices",
      "version": "1.0.0",
      "author": "NEXS Team",
      "tags": ["engineer", "golang", "backend"]
    }
  },
  "id": 1
}
EOF
)

echo "Request:"
echo "$REQUEST" | jq '.'
echo ""

# Send request to MCP server
echo -e "${BLUE}Sending request to MCP server...${NC}\n"
RESPONSE=$(echo "$REQUEST" | ../../bin/nexs-mcp)

echo "Response:"
echo "$RESPONSE" | jq '.'
echo ""

# Check if successful
if echo "$RESPONSE" | jq -e '.result' > /dev/null 2>&1; then
    ELEMENT_ID=$(echo "$RESPONSE" | jq -r '.result.content[0].text' | jq -r '.id')
    echo -e "${GREEN}✓ Element created successfully!${NC}"
    echo -e "Element ID: ${GREEN}$ELEMENT_ID${NC}"
else
    echo "✗ Failed to create element"
    echo "$RESPONSE" | jq '.error'
    exit 1
fi
