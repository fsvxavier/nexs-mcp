#!/usr/bin/env python3
"""
Python MCP Client Example for NEXS MCP Server

Demonstrates how to interact with NEXS MCP Server using Python.
Requires: Python 3.8+

Usage:
    python3 python_client.py
"""

import json
import subprocess
import sys
from typing import Dict, Any, Optional


class NexsMCPClient:
    """Simple MCP client for interacting with NEXS MCP Server."""
    
    def __init__(self, server_path: str = "./bin/nexs-mcp"):
        self.server_path = server_path
        self.request_id = 0
        
    def _next_id(self) -> int:
        """Get next request ID."""
        self.request_id += 1
        return self.request_id
    
    def _send_request(self, method: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Send a JSON-RPC request to the server."""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_id(),
            "method": method,
            "params": params or {}
        }
        
        # For notification methods (no response expected)
        if method.startswith("notifications/"):
            del request["id"]
        
        return request
    
    def initialize(self) -> None:
        """Initialize MCP session."""
        init_request = self._send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "python-example",
                "version": "1.0.0"
            }
        })
        
        initialized_notification = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized"
        }
        
        return [init_request, initialized_notification]
    
    def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Call an MCP tool."""
        return self._send_request("tools/call", {
            "name": tool_name,
            "arguments": arguments
        })
    
    def list_tools(self) -> Dict[str, Any]:
        """List available tools."""
        return self._send_request("tools/list")


def main():
    """Run example workflow."""
    client = NexsMCPClient()
    
    print("=== NEXS MCP Python Client Example ===\n")
    
    # Initialize
    print("1. Initializing connection...")
    init_requests = client.initialize()
    
    # Create a persona
    print("2. Creating a new Persona element...")
    create_request = client.call_tool("create_element", {
        "type": "persona",
        "name": "AI Research Scientist",
        "description": "Expert in machine learning and neural networks",
        "version": "1.0.0",
        "author": "Python Example",
        "tags": ["ai", "research", "ml"],
        "is_active": True
    })
    
    # List elements
    print("3. Listing all elements...")
    list_request = client.call_tool("list_elements", {})
    
    # List tools
    print("4. Listing available tools...")
    tools_request = client.list_tools()
    
    # Combine all requests
    all_requests = init_requests + [create_request, list_request, tools_request]
    
    # Send to server
    print("\n5. Sending requests to server...")
    
    input_data = "\n".join(json.dumps(req) for req in all_requests)
    
    try:
        result = subprocess.run(
            [client.server_path],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=10
        )
        
        print("\n=== Server Responses ===\n")
        
        # Parse and display responses
        for line in result.stdout.strip().split('\n'):
            if line:
                try:
                    response = json.loads(line)
                    print(json.dumps(response, indent=2))
                    print()
                except json.JSONDecodeError:
                    print(f"Invalid JSON: {line}")
        
        if result.stderr:
            print("=== Server Errors ===")
            print(result.stderr)
        
        print("\n=== Example completed successfully! ===")
        
    except subprocess.TimeoutExpired:
        print("ERROR: Server timeout")
        sys.exit(1)
    except FileNotFoundError:
        print(f"ERROR: Server not found at {client.server_path}")
        print("Build the server first: make build")
        sys.exit(1)
    except Exception as e:
        print(f"ERROR: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
