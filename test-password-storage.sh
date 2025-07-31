#!/bin/bash

# Test script to force password authentication
# Temporarily renames SSH keys to force password prompt

echo "=== Testing SSH Password Storage ==="
echo ""

# Backup SSH keys temporarily
echo "🔐 Temporarily moving SSH keys to force password authentication..."
BACKUP_DIR="/tmp/ssh_keys_backup_$$"
mkdir -p "$BACKUP_DIR"

if [ -d ~/.ssh ]; then
    if [ -f ~/.ssh/id_rsa ]; then
        mv ~/.ssh/id_rsa "$BACKUP_DIR/"
        echo "   Moved id_rsa"
    fi
    if [ -f ~/.ssh/id_ecdsa ]; then
        mv ~/.ssh/id_ecdsa "$BACKUP_DIR/"
        echo "   Moved id_ecdsa"
    fi
    if [ -f ~/.ssh/id_ed25519 ]; then
        mv ~/.ssh/id_ed25519 "$BACKUP_DIR/"
        echo "   Moved id_ed25519"
    fi
fi

echo ""
echo "🧪 Testing GaleraHealth with password authentication..."
echo "   (SSH keys temporarily disabled)"
echo ""

# Function to restore SSH keys
restore_keys() {
    echo ""
    echo "🔄 Restoring SSH keys..."
    if [ -d "$BACKUP_DIR" ]; then
        if [ -f "$BACKUP_DIR/id_rsa" ]; then
            mv "$BACKUP_DIR/id_rsa" ~/.ssh/
            echo "   Restored id_rsa"
        fi
        if [ -f "$BACKUP_DIR/id_ecdsa" ]; then
            mv "$BACKUP_DIR/id_ecdsa" ~/.ssh/
            echo "   Restored id_ecdsa"
        fi
        if [ -f "$BACKUP_DIR/id_ed25519" ]; then
            mv "$BACKUP_DIR/id_ed25519" ~/.ssh/
            echo "   Restored id_ed25519"
        fi
        rmdir "$BACKUP_DIR"
    fi
    echo "✓ SSH keys restored"
}

# Set trap to restore keys on script exit
trap restore_keys EXIT

# Test with password
echo "Enter a test node IP (e.g., 10.1.1.92): "
read TEST_NODE

if [ -n "$TEST_NODE" ]; then
    echo ""
    echo "🚀 Running GaleraHealth with forced password authentication..."
    echo "   Node: $TEST_NODE"
    echo ""
    echo "📝 This will test:"
    echo "   1. SSH password prompt (since keys are disabled)"
    echo "   2. Password encryption and storage"
    echo "   3. Password reuse on second run"
    echo ""
    
    # First run - should prompt for password
    echo "=== FIRST RUN (should prompt for password) ==="
    echo "$TEST_NODE" | ./galerahealth -vvv
    
    echo ""
    echo "=== Configuration after first run ==="
    if [ -f ~/.galerahealth ]; then
        cat ~/.galerahealth | jq '.' 2>/dev/null || cat ~/.galerahealth
    fi
    
    echo ""
    echo "=== SECOND RUN (should use saved password) ==="
    echo "$TEST_NODE" | ./galerahealth -vvv
    
else
    echo "No test node specified, skipping test"
fi
