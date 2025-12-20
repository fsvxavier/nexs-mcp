# frozen_string_literal: true

# NEXS-MCP Formula
class NexsMcp < Formula
  desc "NEXS MCP Server - Model Context Protocol server for AI portfolio management"
  homepage "https://github.com/fsvxavier/nexs-mcp"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/fsvxavier/nexs-mcp/releases/download/v1.0.0/nexs-mcp-darwin-arm64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_ARM64"
    else
      url "https://github.com/fsvxavier/nexs-mcp/releases/download/v1.0.0/nexs-mcp-darwin-amd64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/fsvxavier/nexs-mcp/releases/download/v1.0.0/nexs-mcp-linux-arm64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_LINUX_ARM64"
    else
      url "https://github.com/fsvxavier/nexs-mcp/releases/download/v1.0.0/nexs-mcp-linux-amd64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_LINUX_AMD64"
    end
  end

  def install
    # Determine the correct binary name based on OS and architecture
    if OS.mac?
      if Hardware::CPU.arm?
        binary_name = "nexs-mcp-darwin-arm64"
      else
        binary_name = "nexs-mcp-darwin-amd64"
      end
    else # Linux
      if Hardware::CPU.arm?
        binary_name = "nexs-mcp-linux-arm64"
      else
        binary_name = "nexs-mcp-linux-amd64"
      end
    end

    # Install binary
    bin.install binary_name => "nexs-mcp"

    # Create data directory
    (var/"nexs-mcp/data").mkpath
    (var/"nexs-mcp/data/agents").mkpath
    (var/"nexs-mcp/data/personas").mkpath
    (var/"nexs-mcp/data/skills").mkpath
    (var/"nexs-mcp/data/templates").mkpath
    (var/"nexs-mcp/data/memories").mkpath
    (var/"nexs-mcp/data/ensembles").mkpath

    # Create config directory
    (etc/"nexs-mcp").mkpath
  end

  def post_install
    # Set proper permissions
    chmod 0755, bin/"nexs-mcp"
    
    # Create ~/.nexs-mcp directory for auth tokens
    (ENV["HOME"] + "/.nexs-mcp").mkpath
    (ENV["HOME"] + "/.nexs-mcp/auth").mkpath
    
    ohai "NEXS-MCP installed successfully!"
    ohai "Data directory: #{var}/nexs-mcp/data"
    ohai "Config directory: #{etc}/nexs-mcp"
    ohai "Run 'nexs-mcp --help' to get started"
  end

  test do
    # Test that binary runs and shows version
    assert_match version.to_s, shell_output("#{bin}/nexs-mcp --version")
    
    # Test help command
    assert_match "NEXS MCP Server", shell_output("#{bin}/nexs-mcp --help")
  end

  def caveats
    <<~EOS
      NEXS-MCP has been installed!

      Data directory: #{var}/nexs-mcp/data
      Config directory: #{etc}/nexs-mcp
      Auth directory: ~/.nexs-mcp/auth

      To integrate with Claude Desktop, add this to your Claude config:
      
      {
        "mcpServers": {
          "nexs-mcp": {
            "command": "#{bin}/nexs-mcp"
          }
        }
      }

      Configuration file location:
      - macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
      - Linux: ~/.config/Claude/claude_desktop_config.json

      For more information, visit:
      https://github.com/fsvxavier/nexs-mcp
    EOS
  end
end
