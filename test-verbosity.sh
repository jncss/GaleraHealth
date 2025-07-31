#!/bin/bash

echo "=== Testing GaleraHealth Verbosity Levels ==="
echo

echo "1. Default (minimal) verbosity:"
echo "./galerahealth"
echo "   Shows only essential messages"
echo

echo "2. Normal verbosity (-v):"
echo "./galerahealth -v"
echo "   Shows normal operations + warnings"
echo

echo "3. Verbose output (-vv):"
echo "./galerahealth -vv"
echo "   Shows detailed operations + debug info"
echo

echo "4. Debug output (-vvv):"
echo "./galerahealth -vvv"
echo "   Shows full debug output + raw data"
echo

echo "Examples of what you'll see at each level:"
echo
echo "📋 Minimal (always shown):"
echo "   ✓ Successfully connected to node X.X.X.X"
echo "   🔍 Performing cluster coherence analysis..."
echo
echo "📋 Normal (-v and above):"
echo "   💾 Loaded saved configuration from ~/.galerahealth"
echo "   ⚠️  Connection with keys failed"
echo
echo "🔍 Verbose (-vv and above):"
echo "   📁 Searching for configuration files..."
echo "   🔑 Attempting SSH connection without password"
echo "   Gathering cluster information from initial node"
echo
echo "🐛 Debug (-vvv only):"
echo "   Application started with verbosity level 3"
echo "   - /etc/mysql/conf.d/galera.cnf"
echo "   Password successfully encrypted and marked for storage"
