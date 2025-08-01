package main

import (
	"fmt"
	"strings"
)

// progressPrint prints progress messages, suppressed in report mode
func progressPrint(format string, args ...interface{}) {
	if !reportMode {
		fmt.Printf(format, args...)
	}
}

// performClusterAnalysis analyzes cluster coherence across all nodes
func performClusterAnalysis(initialNode *GaleraClusterInfo, connInfo *SSHConnectionInfo, config *Config) (*ClusterAnalysis, string, error) {
	analysis := &ClusterAnalysis{
		InitialNode:  initialNode,
		AllNodes:     []*GaleraClusterInfo{initialNode},
		ConfigErrors: []string{},
		IsCoherent:   true,
	}

	// Extract cluster nodes from wsrep_cluster_address
	if initialNode.ClusterAddress != "" && strings.Contains(initialNode.ClusterAddress, "gcomm://") {
		addresses := strings.TrimPrefix(initialNode.ClusterAddress, "gcomm://")
		if addresses != "" {
			nodes := strings.Split(addresses, ",")
			for _, node := range nodes {
				node = strings.TrimSpace(node)
				if node != "" {
					// Remove port if present (e.g., 192.168.1.100:4567 -> 192.168.1.100)
					if colonIndex := strings.Index(node, ":"); colonIndex != -1 {
						node = node[:colonIndex]
					}
					analysis.ClusterNodes = append(analysis.ClusterNodes, node)
				}
			}
		}
	}

	if len(analysis.ClusterNodes) == 0 {
		return nil, "", fmt.Errorf("no cluster nodes found in wsrep_cluster_address")
	}

	progressPrint("📋 Found %d nodes in cluster configuration\n", len(analysis.ClusterNodes))

	// If initial connection was localhost, try to identify which cluster node represents localhost
	var localhostNodeIP string
	if isLocalhost(initialNode.NodeIP) && initialNode.NodeAddress != "" && !isLocalhost(initialNode.NodeAddress) {
		// The node address from config shows the real IP of this localhost
		localhostNodeIP = initialNode.NodeAddress
		logVerbose("🏠 Identified localhost as %s (from wsrep_node_address)", localhostNodeIP)
	}

	// Check each node in the cluster
	for i, nodeIP := range analysis.ClusterNodes {
		if nodeIP == initialNode.NodeIP || isLocalhost(nodeIP) || nodeIP == localhostNodeIP {
			// Skip initial node (already analyzed), localhost references, or identified localhost IP
			if nodeIP == initialNode.NodeIP {
				progressPrint("   %d. %s (initial node - already analyzed)\n", i+1, nodeIP)
			} else if isLocalhost(nodeIP) {
				progressPrint("   %d. %s (localhost - skipping SSH)\n", i+1, nodeIP)
			} else if nodeIP == localhostNodeIP {
				progressPrint("   %d. %s (this is localhost %s - already analyzed)\n", i+1, nodeIP, initialNode.NodeIP)
			}
			continue
		}

		progressPrint("   %d. %s - connecting...\n", i+1, nodeIP)

		// Check if we have valid SSH connection info
		var sshClient *SSHClient
		var err error

		if connInfo.Username == "local" {
			// Initial connection was localhost, but we can still connect to remote nodes via SSH
			logVerbose("      🌐 Initial node is localhost, attempting SSH connection to remote node %s", nodeIP)
			var newConnInfo *SSHConnectionInfo
			sshClient, newConnInfo, err = createSSHConnectionWithNodeCredentials(nodeIP, config)
			if err != nil {
				// Create a node info with error to include in analysis
				nodeInfo := &GaleraClusterInfo{
					NodeIP:      nodeIP,
					StatusError: fmt.Sprintf("SSH connection failed: %v", err),
				}
				analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
				analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to connect to remote node %s: %v", nodeIP, err))
				analysis.IsCoherent = false
				progressPrint("      ❌ SSH connection failed: %v\n", err)
				continue
			}

			// Save the new connection info for this remote node
			if newConnInfo != nil {
				sshPassword := ""
				if newConnInfo.HasPassword {
					sshPassword = newConnInfo.Password
				}
				err = config.setNodeCredentials(nodeIP, newConnInfo.Username, "", sshPassword, "", newConnInfo.UsedKeys)
				if err != nil {
					progressPrint("      ⚠️  Warning: Could not save credentials for node %s: %v\n", nodeIP, err)
				} else {
					if newConnInfo.HasPassword {
						progressPrint("      ✓ SSH password saved for node %s\n", nodeIP)
					}
					progressPrint("      ✓ SSH credentials saved for node %s\n", nodeIP)
				}
			}
		} else {
			// Use per-node credentials for connection
			var newConnInfo *SSHConnectionInfo
			sshClient, newConnInfo, err = createSSHConnectionWithNodeCredentials(nodeIP, config)
			if err != nil {
				// Create a node info with error to include in analysis
				nodeInfo := &GaleraClusterInfo{
					NodeIP:      nodeIP,
					StatusError: fmt.Sprintf("SSH connection failed: %v", err),
				}
				analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
				analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to connect to node %s: %v", nodeIP, err))
				analysis.IsCoherent = false
				progressPrint("      ❌ Connection failed: %v\n", err)
				continue
			}

			// Save the new connection info for this node if we got new credentials
			if newConnInfo != nil {
				sshPassword := ""
				if newConnInfo.HasPassword {
					sshPassword = newConnInfo.Password
				}
				err = config.setNodeCredentials(nodeIP, newConnInfo.Username, "", sshPassword, "", newConnInfo.UsedKeys)
				if err != nil {
					progressPrint("      ⚠️  Warning: Could not save credentials for node %s: %v\n", nodeIP, err)
				} else {
					if newConnInfo.HasPassword {
						progressPrint("      ✓ SSH password saved for node %s\n", nodeIP)
					}
					progressPrint("      ✓ SSH credentials saved for node %s\n", nodeIP)
				}
			}
		}

		// Verify we have a valid SSH client
		if sshClient == nil {
			// Create a node info with error to include in analysis
			nodeInfo := &GaleraClusterInfo{
				NodeIP:      nodeIP,
				StatusError: "SSH client is nil - connection failed",
			}
			analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
			analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("SSH client is nil for node %s", nodeIP))
			analysis.IsCoherent = false
			progressPrint("      ❌ SSH client is nil\n")
			continue
		}

		// Get cluster info from this node
		nodeInfo, err := getGaleraClusterInfo(sshClient, nodeIP)
		sshClient.Close()

		if err != nil {
			// Create a node info with error to include in analysis
			nodeInfo = &GaleraClusterInfo{
				NodeIP:      nodeIP,
				StatusError: fmt.Sprintf("Failed to get cluster info: %v", err),
			}
			analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
			analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to get cluster info from node %s: %v", nodeIP, err))
			analysis.IsCoherent = false
			progressPrint("      ❌ Failed to get cluster info: %v\n", err)
			continue
		}

		analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
		progressPrint("      ✓ Configuration retrieved\n")
	}

	// Analyze configuration coherence
	analysis.analyzeCoherence()

	return analysis, localhostNodeIP, nil
}

// analyzeCoherence analyzes the coherence of cluster configuration across nodes
func (a *ClusterAnalysis) analyzeCoherence() {
	if len(a.AllNodes) < 2 {
		return
	}

	reference := a.InitialNode

	for _, node := range a.AllNodes[1:] {
		// Check cluster name consistency
		if node.ClusterName != reference.ClusterName {
			a.ConfigErrors = append(a.ConfigErrors,
				fmt.Sprintf("Node %s has different cluster name: '%s' vs '%s'",
					node.NodeIP, node.ClusterName, reference.ClusterName))
			a.IsCoherent = false
		}

		// Check cluster address consistency
		if node.ClusterAddress != reference.ClusterAddress {
			a.ConfigErrors = append(a.ConfigErrors,
				fmt.Sprintf("Node %s has different cluster address: '%s' vs '%s'",
					node.NodeIP, node.ClusterAddress, reference.ClusterAddress))
			a.IsCoherent = false
		}

		// Check if node address matches expected IP
		if node.NodeAddress != "" && node.NodeAddress != node.NodeIP {
			// Check if it's just a different format (hostname vs IP)
			a.ConfigErrors = append(a.ConfigErrors,
				fmt.Sprintf("Node %s: wsrep_node_address (%s) differs from connection IP (%s)",
					node.NodeIP, node.NodeAddress, node.NodeIP))
		}
	}
}

// checkMySQLStatusOnAllNodes checks MySQL/MariaDB status on all nodes in the analysis
func checkMySQLStatusOnAllNodes(analysis *ClusterAnalysis, connInfo *SSHConnectionInfo, mysqlCreds *MySQLConnectionInfo, config *Config, localhostNodeIP string) error {
	for i, node := range analysis.AllNodes {
		progressPrint("   %d. %s - checking MySQL status...\n", i+1, node.NodeIP)

		// Skip nodes that already have connection errors
		if node.StatusError != "" && strings.Contains(node.StatusError, "SSH connection failed") {
			progressPrint("      ❌ Skipping MySQL check due to SSH connection failure: %s\n", node.StatusError)
			continue
		}

		// Check if this is localhost - use direct access instead of SSH
		// Consider both localhost references and the identified localhost IP
		if isLocalhost(node.NodeIP) || node.NodeIP == localhostNodeIP {
			progressPrint("      🏠 Using local MySQL connection for localhost\n")
			// Use nil SSH client for localhost - checkMySQLStatus will handle this
			checkMySQLStatus(nil, node.NodeIP, mysqlCreds, node)
		} else {
			// Connect to remote node using per-node credentials
			sshClient, _, err := createSSHConnectionWithNodeCredentials(node.NodeIP, config)
			if err != nil {
				node.StatusError = fmt.Sprintf("SSH connection failed: %v", err)
				progressPrint("      ❌ SSH connection failed: %v\n", err)
				continue
			}

			// Check MySQL status on remote node
			checkMySQLStatus(sshClient, node.NodeIP, mysqlCreds, node)
			sshClient.Close()
		}

		if node.MySQLResponding {
			progressPrint("      ✓ MySQL responding (Size: %d, Status: %s, Ready: %t, State: %s)\n",
				node.ClusterSize, node.ClusterStatus, node.IsReady, node.LocalStateComment)
		} else {
			progressPrint("      ❌ MySQL not responding: %s\n", node.StatusError)
		}
	}

	return nil
}
