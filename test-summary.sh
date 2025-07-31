#!/bin/bash

# Test different scenarios
echo "=== Testing Single Node (No Coherence Check) ==="
echo -e "10.1.1.91\nroot\nn\n" | timeout 10s ./galerahealth 2>/dev/null | grep -A 20 "RESUM FINAL"

echo ""
echo "=== Testing Coherence Check (No MySQL) ==="
echo -e "10.1.1.91\nroot\ny\nn\n" | timeout 15s ./galerahealth 2>/dev/null | grep -A 20 "RESUM FINAL"

echo ""
echo "=== Testing Full Analysis (Coherence + MySQL) ==="
echo -e "10.1.1.91\nroot\ny\ny\nroot\n\n" | timeout 20s ./galerahealth 2>/dev/null | grep -A 25 "RESUM FINAL"
