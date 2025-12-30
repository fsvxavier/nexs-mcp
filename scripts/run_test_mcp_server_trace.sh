#!/usr/bin/env bash
set -euo pipefail

# Wrapper to run test_mcp_server.sh under bash -x with a timeout
# Usage: TIMEOUT=30s ./scripts/run_test_mcp_server_trace.sh

TIMEOUT=${TIMEOUT:-60s}

if ! command -v timeout >/dev/null 2>&1; then
  echo "Error: 'timeout' command not found. Install coreutils (Linux) or gtimeout via coreutils on macOS (brew install coreutils)." >&2
  exit 1
fi

echo "Running: timeout ${TIMEOUT} bash -x test_mcp_server.sh"
exec timeout "${TIMEOUT}" bash -x test_mcp_server.sh
