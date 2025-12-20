#!/usr/bin/env node

/**
 * Post-install script for @nexs-mcp/server
 * 
 * This script downloads or verifies the appropriate binary for the current platform.
 * Binaries are embedded in the package during publish.
 */

const fs = require('fs');
const path = require('path');

const platform = process.platform;
const arch = process.arch;

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
  console.error(`❌ Unsupported platform: ${platform}-${arch}`);
  console.error('Supported platforms:', Object.keys(binaryMap).join(', '));
  process.exit(1);
}

const binaryPath = path.join(__dirname, '..', 'bin', binaryName);

// Check if binary exists
if (fs.existsSync(binaryPath)) {
  console.log(`✅ NEXS MCP Server binary found: ${binaryName}`);
  
  // Make executable on Unix systems
  if (platform !== 'win32') {
    try {
      fs.chmodSync(binaryPath, '755');
      console.log(`✅ Binary marked as executable`);
    } catch (err) {
      console.warn(`⚠️  Warning: Could not set executable permission: ${err.message}`);
    }
  }
  
  console.log(`\n🎉 Installation complete!`);
  console.log(`\nTo use: npx nexs-mcp [command] [options]`);
  console.log(`Example: npx nexs-mcp --version\n`);
} else {
  console.error(`❌ Binary not found: ${binaryPath}`);
  console.error(`\nPlease report this issue at:`);
  console.error(`https://github.com/fsvxavier/nexs-mcp/issues\n`);
  process.exit(1);
}
