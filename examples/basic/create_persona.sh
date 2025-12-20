#!/bin/bash
# Example: Create a new Persona element using MCP tools/call

cat << 'EOF' | ./bin/nexs-mcp
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"example-client","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_element","arguments":{"type":"persona","name":"Senior Developer","description":"Experienced software engineer with 10+ years","version":"1.0.0","author":"NEXS Team","tags":["development","senior","technical"],"is_active":true}}}
EOF
