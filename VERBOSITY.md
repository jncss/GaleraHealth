# GaleraHealth - Verbosity System

## Overview

GaleraHealth now includes a verbosity system with three different levels that allows users to control the amount of information displayed during execution.

## Verbosity Levels

### Level 0 - Minimal (Default)
**Usage:** `./galerahealth` (no flags)

**What it shows:**
- Only essential messages and final results
- Successful connections
- Main section titles
- Critical errors

**Example output:**
```
=== GaleraHealth - Galera Cluster Monitor ===
✓ Successfully connected to node 10.1.1.91
=== GALERA CLUSTER INFORMATION ===
🔍 Performing cluster coherence analysis...
```

### Level 1 - Normal (-v)
**Usage:** `./galerahealth -v`

**What it shows:**
- Everything from minimal level +
- Saved configuration loading
- Warnings and notices
- Operation confirmations

**Example output:**
```
=== GaleraHealth - Galera Cluster Monitor ===
📋 💾 Loaded saved configuration from /home/user/.galerahealth
📋 ⚠️  Connection with keys failed: permission denied
📋 🔐 Attempting connection with password...
✓ Successfully connected to node 10.1.1.91
```

### Level 2 - Verbose (-vv)
**Usage:** `./galerahealth -vv`

**What it shows:**
- Everything from previous levels +
- Internal operation details
- File search information
- Detailed SSH connection processes
- Step-by-step configuration analysis

**Example output:**
```
🔍 Verbosity level set to: 2
=== GaleraHealth - Galera Cluster Monitor ===
🔍 Application started with verbosity level 2
📋 💾 Loaded saved configuration from /home/user/.galerahealth
🔍    Last used: Node IP: 10.1.1.91, SSH User: root, MySQL User: root
🔍 Attempting SSH connection to root@10.1.1.91
🔍 Gathering cluster information from initial node
📋 🔍 Searching for cluster information...
🔍 📁 Searching for configuration files...
🔍 📁 Configuration files found: 3 files
```

### Level 3 - Debug (-vvv)
**Usage:** `./galerahealth -vvv`

**What it shows:**
- Everything from previous levels +
- Complete debugging information
- Password encryption/decryption details
- Complete lists of found files
- Internal configurations
- Raw data

**Example output:**
```
🐛 Verbosity level set to: 3
=== GaleraHealth - Galera Cluster Monitor ===
🐛 Application started with verbosity level 3
📋 💾 Loaded saved configuration from /home/user/.galerahealth
🔍    Last used: Node IP: 10.1.1.91, SSH User: root, MySQL User: root
🐛 Updated configuration: NodeIP=10.1.1.91, Username=root, CheckCoherence=true
🐛 CheckMySQL set to: true
🔍 Found saved encrypted password
🐛 Attempting to decrypt stored password
🐛 Password successfully decrypted
🐛   - /etc/mysql/conf.d/galera.cnf
🐛   - /etc/mysql/mysql.conf.d/mysqld.cnf
```

## Supported Flag Formats

The system supports different formats for specifying verbosity level:

```bash
./galerahealth -v      # Level 1 (normal)
./galerahealth -vv     # Level 2 (verbose)
./galerahealth -vvv    # Level 3 (debug)
```

## Icons Used by Level

- **📋** - Normal messages (-v and higher)
- **🔍** - Detailed information (-vv and higher)  
- **🐛** - Debug information (-vvv only)

## Compatibility with Other Flags

Verbosity flags can be combined with other options:

```bash
./galerahealth -vv --clear-config    # Clear with detailed verbosity
./galerahealth -v --help             # Help (verbosity doesn't affect help)
```

## When to Use Each Level

- **Minimal (default):** Normal daily use, only want to see results
- **Normal (-v):** When you want more information about what's happening
- **Verbose (-vv):** For troubleshooting or understanding internal processes
- **Debug (-vvv):** For development or deep problem diagnosis

## Technical Implementation

The system uses centralized logging functions:
- `logMinimal()` - Always shown
- `logNormal()` - Only with -v and higher
- `logVerbose()` - Only with -vv and higher
- `logDebug()` - Only with -vvv

This allows granular output control without affecting performance when detailed information is not needed.
