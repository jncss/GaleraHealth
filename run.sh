#!/bin/bash

# Script to run GaleraHealth
# 
# Usage: ./run.sh

echo "Compiling GaleraHealth..."
go build -o galerahealth

if [ $? -eq 0 ]; then
    echo "Compilation successful! Running the application..."
    echo ""
    ./galerahealth
else
    echo "Compilation error!"
    exit 1
fi
