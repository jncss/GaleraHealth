#!/bin/bash

# Script to compile GaleraHealth
# 
# Usage: ./make.sh

echo "Compiling GaleraHealth..."

# Clean previous build if exists
if [ -f "galerahealth" ]; then
    echo "Removing previous build..."
    rm -f galerahealth
fi

# Compile the project
go build -o galerahealth

if [ $? -eq 0 ]; then
    echo "✅ Compilation successful!"
    echo "Binary created: galerahealth"
    echo "To run: ./galerahealth"
else
    echo "❌ Compilation failed!"
    exit 1
fi
