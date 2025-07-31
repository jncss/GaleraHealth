#!/bin/bash

# Test script to check localhost MySQL functionality
cd /home/jncss/go/src/GaleraHealth

# Use expect-like functionality or direct test
echo "Testing localhost MySQL check..."

# Create a test input file
cat > test_input.txt << EOF
localhost
n
y
root

EOF

# Run the test
./galerahealth -v < test_input.txt

# Clean up
rm -f test_input.txt
