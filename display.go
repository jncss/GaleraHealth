package main

import (
	"fmt"
	"strings"
)

// displayClusterInfo displays information about a single cluster node
func displayClusterInfo(info *GaleraClusterInfo) {
	fmt.Println("=== GALERA CLUSTER INFORMATION ===")
	fmt.Println()
	fmt.Printf("🏠 Node IP: %s\n", info.NodeIP)

	if info.ClusterName != "" {
		fmt.Printf("🏷️  Cluster Name: %s\n", info.ClusterName)
	} else {
		fmt.Println("⚠️  Cluster Name: Not configured")
	}

	if info.ClusterAddress != "" {
		fmt.Printf("📍 Cluster Address: %s\n", info.ClusterAddress)
	} else {
		fmt.Println("⚠️  Cluster Address: Not configured")
	}

	if info.NodeName != "" {
		fmt.Printf("🔖 Node Name: %s\n", info.NodeName)
	}

	if info.NodeAddress != "" {
		fmt.Printf("🌐 Node Address: %s\n", info.NodeAddress)
	}

	fmt.Println()
}

// displayClusterAnalysis displays cluster analysis results
func displayClusterAnalysis(analysis *ClusterAnalysis) {
	fmt.Println()
	fmt.Println("=== CLUSTER ANALYSIS RESULTS ===")
	fmt.Println()

	fmt.Printf("📊 Nodes analyzed: %d/%d\n", len(analysis.AllNodes), len(analysis.ClusterNodes))
	fmt.Printf("🎯 Cluster name: %s\n", analysis.InitialNode.ClusterName)

	fmt.Println()
	fmt.Println("📋 All nodes in cluster:")
	for i, node := range analysis.AllNodes {
		fmt.Printf("   %d. %s\n", i+1, node.NodeIP)
		fmt.Printf("      Cluster Name: %s\n", node.ClusterName)
		fmt.Printf("      Cluster Address: %s\n", node.ClusterAddress)
		if node.NodeName != "" {
			fmt.Printf("      Node Name: %s\n", node.NodeName)
		}
		if node.NodeAddress != "" {
			fmt.Printf("      Node Address: %s\n", node.NodeAddress)
		}
		fmt.Println()
	}

	if analysis.IsCoherent {
		fmt.Println("✅ CLUSTER CONFIGURATION IS COHERENT")
		fmt.Println("   All nodes have consistent configuration")
	} else {
		fmt.Println("❌ CLUSTER CONFIGURATION ISSUES DETECTED")
		fmt.Printf("   Found %d configuration errors:\n", len(analysis.ConfigErrors))
		for i, error := range analysis.ConfigErrors {
			fmt.Printf("   %d. %s\n", i+1, error)
		}

		// Check if all errors are localhost-related SSH issues
		localhostErrors := 0
		for _, error := range analysis.ConfigErrors {
			if strings.Contains(error, "initial connection was localhost") {
				localhostErrors++
			}
		}

		if localhostErrors > 0 && localhostErrors == len(analysis.ConfigErrors) {
			fmt.Println()
			fmt.Println("💡 TIP: To analyze all cluster nodes, run GaleraHealth with SSH access:")
			fmt.Printf("   galerahealth  # Enter a remote node IP instead of localhost\n")
			if len(analysis.ClusterNodes) > 1 {
				fmt.Printf("   # Example: Enter %s when prompted for node IP\n", analysis.ClusterNodes[1])
			}
		}
	}
}

// displayClusterAnalysisWithMySQL displays cluster analysis results including MySQL status
func displayClusterAnalysisWithMySQL(analysis *ClusterAnalysis) {
	fmt.Println()
	fmt.Println("=== CLUSTER ANALYSIS WITH MYSQL STATUS ===")
	fmt.Println()

	fmt.Printf("📊 Nodes analyzed: %d/%d\n", len(analysis.AllNodes), len(analysis.ClusterNodes))
	fmt.Printf("🎯 Cluster name: %s\n", analysis.InitialNode.ClusterName)

	// Display cluster health summary with MySQL status
	fmt.Println()
	fmt.Println("🏥 Cluster Health Summary:")

	respondingNodes := 0
	readyNodes := 0
	primaryNodes := 0
	syncedNodes := 0

	for _, node := range analysis.AllNodes {
		if node.MySQLResponding {
			respondingNodes++
			if node.IsReady {
				readyNodes++
			}
			if node.ClusterStatus == "Primary" {
				primaryNodes++
			}
			if node.LocalStateComment == "Synced" {
				syncedNodes++
			}
		}
	}

	totalNodes := len(analysis.AllNodes)

	// MySQL responding status
	respondingIcon := "✅"
	if respondingNodes != totalNodes {
		respondingIcon = "⚠️"
	}
	fmt.Printf("   %s MySQL/MariaDB responding: %d/%d nodes\n", respondingIcon, respondingNodes, totalNodes)

	if respondingNodes > 0 {
		// Ready status
		readyIcon := "✅"
		if readyNodes != respondingNodes {
			readyIcon = "⚠️"
		}
		fmt.Printf("   %s Nodes ready: %d/%d responding nodes\n", readyIcon, readyNodes, respondingNodes)

		// Cluster status (Primary)
		primaryIcon := "✅"
		if primaryNodes != respondingNodes {
			primaryIcon = "⚠️"
		}
		fmt.Printf("   %s Primary status: %d/%d responding nodes\n", primaryIcon, primaryNodes, respondingNodes)

		// Synced status
		syncedIcon := "✅"
		if syncedNodes != respondingNodes {
			syncedIcon = "⚠️"
		}
		fmt.Printf("   %s Synced state: %d/%d responding nodes\n", syncedIcon, syncedNodes, respondingNodes)
	}

	fmt.Println()

	// Display all nodes information with MySQL status
	fmt.Println("📋 Node Details with MySQL Status:")
	for i, node := range analysis.AllNodes {
		fmt.Printf("   %d. %s\n", i+1, node.NodeIP)
		fmt.Printf("      Cluster Name: %s\n", node.ClusterName)
		fmt.Printf("      Cluster Address: %s\n", node.ClusterAddress)
		if node.NodeName != "" {
			fmt.Printf("      Node Name: %s\n", node.NodeName)
		}
		if node.NodeAddress != "" {
			fmt.Printf("      Node Address: %s\n", node.NodeAddress)
		}

		// Display runtime status
		if node.MySQLResponding {
			fmt.Printf("      MySQL/MariaDB: ✅ Responding\n")
			if node.ClusterSize > 0 {
				fmt.Printf("      Cluster Size: %d\n", node.ClusterSize)
			}
			if node.ClusterStatus != "" {
				statusIcon := "✅"
				if node.ClusterStatus != "Primary" {
					statusIcon = "⚠️"
				}
				fmt.Printf("      Cluster Status: %s %s\n", statusIcon, node.ClusterStatus)
			}
			readyIcon := "✅"
			if !node.IsReady {
				readyIcon = "❌"
			}
			fmt.Printf("      Node Ready: %s %t\n", readyIcon, node.IsReady)
			if node.LocalStateComment != "" {
				stateIcon := "✅"
				if node.LocalStateComment != "Synced" {
					stateIcon = "⚠️"
				}
				fmt.Printf("      Local State: %s %s\n", stateIcon, node.LocalStateComment)
			}
		} else {
			fmt.Printf("      MySQL/MariaDB: ❌ Not responding\n")
			if node.StatusError != "" {
				fmt.Printf("      Error: %s\n", node.StatusError)
			}
		}
		fmt.Println()
	}

	// Display coherence status
	if analysis.IsCoherent {
		fmt.Println("✅ CLUSTER CONFIGURATION IS COHERENT")
		fmt.Println("   All nodes have consistent configuration")
	} else {
		fmt.Println("❌ CLUSTER CONFIGURATION ISSUES DETECTED")
		fmt.Printf("   Found %d configuration errors:\n", len(analysis.ConfigErrors))
		for i, error := range analysis.ConfigErrors {
			fmt.Printf("   %d. %s\n", i+1, error)
		}
	}
}

// displayClusterSummary displays a final summary of the cluster health status
func displayClusterSummary(analysis *ClusterAnalysis) {
	logMinimal("")
	logMinimal("=== CLUSTER HEALTH SUMMARY ===")
	logMinimal("")

	totalNodes := len(analysis.AllNodes)
	issues := []string{}
	warnings := []string{}

	// Check configuration coherence
	if !analysis.IsCoherent {
		issues = append(issues, fmt.Sprintf("Incoherent configuration (%d errors)", len(analysis.ConfigErrors)))
	}

	// Check MySQL/MariaDB status if available
	respondingNodes := 0
	readyNodes := 0
	primaryNodes := 0
	syncedNodes := 0
	hasMySQLData := false

	for _, node := range analysis.AllNodes {
		// Check if we have any MySQL data (either responding or error status)
		if node.MySQLResponding || node.StatusError != "" {
			hasMySQLData = true
		}

		if node.MySQLResponding {
			respondingNodes++
			if node.IsReady {
				readyNodes++
			}
			if node.ClusterStatus == "Primary" {
				primaryNodes++
			}
			if node.LocalStateComment == "Synced" {
				syncedNodes++
			}
		}
	}

	// MySQL/MariaDB issues
	if hasMySQLData {
		if respondingNodes != totalNodes {
			issues = append(issues, fmt.Sprintf("MySQL/MariaDB not responding on %d/%d nodes", totalNodes-respondingNodes, totalNodes))
		}

		if respondingNodes > 0 {
			if readyNodes != respondingNodes {
				issues = append(issues, fmt.Sprintf("Nodes not ready: %d/%d", respondingNodes-readyNodes, respondingNodes))
			}
			if primaryNodes != respondingNodes {
				if primaryNodes == 0 {
					issues = append(issues, "No nodes in Primary state")
				} else {
					warnings = append(warnings, fmt.Sprintf("Only %d/%d nodes in Primary state", primaryNodes, respondingNodes))
				}
			}
			if syncedNodes != respondingNodes {
				issues = append(issues, fmt.Sprintf("Nodes not synchronized: %d/%d", respondingNodes-syncedNodes, respondingNodes))
			}
		}
	}

	// Display summary
	if len(issues) == 0 && len(warnings) == 0 {
		logMinimal("🎉 GALERA CLUSTER IN PERFECT HEALTH")
		logMinimal("   ✅ Configuration coherent across all nodes")
		if hasMySQLData {
			logMinimal("   ✅ All MySQL/MariaDB nodes responding correctly")
			logMinimal("   ✅ All nodes synchronized and ready")
			logMinimal("   ✅ Cluster in Primary state")
		}
		logMinimal("")
		logMinimal("📊 Total nodes: %d", totalNodes)
		if hasMySQLData {
			logMinimal("🔗 Active nodes: %d/%d", respondingNodes, totalNodes)
		}
	} else {
		// Display problems
		if len(issues) > 0 {
			logMinimal("❌ CRITICAL ISSUES DETECTED:")
			for i, issue := range issues {
				logMinimal("   %d. %s", i+1, issue)
			}
			logMinimal("")
		}

		// Display warnings
		if len(warnings) > 0 {
			logMinimal("⚠️  WARNINGS:")
			for i, warning := range warnings {
				logMinimal("   %d. %s", i+1, warning)
			}
			logMinimal("")
		}

		// Status summary
		logMinimal("📊 STATUS SUMMARY:")
		logMinimal("   🏠 Total nodes: %d", totalNodes)
		logMinimal("   ⚙️  Configuration coherent: %s", getStatusIcon(analysis.IsCoherent))

		if hasMySQLData {
			logMinimal("   🔗 MySQL/MariaDB active: %d/%d %s", respondingNodes, totalNodes, getStatusIcon(respondingNodes == totalNodes))
			if respondingNodes > 0 {
				logMinimal("   ✅ Nodes ready: %d/%d %s", readyNodes, respondingNodes, getStatusIcon(readyNodes == respondingNodes))
				logMinimal("   🎯 Primary state: %d/%d %s", primaryNodes, respondingNodes, getStatusIcon(primaryNodes == respondingNodes))
				logMinimal("   🔄 Nodes synchronized: %d/%d %s", syncedNodes, respondingNodes, getStatusIcon(syncedNodes == respondingNodes))
			}
		} else {
			logMinimal("   🔗 MySQL/MariaDB: Not checked")
		}

		logMinimal("")
		if len(issues) > 0 {
			logMinimal("🚨 ACTION REQUIRED: Cluster has issues that need immediate attention")
		} else {
			logMinimal("⚠️  ATTENTION: Cluster is functional but has minor warnings")
		}
	}

	logMinimal("")
} // getStatusIcon returns appropriate icon for boolean status
func getStatusIcon(status bool) string {
	if status {
		return "✅"
	}
	return "❌"
}
