# GaleraHea- 🔍 **Comprehensive Cluster Analysis**: Automatically discovers and analyzes all nodes in your Galera cluster
- 🚀 **Automated Mode**: Use `-y` flag to run with saved defaults without any prompts - perfect for monitoring scripts
- 📄 **Summary Mode**: Use `-y -s` for ultra-compact summary-only output ideal for monitoring dashboards
- 🔐 **Smart SSH Authentication**: Supports both SSH keys and password authentication with intelligent fallback
- 🏠 **Localhost Optimization**: Automatically detects and optimizes performance when running on cluster nodes
- 📊 **Configuration Coherence**: Validates that cluster configuration is consistent across all nodes
- 🔗 **MySQL Status Monitoring**: Checks MySQL/MariaDB service status and cluster connectivity
- 💾 **Persistent Configuration**: Saves connection settings with encrypted password storage
- 📈 **Multi-level Verbosity**: Four verbosity levels for different monitoring needs
- 🎯 **Health Summary**: Provides clear cluster health status and actionable recommendations
- 🌐 **Per-node Credentials**: Supports different SSH/MySQL credentials for each cluster node
- 🤖 **Smart Defaults**: Automatically detects multi-node clusters and adjusts behavior accordinglyth - Galera Cluster Monitor

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-1.23+-green)
![License](https://img.shields.io/badge/license-MIT-blue)

**GaleraHealth** is a comprehensive monitoring tool for MariaDB/MySQL Galera clusters that provides detailed cluster analysis, configuration coherence checking, and MySQL status monitoring across all cluster nodes.

## ✨ Features

- 🔍 **Comprehensive Cluster Analysis**: Automatically discovers and analyzes all nodes in your Galera cluster
- � **Automated Mode**: Use `-y` flag to run with saved defaults without any prompts - perfect for monitoring scripts
- �🔐 **Smart SSH Authentication**: Supports both SSH keys and password authentication with intelligent fallback
- 🏠 **Localhost Optimization**: Automatically detects and optimizes performance when running on cluster nodes
- 📊 **Configuration Coherence**: Validates that cluster configuration is consistent across all nodes
- 🔗 **MySQL Status Monitoring**: Checks MySQL/MariaDB service status and cluster connectivity
- 💾 **Persistent Configuration**: Saves connection settings with encrypted password storage
- 📈 **Multi-level Verbosity**: Four verbosity levels for different monitoring needs
- 🎯 **Health Summary**: Provides clear cluster health status and actionable recommendations
- 🌐 **Per-node Credentials**: Supports different SSH/MySQL credentials for each cluster node
- 🤖 **Smart Defaults**: Automatically detects multi-node clusters and adjusts behavior accordingly

## 🚀 Quick Start

### Requirements

- **Go 1.21+** for building from source
- **SSH access** to cluster nodes (key-based or password authentication)
- **MySQL/MariaDB credentials** for cluster nodes
- **Linux/Unix environment** (tested on Ubuntu, CentOS, Debian)

### Installation

1. **Build the binary**:
   ```bash
   go build -o galerahealth .
   ```

2. **Make it executable**:
   ```bash
   chmod +x galerahealth
   ```

3. **Install system-wide (optional)**:
   ```bash
   sudo cp galerahealth /usr/local/bin/
   ```

4. **Run the monitor**:
   ```bash
   # If installed system-wide
   galerahealth
   
   # Or run locally
   ./galerahealth
   ```

### Basic Usage

```bash
# Interactive mode (recommended for first use)
./galerahealth

# Automated mode using saved defaults (no prompts)
./galerahealth -y      # or --yes

# Summary mode - show only final summary (requires -y)
./galerahealth -y -s   # or --summary

# With verbosity
./galerahealth -v      # Normal verbosity
./galerahealth -vv     # Detailed verbosity
./galerahealth -vvv    # Debug verbosity

# Combine automated mode with verbosity
./galerahealth -y -v   # Automated with normal verbosity

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

### Example 3: Automated Mode with Saved Configuration
```bash
$ ./galerahealth -y
=== GaleraHealth - Galera Cluster Monitor ===
🚀 Running in automatic mode (-y) - using saved defaults
✓ Successfully connected to node 10.1.1.91

=== GALERA CLUSTER INFORMATION ===
🏠 Node IP: 10.1.1.91
🏷️  Cluster Name: my_galera_cluster
📍 Cluster Address: gcomm://10.1.1.91,10.1.1.92,10.1.1.93
🔖 Node Name: galera_node_1
🌐 Node Address: 10.1.1.91

🔍 Performing cluster coherence analysis...
📋 Found 3 nodes in cluster configuration
   1. 10.1.1.91 (initial node - already analyzed)
   2. 10.1.1.92 - connecting... ✓ Configuration retrieved
   3. 10.1.1.93 - connecting... ✓ Configuration retrieved

✅ CLUSTER CONFIGURATION IS COHERENT
   All nodes have consistent configuration

=== CLUSTER HEALTH SUMMARY ===
🎉 GALERA CLUSTER IN PERFECT HEALTH
   ✅ Configuration coherent across all nodes
📊 Total nodes: 3

# Perfect for monitoring scripts, CI/CD, or scheduled health checks
```

### Example 4: Summary Mode - Summary Only
```bash
$ ./galerahealth -y -s
=== CLUSTER HEALTH SUMMARY ===
🎉 GALERA CLUSTER IN PERFECT HEALTH
   ✅ Configuration coherent across all nodes
   ✅ All MySQL/MariaDB nodes responding correctly
   ✅ All nodes synchronized and ready
   ✅ Cluster in Primary state

📊 Total nodes: 3
🔗 Active nodes: 3/3

# Ultra-compact output perfect for monitoring dashboards, 
# log parsing, or quick status checks
```

## 🚀 Advanced Features

### Automated Mode (`-y`)
The `-y` flag enables automated execution using previously saved configuration values. This is perfect for monitoring scripts, cron jobs, or continuous integration pipelines:

- **No Interactive Prompts**: Uses saved IP addresses, usernames, and encrypted passwords
- **Smart Multi-node Detection**: Automatically checks all nodes if cluster configuration indicates multiple nodes
- **Monitoring Script Ready**: Designed for unattended execution in monitoring environments
- **Fallback Behavior**: Gracefully handles missing configuration by using sensible defaults

### Summary Mode (`-s`)
The `-s` flag must be combined with `-y` to provide ultra-compact output perfect for monitoring dashboards:

- **Summary Only**: Shows only the final cluster health summary
- **Minimal Output**: Suppresses all progress messages, diagnostics, and verbose information
- **Dashboard Friendly**: Ideal for parsing by monitoring systems or displaying in dashboards
- **Quick Status Check**: Perfect for automated health checks that need just the bottom line

**Example Summary Mode Output:**
```bash
$ ./galerahealth -y -s
=== CLUSTER HEALTH SUMMARY ===
✅ CLUSTER IS HEALTHY
📊 STATUS SUMMARY:
   🏠 Total nodes: 3
   ⚙️  Configuration coherent: ✅
   🔗 MySQL/MariaDB: ✅ All nodes responding
```

### Example 5: Troubleshooting with Verbosity
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

### Example 6: Per-node Credentials
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

### Automated Mode (`-y` flag)

The `-y` or `--yes` flag enables fully automated monitoring without any user prompts:

```bash
# Basic automated mode
./galerahealth -y

# Automated with verbosity for logging
./galerahealth -y -v

# Summary mode - ultra-compact summary only
./galerahealth -y -s
```

**Behavior with `-y` flag:**
- ✅ Uses saved configuration values automatically
- ✅ Intelligently detects multi-node clusters and enables coherence checking
- ✅ Skips password prompts if SSH keys fail (gracefully handles connection errors)
- ✅ Perfect for monitoring scripts, CI/CD pipelines, and scheduled health checks
- ⚠️ Requires existing configuration file (`~/.galerahealth`) from previous interactive run
- ⚠️ If no saved configuration exists, displays helpful error message

**Smart Multi-node Detection:**
When using `-y`, GaleraHealth automatically detects if you're monitoring a multi-node cluster by analyzing the `wsrep_cluster_address` and enables cluster coherence checking even if your saved preference was disabled.

### Summary Mode (`-s` flag)

The `-s` flag must be used in combination with `-y` to provide ultra-compact output:

```bash
# Summary mode - shows only the final summary
./galerahealth -y -s
```

**Summary Mode Features:**
- 📄 **Summary Only**: Displays only the final cluster health summary
- 🔇 **Minimal Output**: Suppresses all progress messages and diagnostic information
- 📊 **Dashboard Ready**: Perfect for monitoring systems and automated parsing
- ⚡ **Quick Status**: Ideal for rapid health checks and monitoring scripts
- ⚠️ **Requires `-y`**: Must be combined with automated mode

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

**Summary Mode Issues**
```bash
❌ Error: -s flag requires -y flag
```
*Solution*: Summary mode (`-s`) must be used together with automated mode (`-y`): `./galerahealth -y -s`

**No Output in Summary Mode**
```bash
$ ./galerahealth -y -s
# (no output)
```
*Solution*: This typically indicates a configuration error. Run without `-s` flag to see detailed error messages: `./galerahealth -y -v`

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

## 🤖 Automated Monitoring Use Cases

### CI/CD Integration
```bash
# In your deployment pipeline
./galerahealth -y || exit 1
echo "Galera cluster verified successfully"
```

### Scheduled Health Checks
```bash
# Add to crontab for regular monitoring
# Check cluster health every 15 minutes with full logging
*/15 * * * * /path/to/galerahealth -y -v >> /var/log/galera-health.log 2>&1

# For minimal logging - only record summaries every 5 minutes
*/5 * * * * /path/to/galerahealth -y -s >> /var/log/galera-health-summary.log 2>&1

# Combined approach - detailed logs hourly, summaries every 5 minutes
0 * * * * /path/to/galerahealth -y -v >> /var/log/galera-health-detailed.log 2>&1
*/5 * * * * /path/to/galerahealth -y -s >> /var/log/galera-health-summary.log 2>&1
```

### Monitoring Scripts
```bash
#!/bin/bash
# Basic monitoring script with automated mode

if ./galerahealth -y > /dev/null 2>&1; then
    echo "✅ Galera cluster is healthy"
    exit 0
else
    echo "❌ Galera cluster has issues - check logs"
    ./galerahealth -y -v  # Get detailed output for debugging
    exit 1
fi
```

```bash
#!/bin/bash
# Advanced monitoring script with summary mode for dashboard integration

# Get cluster status in summary mode (summary only)
STATUS_OUTPUT=$(./galerahealth -y -s 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    # Parse the summary for dashboard/monitoring system
    echo "GALERA_STATUS=HEALTHY"
    echo "$STATUS_OUTPUT" | grep "Total nodes:" | sed 's/📊 /GALERA_NODES=/'
    echo "GALERA_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
else
    echo "GALERA_STATUS=UNHEALTHY"
    echo "GALERA_NODES=UNKNOWN"
    echo "GALERA_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    # Get detailed logs for troubleshooting
    ./galerahealth -y -v
fi
```

### Docker/Kubernetes Health Checks
```dockerfile
# In your Dockerfile
HEALTHCHECK --interval=5m --timeout=30s --retries=3 \
  CMD ./galerahealth -y || exit 1
```

## 🔒 Security Considerations

- **Password Encryption**: All stored passwords use AES-GCM encryption
- **SSH Keys**: Preferred authentication method for security
- **Local Access**: Localhost operations use direct file access (no SSH)
- **Configuration Protection**: Config file permissions restricted to owner

##  License

MIT License - see LICENSE file for details.

## 🤝 Support

For issues, questions, or contributions:

1. **Check troubleshooting section** in this README
2. **Use debug mode** (`-vvv`) to gather detailed logs
3. **For automated mode issues**, first run interactively: `./galerahealth -vv`
4. **Review configuration files** for consistency
5. **Verify network connectivity** between cluster nodes

**Common Commands for Troubleshooting:**
```bash
# Interactive mode with verbose output
./galerahealth -vv

# Check if saved configuration exists
ls -la ~/.galerahealth

# Clear configuration and reconfigure
./galerahealth --clear-config
./galerahealth  # Interactive setup

# Test automated mode with debug output
./galerahealth -y -vvv
```

---

**GaleraHealth** - Keep your Galera cluster healthy! 🚀

