#!/bin/bash

# GaleraHealth Deployment Script
# Builds and deploys the GaleraHealth executable to remote server

set -e  # Exit on any error

# Configuration
REMOTE_HOST="10.1.1.91"
REMOTE_USER="root"
REMOTE_PATH="/usr/local/bin"
BINARY_NAME="galerahealth"
BUILD_TARGET="galerahealth"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    print_error "main.go not found. Please run this script from the GaleraHealth project directory."
    exit 1
fi

print_status "Starting GaleraHealth deployment process..."

# Step 1: Clean previous build
print_status "Cleaning previous build..."
if [ -f "$BUILD_TARGET" ]; then
    rm "$BUILD_TARGET"
    print_status "Removed previous binary: $BUILD_TARGET"
fi

# Step 2: Build the application
print_status "Building GaleraHealth application..."
if go build -o "$BUILD_TARGET" .; then
    print_success "Build completed successfully"
else
    print_error "Build failed"
    exit 1
fi

# Step 3: Verify binary was created
if [ ! -f "$BUILD_TARGET" ]; then
    print_error "Binary $BUILD_TARGET was not created"
    exit 1
fi

# Step 4: Show binary info
BINARY_SIZE=$(ls -lh "$BUILD_TARGET" | awk '{print $5}')
print_status "Binary created: $BUILD_TARGET ($BINARY_SIZE)"

# Step 5: Test connectivity to remote host
print_status "Testing connectivity to $REMOTE_HOST..."
if ! ping -c 1 -W 3 "$REMOTE_HOST" >/dev/null 2>&1; then
    print_warning "Cannot ping $REMOTE_HOST, but proceeding with deployment attempt..."
fi

# Step 6: Copy binary to remote server
print_status "Deploying to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/$BINARY_NAME..."

# Try to remove existing file first (in case it's locked)
ssh "$REMOTE_USER@$REMOTE_HOST" "rm -f $REMOTE_PATH/$BINARY_NAME" 2>/dev/null || true

if scp "$BUILD_TARGET" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/$BINARY_NAME"; then
    print_success "Binary deployed successfully to $REMOTE_HOST"
else
    print_error "Failed to copy binary to remote server"
    print_status "Attempting alternative deployment method..."
    
    # Alternative: copy to temp location first, then move
    TEMP_PATH="/tmp/$BINARY_NAME.new"
    if scp "$BUILD_TARGET" "$REMOTE_USER@$REMOTE_HOST:$TEMP_PATH" && \
       ssh "$REMOTE_USER@$REMOTE_HOST" "mv $TEMP_PATH $REMOTE_PATH/$BINARY_NAME"; then
        print_success "Binary deployed successfully using alternative method"
    else
        print_error "All deployment methods failed"
        exit 1
    fi
fi

# Step 7: Set executable permissions on remote server
print_status "Setting executable permissions on remote server..."
if ssh "$REMOTE_USER@$REMOTE_HOST" "chmod +x $REMOTE_PATH/$BINARY_NAME"; then
    print_success "Executable permissions set"
else
    print_error "Failed to set executable permissions"
    exit 1
fi

# Step 8: Verify deployment
print_status "Verifying deployment..."
if ssh "$REMOTE_USER@$REMOTE_HOST" "ls -la $REMOTE_PATH/$BINARY_NAME"; then
    print_success "Deployment verification completed"
else
    print_error "Failed to verify deployment"
    exit 1
fi

# Step 9: Test remote binary
print_status "Testing remote binary..."
if ssh "$REMOTE_USER@$REMOTE_HOST" "$REMOTE_PATH/$BINARY_NAME --help" >/dev/null 2>&1; then
    print_success "Remote binary is working correctly"
else
    print_warning "Remote binary test failed, but deployment completed"
fi

# Step 10: Show deployment summary
echo ""
print_success "🎉 DEPLOYMENT COMPLETED SUCCESSFULLY!"
echo ""
echo -e "${BLUE}Deployment Summary:${NC}"
echo "  • Local binary:  $(pwd)/$BUILD_TARGET"
echo "  • Remote server: $REMOTE_USER@$REMOTE_HOST"
echo "  • Remote path:   $REMOTE_PATH/$BINARY_NAME"
echo "  • Binary size:   $BINARY_SIZE"
echo ""
echo -e "${BLUE}Usage on remote server:${NC}"
echo "  ssh $REMOTE_USER@$REMOTE_HOST"
echo "  $BINARY_NAME --help"
echo "  $BINARY_NAME"
echo ""
print_status "You can now use GaleraHealth on the remote server!"
