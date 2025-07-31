package main

import (
	"fmt"
	"strings"
)

// performClusterAnalysis analyzes cluster coherence across all nodes
func performClusterAnalysis(initialNode *GaleraClusterInfo, connInfo *SSHConnectionInfo, config *Config) (*ClusterAnalysis, error) {
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
		return nil, fmt.Errorf("no cluster nodes found in wsrep_cluster_address")
	}

	fmt.Printf("📋 Found %d nodes in cluster configuration\n", len(analysis.ClusterNodes))

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
				fmt.Printf("   %d. %s (initial node - already analyzed)\n", i+1, nodeIP)
			} else if isLocalhost(nodeIP) {
				fmt.Printf("   %d. %s (localhost - skipping SSH)\n", i+1, nodeIP)
			} else if nodeIP == localhostNodeIP {
				fmt.Printf("   %d. %s (this is localhost %s - already analyzed)\n", i+1, nodeIP, initialNode.NodeIP)
			}
			continue
		}

		fmt.Printf("   %d. %s - connecting...\n", i+1, nodeIP)

		// Check if we have valid SSH connection info
		var sshClient *SSHClient
		var err error

		if connInfo.Username == "local" {
			// Initial connection was localhost, cannot connect to remote nodes without SSH credentials
			analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Cannot connect to remote node %s: initial connection was localhost (no SSH credentials available)", nodeIP))
			analysis.IsCoherent = false
			fmt.Printf("      ⚠️  Skipped: initial connection was localhost, no SSH credentials for remote nodes\n")
			continue
		} else {
			// Use per-node credentials for connection
			sshClient, newConnInfo, err := createSSHConnectionWithNodeCredentials(nodeIP, config)
			if err != nil {
				analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to connect to node %s: %v", nodeIP, err))
				analysis.IsCoherent = false
				fmt.Printf("      ❌ Connection failed: %v\n", err)
				continue
			}
			sshClient.Close() // We'll reopen it when needed
			
			// Save the new connection info for this node if we got new credentials
			if newConnInfo != nil {
				err = config.setNodeCredentials(nodeIP, newConnInfo.Username, "", "", "", newConnInfo.UsedKeys)
				if err != nil {
					fmt.Printf("      ⚠️  Warning: Could not save credentials for node %s: %v\n", nodeIP, err)
				}
			}
			
			// Now create connection using saved info
			sshClient, _, err = createSSHConnectionWithNodeCredentials(nodeIP, config)
		}

		if err != nil {
			analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to connect to node %s: %v", nodeIP, err))
			analysis.IsCoherent = false
			fmt.Printf("      ❌ Connection failed: %v\n", err)
			continue
		}

		// Get cluster info from this node
		nodeInfo, err := getGaleraClusterInfo(sshClient, nodeIP)
		sshClient.Close()

		if err != nil {
			analysis.ConfigErrors = append(analysis.ConfigErrors, fmt.Sprintf("Failed to get cluster info from node %s: %v", nodeIP, err))
			analysis.IsCoherent = false
			fmt.Printf("      ❌ Failed to get cluster info: %v\n", err)
			continue
		}

		analysis.AllNodes = append(analysis.AllNodes, nodeInfo)
		fmt.Printf("      ✓ Configuration retrieved\n")
	}

	// Analyze configuration coherence
	analysis.analyzeCoherence()

	return analysis, nil
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
func checkMySQLStatusOnAllNodes(analysis *ClusterAnalysis, connInfo *SSHConnectionInfo, mysqlCreds *MySQLConnectionInfo, config *Config) error {
	for i, node := range analysis.AllNodes {
		fmt.Printf("   %d. %s - checking MySQL status...\n", i+1, node.NodeIP)

		// Connect to the node using per-node credentials
		sshClient, _, err := createSSHConnectionWithNodeCredentials(node.NodeIP, config)
		if err != nil {
			node.StatusError = fmt.Sprintf("SSH connection failed: %v", err)
			fmt.Printf("      ❌ SSH connection failed: %v\n", err)
			continue
		}

		// Check MySQL status
		checkMySQLStatus(sshClient, node.NodeIP, mysqlCreds, node)
		sshClient.Close()

		if node.MySQLResponding {
			fmt.Printf("      ✓ MySQL responding (Size: %d, Status: %s, Ready: %t, State: %s)\n",
				node.ClusterSize, node.ClusterStatus, node.IsReady, node.LocalStateComment)
		} else {
			fmt.Printf("      ❌ MySQL not responding: %s\n", node.StatusError)
		}
	}

	return nil
}
