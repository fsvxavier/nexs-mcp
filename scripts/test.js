#!/usr/bin/env node

/**
 * Test runner for @nexs-mcp/server
 * 
 * Runs basic functionality tests to verify the installation.
 */

const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

console.log('🧪 Running NEXS MCP Server tests...\n');

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
const binaryPath = path.join(__dirname, '..', 'bin', binaryName);

let testsPassed = 0;
let testsFailed = 0;

function test(name, fn) {
  try {
    fn();
    console.log(`✅ ${name}`);
    testsPassed++;
  } catch (err) {
    console.error(`❌ ${name}`);
    console.error(`   ${err.message}`);
    testsFailed++;
  }
}

// Test 1: Binary exists
test('Binary exists', () => {
  if (!fs.existsSync(binaryPath)) {
    throw new Error(`Binary not found at ${binaryPath}`);
  }
});

// Test 2: Binary is executable
test('Binary is executable', () => {
  try {
    const result = spawnSync(binaryPath, ['--help'], { timeout: 5000 });
    if (result.error) {
      throw result.error;
    }
  } catch (err) {
    throw new Error(`Binary execution failed: ${err.message}`);
  }
});

// Test 3: Help command works
test('Help command works', () => {
  const result = spawnSync(binaryPath, ['--help'], { encoding: 'utf8', timeout: 5000 });
  if (result.status !== 0) {
    throw new Error(`Help command failed with exit code ${result.status}`);
  }
  if (!result.stdout || result.stdout.length === 0) {
    throw new Error('Help command produced no output');
  }
});

// Test 4: Version command works
test('Version command works', () => {
  const result = spawnSync(binaryPath, ['--version'], { encoding: 'utf8', timeout: 5000 });
  // Version command might exit with 0 or 1, both acceptable
  if (result.status > 1) {
    throw new Error(`Version command failed with exit code ${result.status}`);
  }
});

// Print summary
console.log(`\n${'='.repeat(50)}`);
console.log(`Tests passed: ${testsPassed}`);
console.log(`Tests failed: ${testsFailed}`);
console.log(`${'='.repeat(50)}\n`);

process.exit(testsFailed > 0 ? 1 : 0);
