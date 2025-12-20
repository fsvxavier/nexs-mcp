#!/bin/bash

# Example: List elements with filtering using NEXS MCP

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Listing Elements ===${NC}\n"

# Example 1: List all elements
echo -e "${YELLOW}Example 1: List all elements${NC}"
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {}
  },
  "id": 1
}
EOF
)

echo "$REQUEST" | ../../bin/nexs-mcp | jq '.result.content[0].text | fromjson'
echo ""

# Example 2: Filter by type
echo -e "${YELLOW}Example 2: Filter by type (persona)${NC}"
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "type": "persona"
    }
  },
  "id": 2
}
EOF
)

echo "$REQUEST" | ../../bin/nexs-mcp | jq '.result.content[0].text | fromjson'
echo ""

# Example 3: Filter by active status
echo -e "${YELLOW}Example 3: Filter by active status${NC}"
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "is_active": true
    }
  },
  "id": 3
}
EOF
)

echo "$REQUEST" | ../../bin/nexs-mcp | jq '.result.content[0].text | fromjson'
echo ""

# Example 4: Filter by tags
echo -e "${YELLOW}Example 4: Filter by tags${NC}"
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "tags": ["engineer"]
    }
  },
  "id": 4
}
EOF
)

echo "$REQUEST" | ../../bin/nexs-mcp | jq '.result.content[0].text | fromjson'
echo ""

# Example 5: Pagination
echo -e "${YELLOW}Example 5: Pagination (limit 5, offset 0)${NC}"
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "limit": 5,
      "offset": 0
    }
  },
  "id": 5
}
EOF
)

echo "$REQUEST" | ../../bin/nexs-mcp | jq '.result.content[0].text | fromjson'
echo -e "\n${GREEN}✓ Examples completed${NC}"
