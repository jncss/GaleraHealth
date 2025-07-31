# GaleraHealth

A Go application for monitoring the status of a Galera MySQL/MariaDB cluster.

## Description

GaleraHealth is a tool that allows you to connect to a node in a Galera cluster via SSH to obtain information about the cluster configuration, specifically:

- `wsrep_cluster_name`: The cluster name
- `wsrep_cluster_address`: The cluster nodes address
- `wsrep_node_name`: The individual node name
- `wsrep_node_address`: The individual node address

Additionally, it can perform a **cluster coherence analysis** by connecting to all nodes in the cluster and verifying that their configurations are consistent.

## Features

- ✅ Smart SSH authentication:
  - First attempt with SSH keys (id_rsa, id_ecdsa, id_ed25519)
  - Second attempt with password if keys fail
  - **Password reuse**: Automatically reuses entered password for subsequent nodes
  - **Per-node credentials**: Supports different SSH/MySQL credentials for each cluster node
- ✅ Automatic search in multiple configuration file locations
- ✅ Runtime MySQL variables verification
- ✅ Cluster nodes parsing
- ✅ **Cluster coherence analysis**:
  - Connects to all nodes in the cluster
  - Verifies configuration consistency across nodes
  - Checks `wsrep_cluster_name`, `wsrep_cluster_address`, `wsrep_node_name`, `wsrep_node_address`
  - Provides detailed analysis and recommendations
  - **Mixed authentication support**: Different nodes can use different authentication methods
- ✅ **Persistent configuration with encryption**:
  - Saves connection preferences and credentials securely
  - AES-GCM encryption for sensitive data
  - Per-node credential storage for mixed environments
- ✅ **Localhost optimization**: Direct file access when running on cluster nodes
- ✅ **Verbosity control**: Multiple output levels (-v, -vv, -vvv) for different use cases
- ✅ **Comprehensive health summary**: Clear overview of cluster status and issues
- ✅ User-friendly interface

## Usage

1. Compile the application:
```bash
go build -o galerahealth
```
   Or use the build script:
```bash
./make.sh
```

2. Run the application:
```bash
./galerahealth
```
   Or compile and run in one step:
```bash
./run.sh
```

3. Follow the instructions:
   - Enter the cluster node IP
   - Enter SSH username (default: root - just press Enter to use it)
   - The application will try to connect with SSH keys first
   - If keys fail, it will ask for the SSH password
   - Choose whether to perform cluster coherence analysis (connects to all nodes)

## Supported configuration files

The application searches for the `wsrep_cluster_name`, `wsrep_cluster_address`, `wsrep_node_name`, and `wsrep_node_address` variables by:

- **Recursively scanning** all `*.cnf` files in `/etc/mysql/`
- **Checking** `/etc/my.cnf` if it exists

This approach ensures that all MySQL/MariaDB configuration files are analyzed, regardless of their specific location within the `/etc/mysql` directory structure.

It also attempts to get the information from MySQL runtime variables if accessible.

## Example output

```
=== GaleraHealth - Galera Cluster Monitor ===

Enter the Galera cluster node IP: 192.168.1.100
Enter SSH username (default: root): 
🔑 Attempting SSH connection without password to node 192.168.1.100...
✓ SSH connection successful using keys!

🔍 Searching for cluster information...
📁 Searching for configuration files...
📁 Configuration files found: 4 files
   - /etc/mysql/my.cnf
   - /etc/mysql/conf.d/galera.cnf
   - /etc/mysql/mariadb.conf.d/50-server.cnf
   - /etc/mysql/mysql.conf.d/mysqld.cnf
   Analyzing /etc/mysql/my.cnf...
   ✓ wsrep_cluster_name found in /etc/mysql/my.cnf
   ✓ wsrep_cluster_address found in /etc/mysql/my.cnf

=== GALERA CLUSTER INFORMATION ===

🖥️  Node analyzed: 192.168.1.100

📛 Cluster name (wsrep_cluster_name): my_galera_cluster
🌐 Cluster address (wsrep_cluster_address): gcomm://192.168.1.100,192.168.1.101,192.168.1.102
📍 Cluster nodes detected: 3
   1. 192.168.1.100
   2. 192.168.1.101
   3. 192.168.1.102
🏷️  Node name (wsrep_node_name): node1
🌍 Node address (wsrep_node_address): 192.168.1.100

✅ Cluster information obtained successfully!

Do you want to check cluster configuration coherence across all nodes? (y/N): y

🔍 Performing cluster coherence analysis...
📋 Found 3 nodes in cluster configuration
   1. 192.168.1.100 (initial node - already analyzed)
   2. 192.168.1.101 - connecting...
      🔐 Using saved password...
      ✓ Connected using saved password
      ✓ Configuration retrieved
   3. 192.168.1.102 - connecting...
      🔐 Using saved password...
      ✓ Connected using saved password
      ✓ Configuration retrieved

=== CLUSTER COHERENCE ANALYSIS ===

📊 Nodes analyzed: 3/3
🎯 Cluster name: my_galera_cluster

📋 Node Details:
   1. 192.168.1.100
      Cluster Name: my_galera_cluster
      Cluster Address: gcomm://192.168.1.100,192.168.1.101,192.168.1.102
      Node Name: node1
      Node Address: 192.168.1.100

   2. 192.168.1.101
      Cluster Name: my_galera_cluster
      Cluster Address: gcomm://192.168.1.100,192.168.1.101,192.168.1.102
      Node Name: node2
      Node Address: 192.168.1.101

   3. 192.168.1.102
      Cluster Name: my_galera_cluster
      Cluster Address: gcomm://192.168.1.100,192.168.1.101,192.168.1.102
      Node Name: node3
      Node Address: 192.168.1.102

✅ CLUSTER CONFIGURATION IS COHERENT
   All nodes have consistent configuration

💡 Recommendations:
   - Configuration looks good!
   - Monitor cluster status regularly
```

## Requirements

- Go 1.21 or higher
- SSH access to the Galera cluster node
- The node must have MySQL/MariaDB with Galera configured

## Dependencies

- `golang.org/x/crypto/ssh`: For SSH connection
- `golang.org/x/term`: For secure password reading

## Build Scripts

The project includes two convenient scripts:

- **`make.sh`**: Only compiles the project
  ```bash
  ./make.sh
  ```

- **`run.sh`**: Compiles and runs the application
  ```bash
  ./run.sh
  ```

## SSH Authentication

The application attempts to connect in this order:

1. **SSH Keys**: Automatically searches for keys in:
   - `~/.ssh/id_rsa`
   - `~/.ssh/id_ecdsa`
   - `~/.ssh/id_ed25519`

2. **Password**: If keys fail, asks for password

## Security Notes

⚠️ **IMPORTANT**: This application uses `ssh.InsecureIgnoreHostKey()` for simplicity. In a production environment, proper SSH host key verification should be implemented.

## Future Development

Possible improvements:

- [x] Support for SSH key authentication
- [x] Cluster coherence analysis across all nodes
- [ ] Secure host key verification
- [ ] Real-time cluster status monitoring (node states, sync status)
- [ ] Performance metrics collection
- [ ] Mode batch for multiple clusters
- [ ] Export results in JSON/XML format
- [ ] Configuration file support
- [ ] Web dashboard interface
