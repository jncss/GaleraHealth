# GaleraHealth - Galera Cluster Monitor

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-1.23+-green)
![License](https://img.shields.io/badge/license-MIT-blue)

**GaleraHealth** is a comprehensive monitoring tool for MariaDB/MySQL Galera clusters that provides detailed cluster analysis, configuration coherence checking, and MySQL status monitoring across all cluster nodes.

## ✨ Features

- 🔍 **Comprehensive Cluster Analysis**: Automatically discovers and analyzes all nodes in your Galera cluster
- 🔐 **Smart SSH Authentication**: Supports both SSH keys and password authentication with intelligent fallback
- 🏠 **Localhost Optimization**: Automatically detects and optimizes performance when running on cluster nodes
- 📊 **Configuration Coherence**: Validates that cluster configuration is consistent across all nodes
- 🔗 **MySQL Status Monitoring**: Checks MySQL/MariaDB service status and cluster connectivity
- 💾 **Persistent Configuration**: Saves connection settings with encrypted password storage
- 📈 **Multi-level Verbosity**: Four verbosity levels for different monitoring needs
- 🎯 **Health Summary**: Provides clear cluster health status and actionable recommendations
- 🌐 **Per-node Credentials**: Supports different SSH/MySQL credentials for each cluster node

## 🚀 Quick Start

### Installation

1. **Download or build the binary**:
   ```bash
   # Option 1: Build from source
   go build -o galerahealth .
   
   # Option 2: Use deployment script
   ./deploy.sh
   ```

2. **Make it executable**:
   ```bash
   chmod +x galerahealth
   ```

3. **Run the monitor**:
   ```bash
   ./galerahealth
   ```

### Basic Usage

```bash
# Interactive mode (recommended for first use)
./galerahealth

# With verbosity
./galerahealth -v      # Normal verbosity
./galerahealth -vv     # Detailed verbosity
./galerahealth -vvv    # Debug verbosity

# Configuration management
./galerahealth --clear-config    # Clear saved settings
./galerahealth --help           # Show help
```

## 📋 Usage Examples

### Example 1: First Time Setup
```bash
$ ./galerahealth
=== GaleraHealth - Galera Cluster Monitor ===

Enter the Galera cluster node IP (default: localhost): 10.1.1.91
Enter SSH username (default: root): 
🔐 SSH Key Authentication: Attempting connection...
✅ Connected successfully using SSH keys

🔍 Analyzing Galera configuration...
=== GALERA CLUSTER INFORMATION ===
🏷️  Cluster Name: production_cluster
📍 Cluster Address: gcomm://10.1.1.91,10.1.1.92,10.1.1.93
🔖 Node Name: node1
🌐 Node Address: 10.1.1.91

Do you want to check cluster configuration coherence across all nodes? (Y/n): y
✅ All nodes have coherent configuration

Do you want to check MySQL/MariaDB cluster status on all nodes? (y/N): y
✅ All MySQL services are healthy

=== CLUSTER HEALTH SUMMARY ===
✅ CLUSTER IS HEALTHY
📊 STATUS SUMMARY:
   🏠 Total nodes: 3
   ⚙️  Configuration coherent: ✅
   🔗 MySQL/MariaDB: ✅ All nodes responding
```

### Example 2: Localhost Monitoring
```bash
$ ./galerahealth
Enter the Galera cluster node IP (default: localhost): localhost
🏠 Local connection detected - skipping SSH authentication
🔍 Analyzing local Galera configuration...

# Automatically uses local file access and command execution
# No SSH overhead for optimal performance
```

### Example 3: Troubleshooting with Verbosity
```bash
$ ./galerahealth -vv
=== GaleraHealth - Galera Cluster Monitor ===
📋 💾 Loaded saved configuration from ~/.galerahealth
Enter the Galera cluster node IP (default: 10.1.1.91): 

📋 🔐 SSH Key Authentication: Attempting connection...
📋 📁 Searching for configuration files...
📋    - /etc/mysql/mariadb.conf.d/60-galera.cnf
📋    - /etc/mysql/my.cnf
📋 🔍 Analyzing /etc/mysql/mariadb.conf.d/60-galera.cnf...
📋 ✅ Found wsrep_cluster_name: production_cluster
📋 ✅ Found wsrep_cluster_address: gcomm://10.1.1.91,10.1.1.92,10.1.1.93
```

### Example 4: Per-node Credentials
```bash
$ ./galerahealth
# First node uses default credentials
Enter the Galera cluster node IP: 10.1.1.91
Enter SSH username: root

# System discovers additional nodes and prompts for their credentials
🔍 Found additional cluster nodes: 10.1.1.92, 10.1.1.93

Configure credentials for node 10.1.1.92:
  SSH Username (default: root): dbadmin
  🔐 Enter SSH password for dbadmin@10.1.1.92: [encrypted storage]
  
Configure credentials for node 10.1.1.93:
  SSH Username (default: root): ubuntu
  🔐 Enter SSH password for ubuntu@10.1.1.93: [encrypted storage]
```

## ⚙️ Configuration

### Configuration File
GaleraHealth stores configuration in `~/.galerahealth` (JSON format with encrypted passwords):

```json
{
  "node_credentials": [
    {
      "node_ip": "10.1.1.91",
      "ssh_username": "root",
      "encrypted_ssh_password": "...",
      "mysql_username": "root",
      "encrypted_mysql_password": "..."
    }
  ],
  "check_mysql_status": true,
  "check_cluster_coherence": true
}
```

### Environment Variables
- `GALERAHEALTH_CONFIG`: Custom configuration file path
- `GALERAHEALTH_LOG_LEVEL`: Default verbosity level (0-3)

## 🔧 Advanced Features

### Verbosity Levels

| Level | Flag | Description | Use Case |
|-------|------|-------------|----------|
| **Silent** | (none) | Minimal output | Production monitoring |
| **Normal** | `-v` | Standard operations | Daily monitoring |
| **Verbose** | `-vv` | Detailed operations | Troubleshooting |
| **Debug** | `-vvv` | Full debug output | Development/Support |

### Cluster Health Status

GaleraHealth provides comprehensive health assessment:

- ✅ **HEALTHY**: All nodes responsive, configuration coherent
- ⚠️ **WARNING**: Minor issues detected, cluster functional
- ❌ **CRITICAL**: Major issues requiring immediate attention

### SSH Authentication Methods

1. **SSH Keys** (preferred): Automatic key-based authentication
2. **Password Authentication**: Fallback with encrypted storage
3. **Mixed Credentials**: Different authentication per node

## 🚨 Troubleshooting

### Common Issues

**SSH Connection Failures**
```bash
❌ SSH connection failed: authentication failed
```
*Solution*: Verify SSH credentials, check SSH service status, ensure network connectivity

**MySQL Connection Issues**
```bash
❌ MySQL connection failed: access denied
```  
*Solution*: Verify MySQL credentials, check MySQL service status, validate user permissions

**Configuration Incoherence**
```bash
❌ CLUSTER CONFIGURATION ISSUES DETECTED
   Found 2 configuration errors:
   1. Node 10.1.1.92: Different cluster name 'old_cluster'
   2. Node 10.1.1.93: Missing wsrep_node_address
```
*Solution*: Review and synchronize configuration files across all nodes

### Debug Mode
Use maximum verbosity for detailed troubleshooting:
```bash
./galerahealth -vvv 2>&1 | tee galerahealth-debug.log
```

## 📊 Output Formats

### Summary Format
```
=== CLUSTER HEALTH SUMMARY ===
✅ CLUSTER IS HEALTHY

📊 STATUS SUMMARY:
   🏠 Total nodes: 3
   ⚙️  Configuration coherent: ✅
   🔗 MySQL/MariaDB: ✅ All nodes responding

🎯 RECOMMENDATIONS:
   • All systems operating normally
   • Regular monitoring recommended
```

### Detailed Analysis
```
=== CLUSTER ANALYSIS RESULTS ===
📊 Nodes analyzed: 3/3
🎯 Cluster name: production_cluster

📋 All nodes in cluster:
   1. 10.1.1.91 (localhost)
      ✅ Cluster Name: production_cluster
      ✅ Node Name: node1
      ✅ MySQL Status: Active, Cluster Size: 3
      
   2. 10.1.1.92
      ✅ Cluster Name: production_cluster  
      ✅ Node Name: node2
      ✅ MySQL Status: Active, Cluster Size: 3
```

## 🔒 Security Considerations

- **Password Encryption**: All stored passwords use AES-GCM encryption
- **SSH Keys**: Preferred authentication method for security
- **Local Access**: Localhost operations use direct file access (no SSH)
- **Configuration Protection**: Config file permissions restricted to owner

## 🚀 Deployment

### Remote Deployment
Use the included deployment script:
```bash
# Edit deploy.sh with your target server details
vim deploy.sh

# Deploy to remote server
./deploy.sh
```

### System Service Integration
```bash
# Create systemd service for regular monitoring
sudo tee /etc/systemd/system/galerahealth.service << EOF
[Unit]
Description=GaleraHealth Cluster Monitor
After=network.target

[Service]
Type=oneshot
User=galera
ExecStart=/usr/local/bin/galerahealth -v
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl enable galerahealth.service
sudo systemctl start galerahealth.service
```

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Support

For issues, questions, or contributions:

1. **Check troubleshooting section** in this README
2. **Use debug mode** (`-vvv`) to gather detailed logs
3. **Review configuration files** for consistency
4. **Verify network connectivity** between cluster nodes

---

**GaleraHealth** - Keep your Galera cluster healthy! 🚀

