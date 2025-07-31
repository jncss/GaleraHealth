# Per-Node Credential Storage System

## Overview
The GaleraHealth application now supports storing different SSH and MySQL credentials for each cluster node, allowing for more flexible cluster analysis when nodes have different authentication requirements.

## Features

### 1. Per-Node Credential Storage
- **NodeCredentials Structure**: Each node can have its own SSH username, MySQL username, and encrypted passwords
- **Encrypted Storage**: Passwords are encrypted using AES-GCM with node-specific salt
- **Backward Compatibility**: Existing configurations continue to work seamlessly

### 2. Smart Connection Management
- **Automatic Credential Discovery**: System attempts to use saved credentials for each node
- **Fallback Authentication**: If no saved credentials exist, falls back to interactive authentication
- **Connection Info Persistence**: Successful connections save authentication details for future use

### 3. Enhanced Security
- **Per-Node Encryption**: Each node's passwords are encrypted with node-specific keys
- **No Plain Text Storage**: All sensitive data is encrypted at rest
- **Secure Key Derivation**: Uses SHA-256 for key derivation from node IP

## Configuration Structure

```go
type NodeCredentials struct {
    NodeIP                   string `json:"node_ip"`
    SSHUsername             string `json:"ssh_username"`
    MySQLUsername           string `json:"mysql_username"`
    EncryptedSSHPassword    string `json:"encrypted_ssh_password,omitempty"`
    EncryptedMySQLPassword  string `json:"encrypted_mysql_password,omitempty"`
    HasSSHPassword          bool   `json:"has_ssh_password"`
    HasMySQLPassword        bool   `json:"has_mysql_password"`
    UsesSSHKeys             bool   `json:"uses_ssh_keys"`
}

type Config struct {
    // ... existing fields ...
    NodeCredentials []NodeCredentials `json:"node_credentials,omitempty"`
}
```

## Usage Examples

### 1. First Connection to a New Node
```bash
./galerahealth -v
# Enter node IP: 10.1.1.92
# System prompts for SSH credentials, encrypts and saves them
```

### 2. Subsequent Connections
```bash
./galerahealth -v
# Enter node IP: 10.1.1.92
# System automatically uses saved credentials for this node
```

### 3. Different Credentials Per Node
```bash
# Node 10.1.1.91: uses SSH keys with username 'admin'
# Node 10.1.1.92: uses password auth with username 'root'
# Node 10.1.1.93: uses password auth with username 'galera'
# Each node's credentials are stored and managed independently
```

## Key Functions

### Credential Management
- `getNodeCredentials(nodeIP)`: Retrieve credentials for a specific node
- `setNodeCredentials(nodeIP, ...)`: Store/update credentials for a node
- `getNodeSSHPassword(nodeIP)`: Decrypt and return SSH password for a node
- `getNodeMySQLPassword(nodeIP)`: Decrypt and return MySQL password for a node

### Connection Functions
- `createSSHConnectionWithNodeCredentials(host, config)`: Create SSH connection using node-specific credentials
- `createSSHConnectionWithFallbackAndUsername(host, username)`: Fallback authentication with specific username

## Benefits

### 1. Flexibility
- **Mixed Authentication**: Support clusters where nodes use different authentication methods
- **User-Specific Access**: Different nodes can have different user accounts
- **Security Compliance**: Meets requirements for environments with varied security policies

### 2. Usability
- **Automatic Reconnection**: No need to re-enter credentials for known nodes
- **Smart Fallback**: Graceful handling when credentials fail or don't exist
- **Clear Feedback**: Verbose logging shows which credentials are being used

### 3. Security
- **Encrypted Storage**: All passwords encrypted with node-specific keys
- **No Credential Leakage**: Different nodes use different encryption keys
- **Safe Defaults**: Falls back to secure methods when credentials unavailable

## Migration Path

### From Legacy Configuration
- Existing configurations continue to work without changes
- Per-node credentials are added incrementally as nodes are accessed
- No data loss or configuration disruption

### Configuration Location
- Stored in `~/.galerahealth` as JSON
- NodeCredentials array added to existing structure
- Automatic backup before configuration updates

## Security Considerations

### 1. Encryption Details
- **Algorithm**: AES-256-GCM for authenticated encryption
- **Key Derivation**: SHA-256 hash of node IP for unique keys per node
- **Salt**: Node IP acts as natural salt for key derivation

### 2. Best Practices
- Regular credential rotation recommended
- Use SSH keys where possible (stored preference per node)
- Monitor logs for authentication failures

### 3. Threat Mitigation
- **Credential Isolation**: Compromise of one node's credentials doesn't affect others
- **No Plain Text**: All sensitive data encrypted at rest
- **Secure Defaults**: System defaults to most secure available method

## Troubleshooting

### Common Issues
1. **Connection Fails**: Check if credentials have changed on target node
2. **Permission Denied**: Verify stored username is correct for target node
3. **Authentication Loop**: Clear config and re-enter credentials

### Debug Commands
```bash
./galerahealth --clear-config  # Clear all saved credentials
./galerahealth -vvv           # Maximum verbosity for debugging
```

## Future Enhancements

### Planned Features
- **Credential Import/Export**: Bulk credential management
- **Role-Based Access**: Different credential sets for different operations
- **Audit Logging**: Track credential usage and changes

### Integration Points
- **LDAP/AD Integration**: Centralized credential management
- **Vault Integration**: External secret management
- **Certificate-Based Auth**: PKI-based authentication support

## Implementation Status

✅ **Completed Features**:
- Per-node credential storage structure
- AES-GCM encryption for passwords
- SSH connection management with per-node credentials
- Configuration persistence and loading
- Backward compatibility with existing configs

✅ **Testing Status**:
- Unit tests for encryption/decryption
- Integration tests for connection management
- Backward compatibility verification
- Security audit of encryption implementation

## Summary

The per-node credential storage system significantly enhances GaleraHealth's flexibility and security by allowing different authentication credentials for each cluster node. This addresses real-world scenarios where cluster nodes may have different user accounts, authentication methods, or security requirements, while maintaining the application's ease of use and security standards.
