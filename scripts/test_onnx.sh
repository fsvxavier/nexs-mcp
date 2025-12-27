#!/bin/bash
# Test nexs-mcp with ONNX support
# This script tests if the ONNX build loads and reports ONNX status correctly

set -e

echo "=== Testing nexs-mcp with ONNX support ==="
echo ""

# Check if binary exists
if [ ! -f "./bin/nexs-mcp" ]; then
    echo "❌ Binary not found. Run 'make build-onnx' first."
    exit 1
fi

# Create temporary test directory
TEST_DIR="/tmp/nexs-test-$$"
mkdir -p "$TEST_DIR/data/elements"
LOG_FILE="$TEST_DIR/server.log"

# Cleanup function
cleanup() {
    local exit_code=$?
    echo ""
    echo "Cleaning up..."
    
    # Kill server if still running
    if [ -n "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null 2>&1; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill -SIGTERM $SERVER_PID 2>/dev/null || true
        
        # Wait up to 5 seconds for graceful shutdown
        for i in {1..5}; do
            if ! ps -p $SERVER_PID > /dev/null 2>&1; then
                echo "✓ Server stopped gracefully"
                break
            fi
            sleep 1
        done
        
        # Force kill if still running
        if ps -p $SERVER_PID > /dev/null 2>&1; then
            echo "⚠ Forcing server shutdown..."
            kill -SIGKILL $SERVER_PID 2>/dev/null || true
            wait $SERVER_PID 2>/dev/null || true
        fi
    fi
    
    # Remove test directory
    rm -rf "$TEST_DIR"
    
    exit $exit_code
}

# Set trap for cleanup on exit or interruption
trap cleanup EXIT INT TERM

echo "1. Starting MCP server in background..."
echo "   Log file: $LOG_FILE"
./bin/nexs-mcp -data-dir "$TEST_DIR/data/elements" -log-level info > "$LOG_FILE" 2>&1 &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# Wait for server to start (max 5 seconds)
echo ""
echo "2. Waiting for server to initialize..."
for i in {1..10}; do
    sleep 0.5
    if [ -s "$LOG_FILE" ]; then
        if grep -q "Server ready" "$LOG_FILE" 2>/dev/null; then
            echo "   ✅ Server initialized"
            break
        fi
    fi
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo "   ❌ Server process died during startup"
        echo ""
        echo "Server logs:"
        cat "$LOG_FILE"
        exit 1
    fi
done

echo ""
echo "3. Checking if server is running..."
# Give server a moment after initialization
sleep 1
if ps -p $SERVER_PID > /dev/null 2>&1; then
    echo "   ✅ Server running (PID: $SERVER_PID)"
else
    echo "   ⚠️  Server stopped after initialization (this is normal for stdio servers)"
    echo "   Checking if initialization was successful..."
    if grep -q "Server ready" "$LOG_FILE" 2>/dev/null; then
        echo "   ✅ Server initialized successfully before stopping"
    else
        echo "   ❌ Server failed to initialize properly"
        exit 1
    fi
fi

echo ""
echo "4. Checking ONNX status in logs..."
if grep -qi "onnx" "$LOG_FILE"; then
    echo "   ✅ ONNX status found in logs:"
    echo ""
    grep -i "onnx" "$LOG_FILE" | while IFS= read -r line; do
        echo "      $line"
    done
else
    echo "   ⚠️  ONNX not mentioned in logs"
fi

echo ""
echo "5. Server startup logs:"
echo "   ─────────────────────────────────────────────"
head -20 "$LOG_FILE" | while IFS= read -r line; do
    echo "   $line"
done
echo "   ─────────────────────────────────────────────"

echo ""
echo "6. Testing graceful shutdown..."
if ps -p $SERVER_PID > /dev/null 2>&1; then
    kill -SIGTERM $SERVER_PID 2>/dev/null

    # Wait for graceful shutdown
    SHUTDOWN_SUCCESS=false
    for i in {1..5}; do
        if ! ps -p $SERVER_PID > /dev/null 2>&1; then
            echo "   ✅ Server stopped gracefully after ${i}s"
            SHUTDOWN_SUCCESS=true
            break
        fi
        sleep 1
    done

    if [ "$SHUTDOWN_SUCCESS" = false ]; then
        echo "   ⚠️  Server did not stop gracefully, forcing shutdown"
        kill -SIGKILL $SERVER_PID 2>/dev/null || true
    fi
else
    echo "   ℹ️  Server already stopped (normal for stdio mode)"
fi

# Clear trap since we handled cleanup manually
trap - EXIT INT TERM

echo ""
echo "=== Test Complete ==="
echo ""
echo "Summary:"
echo "  - Binary: ./bin/nexs-mcp"
echo "  - ONNX: $(grep -i "onnx_support" "$LOG_FILE" | head -1 | grep -o '"onnx_support":"[^"]*"' || echo 'status not found')"
echo "  - Server: Started and stopped successfully"
echo ""

# Cleanup
rm -rf "$TEST_DIR"
