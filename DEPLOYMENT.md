# GaleraHealth - Deployment Scripts

## 📦 Deployment Options

Two deployment scripts are provided for copying the GaleraHealth executable to remote servers:

### 1. 🚀 Full Deployment Script (`deploy.sh`)
**Comprehensive deployment with full validation and reporting**

#### Features:
- ✅ **Complete build process** with cleanup
- ✅ **Connectivity testing** before deployment
- ✅ **Binary verification** and size reporting
- ✅ **Remote permissions setup** 
- ✅ **Deployment verification** with remote testing
- ✅ **Colored output** for better readability
- ✅ **Detailed summary** with usage instructions

#### Usage:
```bash
./deploy.sh
```

#### Example Output:
```bash
[INFO] Starting GaleraHealth deployment process...
[INFO] Cleaning previous build...
[INFO] Building GaleraHealth application...
[SUCCESS] Build completed successfully
[INFO] Binary created: galerahealth (7.1M)
[INFO] Testing connectivity to 10.1.1.91...
[INFO] Deploying to root@10.1.1.91:/usr/local/bin/galerahealth...
[SUCCESS] Binary deployed successfully to 10.1.1.91
[INFO] Setting executable permissions on remote server...
[SUCCESS] Executable permissions set
[INFO] Verifying deployment...
[SUCCESS] Deployment verification completed
[INFO] Testing remote binary...
[SUCCESS] Remote binary is working correctly

🎉 DEPLOYMENT COMPLETED SUCCESSFULLY!

Deployment Summary:
  • Local binary:  /home/user/GaleraHealth/galerahealth
  • Remote server: root@10.1.1.91
  • Remote path:   /usr/local/bin/galerahealth
  • Binary size:   7.1M

Usage on remote server:
  ssh root@10.1.1.91
  galerahealth --help
  galerahealth
```

### 2. ⚡ Quick Deployment Script (`deploy-quick.sh`)
**Minimal deployment for rapid iteration**

#### Features:
- ✅ **Fast execution** - minimal output
- ✅ **Essential steps only** - build, copy, permissions
- ✅ **Compact script** - easy to modify
- ✅ **Quick feedback** - immediate success/failure

#### Usage:
```bash
./deploy-quick.sh
```

#### Example Output:
```bash
🔨 Building...
🚀 Deploying to root@10.1.1.91...
galerahealth     100% 7245KB  21.6MB/s   00:00    
🔧 Setting permissions...
✅ Deployed successfully!
Usage: ssh root@10.1.1.91 && galerahealth
```

## ⚙️ Configuration

Both scripts are configured for:
- **Target Server**: `10.1.1.91`
- **Remote User**: `root`
- **Remote Path**: `/usr/local/bin/galerahealth`
- **Local Binary**: `galerahealth`

### Customizing Target Server
To deploy to a different server, edit the scripts:

**deploy.sh:**
```bash
REMOTE_HOST="your.server.ip"
REMOTE_USER="your_username"
REMOTE_PATH="/path/to/install"
```

**deploy-quick.sh:**
```bash
REMOTE="user@your.server.ip"
REMOTE_PATH="/path/to/install/galerahealth"
```

## 🔧 Prerequisites

### Local Machine:
- Go 1.23+ installed
- GaleraHealth source code
- SSH client configured

### Remote Server:
- SSH access configured
- Write permissions to `/usr/local/bin` (or target directory)
- Linux compatible architecture

### SSH Configuration:
Ensure SSH key-based authentication is set up:
```bash
# Test SSH access
ssh root@10.1.1.91 "echo 'SSH access works'"

# If needed, copy SSH keys
ssh-copy-id root@10.1.1.91
```

## 🎯 Usage Scenarios

### Development Workflow:
```bash
# Make code changes
vim main.go

# Quick deployment for testing
./deploy-quick.sh

# Test on remote server
ssh root@10.1.1.91 "galerahealth -v"
```

### Production Deployment:
```bash
# Full deployment with validation
./deploy.sh

# Verify everything is working
ssh root@10.1.1.91 "galerahealth --help"
```

### Batch Deployment:
```bash
# Deploy to multiple servers
for server in 10.1.1.91 10.1.1.92 10.1.1.93; do
    sed "s/10.1.1.91/$server/g" deploy-quick.sh > deploy-$server.sh
    chmod +x deploy-$server.sh
    ./deploy-$server.sh
done
```

## 🚨 Troubleshooting

### SSH Connection Issues:
```bash
# Test connectivity
ping 10.1.1.91

# Test SSH access
ssh root@10.1.1.91 "whoami"

# Check SSH configuration
ssh -v root@10.1.1.91
```

### Permission Issues:
```bash
# Check remote directory permissions
ssh root@10.1.1.91 "ls -la /usr/local/bin"

# Create directory if needed
ssh root@10.1.1.91 "mkdir -p /usr/local/bin"
```

### Build Issues:
```bash
# Clean and rebuild locally
go clean
go build -o galerahealth .

# Test local binary
./galerahealth --help
```

## 📊 Script Comparison

| Feature | deploy.sh | deploy-quick.sh |
|---------|-----------|-----------------|
| Build verification | ✅ | ❌ |
| Connectivity test | ✅ | ❌ |
| Colored output | ✅ | ❌ |
| Deployment verification | ✅ | ❌ |
| Remote binary test | ✅ | ❌ |
| Detailed summary | ✅ | ❌ |
| Speed | Slower | ⚡ Faster |
| Output verbosity | High | Minimal |
| Best for | Production | Development |

## 🎉 Success

After successful deployment, GaleraHealth will be available system-wide on the remote server:

```bash
# Connect to remote server
ssh root@10.1.1.91

# Use GaleraHealth from anywhere
galerahealth
galerahealth -v
galerahealth --help

# Works with localhost optimization
galerahealth
# Enter IP: localhost (will use local file access)
```
