/**
 * NEXS MCP Server - NPM Package Entry Point
 * 
 * This is the main entry point for the @nexs-mcp/server package.
 * It re-exports the binary path for programmatic use.
 */

const path = require('path');
const fs = require('fs');

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
  throw new Error(`Unsupported platform: ${platform}-${arch}`);
}

const binaryPath = path.join(__dirname, 'bin', binaryName);

if (!fs.existsSync(binaryPath)) {
  throw new Error(`Binary not found: ${binaryPath}. Please run: npm install`);
}

module.exports = {
  binaryPath,
  binaryName,
  platform,
  arch
};
