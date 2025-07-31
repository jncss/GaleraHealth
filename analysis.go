package main

import (
	"fmt"
	"strings"
)

// performClusterAnalysis analyzes cluster coherence across all nodes
func performClusterAnalysis(initialNode *GaleraClusterInfo, connInfo *SSHConnectionInfo) (*ClusterAnalysis, error) {
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

	// Check each node in the cluster
	for i, nodeIP := range analysis.ClusterNodes {
		if nodeIP == initialNode.NodeIP {
			// Skip initial node, already analyzed
			fmt.Printf("   %d. %s (initial node - already analyzed)\n", i+1, nodeIP)
			continue
		}

		fmt.Printf("   %d. %s - connecting...\n", i+1, nodeIP)

		// Connect to the node using saved connection info
		sshClient, err := createSSHConnectionWithInfo(nodeIP, connInfo)
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
func checkMySQLStatusOnAllNodes(analysis *ClusterAnalysis, connInfo *SSHConnectionInfo, mysqlCreds *MySQLConnectionInfo) error {
	for i, node := range analysis.AllNodes {
		fmt.Printf("   %d. %s - checking MySQL status...\n", i+1, node.NodeIP)

		// Connect to the node
		sshClient, err := createSSHConnectionWithInfo(node.NodeIP, connInfo)
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
