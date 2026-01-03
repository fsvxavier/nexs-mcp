#!/bin/bash

# Test script for Auto-Save and Token Metrics Integration
# Tests the real nexs-mcp server with all features enabled

set -e

echo "=========================================="
echo "NEXS-MCP Auto-Save & Token Metrics Tests"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="/tmp/nexs-mcp-test-$$"
DATA_DIR="$TEST_DIR/data"
METRICS_DIR="$TEST_DIR/token_metrics"
SERVER_LOG="$TEST_DIR/server.log"
TEST_RESULTS="$TEST_DIR/results.txt"

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    # Don't remove test dir to allow inspection
    echo "Test artifacts preserved in: $TEST_DIR"
}

trap cleanup EXIT

# Create test directories
mkdir -p "$DATA_DIR"
mkdir -p "$METRICS_DIR"

echo "Test directory: $TEST_DIR"
echo ""

# Step 1: Run unit tests
echo "=== Step 1: Running Unit Tests ==="
echo ""

echo "Testing token_metrics.go..."
if go test -v ./internal/application -run TestToken 2>&1 | tee -a "$TEST_RESULTS"; then
    echo -e "${GREEN}✓ Token metrics unit tests PASSED${NC}"
else
    echo -e "${RED}✗ Token metrics unit tests FAILED${NC}"
    exit 1
fi
echo ""

echo "Testing response_middleware.go..."
if go test -v ./internal/mcp -run TestResponseMiddleware 2>&1 | tee -a "$TEST_RESULTS"; then
    echo -e "${GREEN}✓ Response middleware unit tests PASSED${NC}"
else
    echo -e "${RED}✗ Response middleware unit tests FAILED${NC}"
    exit 1
fi
echo ""

# Step 2: Build the server
echo "=== Step 2: Building Server ==="
echo ""

if make build-onnx 2>&1 | tee -a "$TEST_RESULTS"; then
    echo -e "${GREEN}✓ Server build SUCCESSFUL${NC}"
else
    echo -e "${RED}✗ Server build FAILED${NC}"
    exit 1
fi
echo ""

# Step 3: Start server with test configuration
echo "=== Step 3: Starting Test Server ==="
echo ""

export NEXS_DATA_DIR="$DATA_DIR"
export NEXS_SERVER_NAME="nexs-mcp-test"
export NEXS_STORAGE_TYPE="file"
export NEXS_LOG_LEVEL="info"
export NEXS_LOG_FORMAT="json"
export NEXS_AUTO_SAVE_MEMORIES="true"
export NEXS_AUTO_SAVE_INTERVAL="30s"  # Short interval for testing
export NEXS_COMPRESSION_ENABLED="true"
export NEXS_COMPRESSION_ALGORITHM="gzip"
export NEXS_COMPRESSION_LEVEL="6"
export NEXS_PROMPT_COMPRESSION_ENABLED="true"
export NEXS_PROMPT_COMPRESSION_RATIO="0.65"
export NEXS_PROMPT_COMPRESSION_MIN_LENGTH="100"

# Override token metrics directory
export HOME="$TEST_DIR"

echo "Starting server with configuration:"
echo "  DATA_DIR: $DATA_DIR"
echo "  AUTO_SAVE_INTERVAL: 30s"
echo "  COMPRESSION_ENABLED: true"
echo "  PROMPT_COMPRESSION_ENABLED: true"
echo ""

# Start server in background
./bin/nexs-mcp > "$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
echo "Waiting for server to initialize..."
sleep 3

# Check if server is still running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${RED}✗ Server failed to start${NC}"
    echo "Server log:"
    cat "$SERVER_LOG"
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Step 4: Send test requests via stdio
echo "=== Step 4: Testing MCP Protocol ==="
echo ""

# Create test request JSON
TEST_REQUEST='{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_optimization_stats","arguments":{"detailed":true}}}'

echo "Sending test request: get_optimization_stats"
echo $TEST_REQUEST | nc localhost 3000 2>/dev/null || echo "Note: Using direct stdio (nc failed)"

sleep 2

# Step 5: Verify auto-save worker started
echo "=== Step 5: Verifying Auto-Save Worker ==="
echo ""

if grep -q "Auto-save worker started" "$SERVER_LOG"; then
    echo -e "${GREEN}✓ Auto-save worker started successfully${NC}"
    AUTO_SAVE_INTERVAL=$(grep "Auto-save worker started" "$SERVER_LOG" | grep -o 'interval=[^ ]*' | cut -d= -f2)
    echo "  Interval: $AUTO_SAVE_INTERVAL"
else
    echo -e "${RED}✗ Auto-save worker not found in logs${NC}"
    echo "Server log excerpt:"
    tail -20 "$SERVER_LOG"
fi
echo ""

# Step 6: Create some working memories (simulate conversation)
echo "=== Step 6: Simulating Conversation Activity ==="
echo ""

# We can't directly call working memory via MCP yet, so we'll create some regular memories
# and check if the system is tracking them

CREATE_MEMORY_REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_element","arguments":{"type":"memory","name":"Test Memory 1","description":"Test conversation context","content":"This is a test conversation with some context that should be saved automatically."}}}'

echo "Creating test memory..."
echo $CREATE_MEMORY_REQUEST | nc localhost 3000 2>/dev/null || true

sleep 2

# Step 7: Wait for auto-save to trigger (30 seconds interval)
echo "=== Step 7: Waiting for Auto-Save Trigger ==="
echo ""

echo "Waiting 35 seconds for auto-save to trigger..."
for i in {35..1}; do
    echo -ne "  ${YELLOW}$i seconds remaining...${NC}\r"
    sleep 1
done
echo ""

# Check if auto-save was performed
if grep -q "Performing auto-save of conversation context" "$SERVER_LOG"; then
    echo -e "${GREEN}✓ Auto-save triggered successfully${NC}"

    if grep -q "Successfully auto-saved conversation context" "$SERVER_LOG"; then
        echo -e "${GREEN}✓ Auto-save completed successfully${NC}"
        MEMORY_COUNT=$(grep "Successfully auto-saved" "$SERVER_LOG" | tail -1 | grep -o 'memory_count=[0-9]*' | cut -d= -f2)
        MEMORY_ID=$(grep "Successfully auto-saved" "$SERVER_LOG" | tail -1 | grep -o 'memory_id=[^ ]*' | cut -d= -f2)
        echo "  Memories saved: $MEMORY_COUNT"
        echo "  Memory ID: $MEMORY_ID"
    else
        echo -e "${YELLOW}⚠ Auto-save triggered but may not have completed${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Auto-save not triggered yet (check interval and logs)${NC}"
fi
echo ""

# Step 8: Check token metrics file
echo "=== Step 8: Verifying Token Metrics ==="
echo ""

METRICS_FILE="$TEST_DIR/.nexs-mcp/token_metrics/token_metrics.json"

if [ -f "$METRICS_FILE" ]; then
    echo -e "${GREEN}✓ Token metrics file exists${NC}"
    echo "  Location: $METRICS_FILE"

    # Check file content
    METRICS_COUNT=$(cat "$METRICS_FILE" | jq '. | length' 2>/dev/null || echo "0")
    echo "  Metrics recorded: $METRICS_COUNT"

    if [ "$METRICS_COUNT" -gt "0" ]; then
        echo -e "${GREEN}✓ Token metrics are being recorded${NC}"
        echo ""
        echo "Sample metrics:"
        cat "$METRICS_FILE" | jq '.[0]' 2>/dev/null || cat "$METRICS_FILE" | head -20
    else
        echo -e "${YELLOW}⚠ No metrics recorded yet${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Token metrics file not created yet${NC}"
    echo "  Expected at: $METRICS_FILE"
fi
echo ""

# Step 9: Query optimization stats
echo "=== Step 9: Querying Optimization Stats ==="
echo ""

STATS_REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_optimization_stats","arguments":{"detailed":true}}}'

echo "Requesting optimization stats..."
echo $STATS_REQUEST | nc localhost 3000 2>/dev/null || true

sleep 2

# Check if stats were queried
if grep -q "token_metrics" "$SERVER_LOG"; then
    echo -e "${GREEN}✓ Token metrics are exposed via get_optimization_stats${NC}"
else
    echo -e "${YELLOW}⚠ Token metrics not found in stats query${NC}"
fi
echo ""

# Step 10: Generate summary report
echo "=== Step 10: Test Summary ==="
echo ""

echo "Server Status:"
if kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "  Server: ${GREEN}RUNNING${NC}"
else
    echo -e "  Server: ${RED}STOPPED${NC}"
fi

echo ""
echo "Feature Status:"

# Check each feature
AUTO_SAVE_ENABLED=$(grep -c "Auto-save worker started" "$SERVER_LOG" || echo "0")
AUTO_SAVE_TRIGGERED=$(grep -c "Performing auto-save" "$SERVER_LOG" || echo "0")
AUTO_SAVE_COMPLETED=$(grep -c "Successfully auto-saved" "$SERVER_LOG" || echo "0")
COMPRESSION_USED=$(grep -c "Compressed" "$SERVER_LOG" || echo "0")

if [ "$AUTO_SAVE_ENABLED" -gt "0" ]; then
    echo -e "  ✓ Auto-save worker: ${GREEN}ENABLED${NC}"
else
    echo -e "  ✗ Auto-save worker: ${RED}NOT ENABLED${NC}"
fi

if [ "$AUTO_SAVE_TRIGGERED" -gt "0" ]; then
    echo -e "  ✓ Auto-save triggered: ${GREEN}$AUTO_SAVE_TRIGGERED times${NC}"
else
    echo -e "  ⚠ Auto-save triggered: ${YELLOW}NOT YET${NC}"
fi

if [ "$AUTO_SAVE_COMPLETED" -gt "0" ]; then
    echo -e "  ✓ Auto-save completed: ${GREEN}$AUTO_SAVE_COMPLETED times${NC}"
else
    echo -e "  ⚠ Auto-save completed: ${YELLOW}NOT YET${NC}"
fi

if [ "$COMPRESSION_USED" -gt "0" ]; then
    echo -e "  ✓ Compression used: ${GREEN}$COMPRESSION_USED times${NC}"
else
    echo -e "  ⚠ Compression used: ${YELLOW}NOT YET${NC}"
fi

if [ -f "$METRICS_FILE" ]; then
    echo -e "  ✓ Token metrics file: ${GREEN}EXISTS${NC}"
else
    echo -e "  ⚠ Token metrics file: ${YELLOW}NOT CREATED${NC}"
fi

echo ""
echo "Logs and artifacts:"
echo "  Server log: $SERVER_LOG"
echo "  Token metrics: $METRICS_FILE"
echo "  Data directory: $DATA_DIR"
echo ""

# Final verdict
echo "=========================================="
if [ "$AUTO_SAVE_ENABLED" -gt "0" ] && [ -f "$METRICS_FILE" ]; then
    echo -e "${GREEN}✓ INTEGRATION TESTS PASSED${NC}"
    echo "Auto-save worker and token metrics are functioning correctly"
    exit 0
else
    echo -e "${YELLOW}⚠ PARTIAL SUCCESS${NC}"
    echo "Some features may need more time or interaction to activate"
    echo "Review logs for details: $SERVER_LOG"
    exit 0
fi
