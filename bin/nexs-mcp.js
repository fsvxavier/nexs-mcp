#!/usr/bin/env node

/**
 * NEXS MCP Server - NPM wrapper
 * 
 * This script wraps the Go binary and provides a Node.js interface
 * for easy integration with Claude Desktop and other MCP clients.
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Detect platform and architecture
const platform = process.platform;
const arch = process.arch;

// Map Node.js platform/arch to binary names
const binaryMap = {
  'darwin-x64': 'nexs-mcp-darwin-amd64',
  'darwin-arm64': 'nexs-mcp-darwin-arm64',
  'linux-x64': 'nexs-mcp-linux-amd64',
  'linux-arm64': 'nexs-mcp-linux-arm64',
  'win32-x64': 'nexs-mcp-windows-amd64.exe',
  'win32-arm64': 'nexs-mcp-windows-arm64.exe'
};

const binaryKey = `${platform}-${arch}`;
const binaryName = binaryMap[binaryKey];

if (!binaryName) {
  console.error(`Unsupported platform: ${platform}-${arch}`);
  console.error('Supported platforms:', Object.keys(binaryMap).join(', '));
  process.exit(1);
}

// Locate the binary
const binaryPath = path.join(__dirname, '..', 'bin', binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(`Binary not found: ${binaryPath}`);
  console.error('Please run: npm install');
  process.exit(1);
}

// Ensure binary is executable (Unix systems)
if (platform !== 'win32') {
  try {
    fs.chmodSync(binaryPath, '755');
  } catch (err) {
    console.warn(`Warning: Could not set executable permission: ${err.message}`);
  }
}

// Spawn the Go binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  env: process.env
});

// Forward exit code
child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code || 0);
  }
});

// Handle errors
child.on('error', (err) => {
  console.error(`Failed to start NEXS MCP Server: ${err.message}`);
  process.exit(1);
});
