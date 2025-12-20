#!/bin/bash

# Test script for NEXS MCP Server with file persistence

echo "=== Testing NEXS MCP Server File Persistence ==="
echo ""

# Create test data directory
TEST_DIR="test_data_$(date +%s)"
echo "Using test directory: $TEST_DIR"

# Start server with file storage (in background for testing)
echo ""
echo "Starting server with file storage..."
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{"tools":true}},"id":1}' | ./bin/nexs-mcp -storage file -data-dir "$TEST_DIR" &
SERVER_PID=$!

# Give server time to start
sleep 1

# Check if data directory was created
if [ -d "$TEST_DIR" ]; then
    echo "✓ Data directory created successfully"
else
    echo "✗ Data directory not created"
fi

# Kill the test server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# Clean up
rm -rf "$TEST_DIR"

echo ""
echo "=== Test Complete ==="
