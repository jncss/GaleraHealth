#!/bin/bash

# Quick deployment script - minimal version
# Usage: ./deploy-quick.sh

set -e

REMOTE="root@10.1.1.91"
REMOTE_PATH="/usr/local/bin/galerahealth"

echo "🔨 Building..."
go build -o galerahealth .

echo "🚀 Deploying to $REMOTE..."
scp galerahealth "$REMOTE:$REMOTE_PATH"

echo "🔧 Setting permissions..."
ssh "$REMOTE" "chmod +x $REMOTE_PATH"

echo "✅ Deployed successfully!"
echo "Usage: ssh $REMOTE && galerahealth"
