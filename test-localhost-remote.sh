#!/bin/bash

# Test script for localhost + remote SSH functionality

echo "=== Testing Localhost + Remote SSH Functionality ==="
echo ""

# Clear previous config for clean test
echo "🗑️  Clearing previous configuration..."
./galerahealth --clear-config

echo ""
echo "🧪 Test 1: Using localhost with cluster coherence analysis"
echo "   This should now connect to remote nodes via SSH"
echo ""

# Create input for the test
cat > /tmp/test_input << 'EOF'
localhost
y
n
EOF

echo "📝 Test input prepared:"
echo "   Node IP: localhost"
echo "   Check coherence: y (yes)"
echo "   Check MySQL: n (no)"
echo ""

echo "🚀 Running test..."
echo ""

# Run the test
timeout 60s ./galerahealth -v < /tmp/test_input

echo ""
echo "✅ Test completed"

# Cleanup
rm -f /tmp/test_input
