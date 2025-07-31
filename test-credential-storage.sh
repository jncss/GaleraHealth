#!/bin/bash

# Test script to demonstrate per-node credential storage
# This script shows how GaleraHealth saves SSH users and passwords

echo "=== GaleraHealth Credential Storage Demo ==="
echo ""

echo "📋 Current configuration:"
if [ -f ~/.galerahealth ]; then
    echo "Configuration file exists at ~/.galerahealth"
    echo "File permissions: $(ls -la ~/.galerahealth | awk '{print $1}')"
    echo ""
    echo "Current saved credentials:"
    cat ~/.galerahealth | jq '.' 2>/dev/null || cat ~/.galerahealth
else
    echo "No configuration file found"
fi

echo ""
echo "🔐 How SSH credentials are stored:"
echo "1. SSH usernames: Stored in plain text per node"
echo "2. SSH passwords: Encrypted with AES-GCM using node-specific keys"
echo "3. SSH key usage: Tracked per node (uses_ssh_keys: true/false)"
echo "4. Per-node isolation: Each node can have different credentials"

echo ""
echo "🏗️  Configuration structure:"
echo "node_credentials: [
  {
    \"node_ip\": \"10.1.1.91\",
    \"ssh_username\": \"root\",
    \"mysql_username\": \"galera_user\",
    \"encrypted_ssh_password\": \"AES_encrypted_data...\",
    \"encrypted_mysql_password\": \"AES_encrypted_data...\",
    \"has_ssh_password\": true,
    \"has_mysql_password\": true,
    \"uses_ssh_keys\": false
  },
  {
    \"node_ip\": \"10.1.1.92\",
    \"ssh_username\": \"admin\",
    \"uses_ssh_keys\": true
  }
]"

echo ""
echo "🔒 Security features:"
echo "• Passwords encrypted per node with different keys"
echo "• SHA-256 key derivation from node IP"
echo "• AES-256-GCM authenticated encryption"
echo "• Configuration file has restrictive permissions (600)"

echo ""
echo "✨ Usage examples:"
echo "• Different SSH users per node: ✅"
echo "• Mixed auth methods (keys + passwords): ✅" 
echo "• Encrypted password storage: ✅"
echo "• Automatic credential reuse: ✅"
