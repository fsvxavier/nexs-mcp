#!/bin/bash
# Complete CRUD workflow example

set -e

echo "=== NEXS MCP Complete Workflow Example ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Step 1: Initialize connection${NC}"
cat << 'EOF' | ./bin/nexs-mcp > /tmp/nexs_init.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
EOF

echo -e "${GREEN}✓ Connection initialized${NC}"
echo ""

echo -e "${BLUE}Step 2: Create a new Skill element${NC}"
cat << 'EOF' | ./bin/nexs-mcp > /tmp/nexs_create.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_element","arguments":{"type":"skill","name":"Code Review","description":"Expert code review and feedback","version":"1.0.0","author":"Workflow Demo","tags":["code","review","quality"],"is_active":true}}}
EOF

ELEMENT_ID=$(cat /tmp/nexs_create.json | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Created element with ID: $ELEMENT_ID${NC}"
echo ""

echo -e "${BLUE}Step 3: List all elements${NC}"
cat << 'EOF' | ./bin/nexs-mcp > /tmp/nexs_list.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_elements","arguments":{}}}
EOF

TOTAL=$(cat /tmp/nexs_list.json | grep -o '"total":[0-9]*' | head -1 | cut -d':' -f2)
echo -e "${GREEN}✓ Found $TOTAL total elements${NC}"
echo ""

echo -e "${BLUE}Step 4: Get specific element${NC}"
cat << EOF | ./bin/nexs-mcp > /tmp/nexs_get.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_element","arguments":{"id":"$ELEMENT_ID"}}}
EOF

echo -e "${GREEN}✓ Retrieved element details${NC}"
echo ""

echo -e "${BLUE}Step 5: Update element${NC}"
cat << EOF | ./bin/nexs-mcp > /tmp/nexs_update.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"update_element","arguments":{"id":"$ELEMENT_ID","description":"Advanced code review with best practices","tags":["code","review","quality","best-practices"]}}}
EOF

echo -e "${GREEN}✓ Updated element${NC}"
echo ""

echo -e "${BLUE}Step 6: Delete element${NC}"
cat << EOF | ./bin/nexs-mcp > /tmp/nexs_delete.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"workflow-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"delete_element","arguments":{"id":"$ELEMENT_ID"}}}
EOF

echo -e "${GREEN}✓ Deleted element${NC}"
echo ""

echo -e "${GREEN}=== Workflow completed successfully! ===${NC}"
echo ""
echo "Output files saved in /tmp/nexs_*.json"

# Cleanup
rm -f /tmp/nexs_*.json
