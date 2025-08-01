package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// NodeState represents the state of a Galera node
type NodeState struct {
	IP          string
	IsUp        bool
	SeqNo       int64
	LatestIDB   time.Time
	HasGrastate bool
}

// ClusterState represents the overall state of the cluster
type ClusterState struct {
	Nodes    []NodeState
	AllDown  bool
	SomeDown bool
	AllUp    bool
}

// analyzeClusterState analyzes the current state of all cluster nodes
func analyzeClusterState(clusterIPs []string, config *Config) (*ClusterState, error) {
	logNormal("🔍 Analyzing cluster state for recovery assessment...")

	state := &ClusterState{
		Nodes: make([]NodeState, len(clusterIPs)),
	}

	upCount := 0

	for i, ip := range clusterIPs {
		nodeState := NodeState{IP: ip}

		// Check if MySQL/MariaDB is running on this node
		logVerbose("Checking MySQL/MariaDB status on node %s", ip)

		cmd := "systemctl is-active mariadb mysqld mysql 2>/dev/null | head -1"
		output, err := executeCommandOnNode(ip, cmd, config)
		nodeState.IsUp = (err == nil && strings.TrimSpace(output) == "active")

		if nodeState.IsUp {
			upCount++
			logVerbose("✅ Node %s: MySQL/MariaDB is running", ip)
		} else {
			logVerbose("❌ Node %s: MySQL/MariaDB is not running", ip)

			// Get seqno from grastate.dat for down nodes
			seqno, hasGrastate := getNodeSeqNo(ip, config)
			nodeState.SeqNo = seqno
			nodeState.HasGrastate = hasGrastate

			// Get latest IDB timestamp for fallback method
			latestIDB := getLatestIDBTimestamp(ip, config)
			nodeState.LatestIDB = latestIDB
		}

		state.Nodes[i] = nodeState
	}

	// Determine cluster state
	state.AllUp = (upCount == len(clusterIPs))
	state.AllDown = (upCount == 0)
	state.SomeDown = (upCount > 0 && upCount < len(clusterIPs))

	logNormal("📊 Cluster state: %d/%d nodes running", upCount, len(clusterIPs))

	return state, nil
}

// getNodeSeqNo retrieves the seqno from grastate.dat
func getNodeSeqNo(ip string, config *Config) (int64, bool) {
	logVerbose("🔍 Reading grastate.dat seqno on node %s", ip)

	cmd := "cat /var/lib/mysql/grastate.dat 2>/dev/null | grep 'seqno:' | awk '{print $2}'"
	output, err := executeCommandOnNode(ip, cmd, config)
	if err != nil {
		logVerbose("⚠️ Node %s: Could not read grastate.dat", ip)
		return -1, false
	}

	seqnoStr := strings.TrimSpace(output)
	seqno, err := strconv.ParseInt(seqnoStr, 10, 64)
	if err != nil {
		logVerbose("⚠️ Node %s: Invalid seqno format: %s", ip, seqnoStr)
		return -1, false
	}

	logVerbose("📋 Node %s: seqno = %d", ip, seqno)
	return seqno, true
}

// getLatestIDBTimestamp finds the most recent .ibd file timestamp
func getLatestIDBTimestamp(ip string, config *Config) time.Time {
	logVerbose("🔍 Finding latest .ibd file on node %s", ip)

	cmd := "find /var/lib/mysql -name '*.ibd' -type f -printf '%T@ %p\\n' 2>/dev/null | sort -nr | head -1 | awk '{print $1}'"
	output, err := executeCommandOnNode(ip, cmd, config)
	if err != nil {
		logVerbose("⚠️ Node %s: Could not find .ibd files", ip)
		return time.Time{}
	}

	timestampStr := strings.TrimSpace(output)
	if timestampStr == "" {
		logVerbose("⚠️ Node %s: No .ibd files found", ip)
		return time.Time{}
	}

	timestamp, err := strconv.ParseFloat(timestampStr, 64)
	if err != nil {
		logVerbose("⚠️ Node %s: Invalid timestamp format: %s", ip, timestampStr)
		return time.Time{}
	}

	t := time.Unix(int64(timestamp), 0)
	logVerbose("📋 Node %s: latest .ibd timestamp = %s", ip, t.Format("2006-01-02 15:04:05"))
	return t
}

// selectBootstrapNode chooses the best node to bootstrap the cluster
func selectBootstrapNode(state *ClusterState) (string, string, error) {
	downNodes := []NodeState{}
	for _, node := range state.Nodes {
		if !node.IsUp {
			downNodes = append(downNodes, node)
		}
	}

	if len(downNodes) == 0 {
		return "", "", fmt.Errorf("no down nodes to select from")
	}

	// Method 1: Use highest seqno (if all nodes have valid seqno and none is -1)
	validSeqnos := true
	maxSeqno := int64(-2) // Start below -1

	for _, node := range downNodes {
		if !node.HasGrastate || node.SeqNo == -1 {
			validSeqnos = false
			break
		}
		if node.SeqNo > maxSeqno {
			maxSeqno = node.SeqNo
		}
	}

	if validSeqnos && maxSeqno >= 0 {
		// Find node with highest seqno
		for _, node := range downNodes {
			if node.SeqNo == maxSeqno {
				method := fmt.Sprintf("Selected node %s based on highest seqno (%d)", node.IP, maxSeqno)
				logReport("🎯 %s", method)
				return node.IP, method, nil
			}
		}
	}

	// Method 2: Use latest .ibd file timestamp
	logReport("⚠️ Cannot use seqno method (some nodes have seqno = -1 or missing grastate.dat)")
	logReport("🔄 Using fallback method: latest .ibd file timestamp")

	var bestNode NodeState
	var latestTime time.Time

	for _, node := range downNodes {
		if node.LatestIDB.After(latestTime) {
			latestTime = node.LatestIDB
			bestNode = node
		}
	}

	if latestTime.IsZero() {
		return "", "", fmt.Errorf("could not determine best node using any method")
	}

	method := fmt.Sprintf("Selected node %s based on latest .ibd file timestamp (%s)",
		bestNode.IP, latestTime.Format("2006-01-02 15:04:05"))
	logReport("🎯 %s", method)

	return bestNode.IP, method, nil
}

// performClusterRecovery attempts to recover the cluster
func performClusterRecovery(state *ClusterState, config *Config) error {
	if state.AllUp {
		logReport("✅ All cluster nodes are already running - no recovery needed")
		return nil
	}

	if state.SomeDown {
		logReport("⚠️ Some cluster nodes are down - attempting to start them...")
		return startDownNodes(state, config)
	}

	if state.AllDown {
		logReport("❌ All cluster nodes are down - attempting full cluster recovery...")
		return bootstrapCluster(state, config)
	}

	return nil
}

// startDownNodes attempts to start the down nodes
func startDownNodes(state *ClusterState, config *Config) error {
	for _, node := range state.Nodes {
		if !node.IsUp {
			logReport("🔄 Attempting to start MySQL/MariaDB on node %s...", node.IP)

			if !askUserPermission(fmt.Sprintf("Start MySQL/MariaDB service on node %s", node.IP)) {
				logReport("⏭️ Skipping node %s (user declined)", node.IP)
				continue
			}

			err := startMySQLService(node.IP, config)
			if err != nil {
				logReport("❌ Failed to start MySQL/MariaDB on node %s: %v", node.IP, err)
			} else {
				logReport("✅ Successfully started MySQL/MariaDB on node %s", node.IP)
			}
		}
	}
	return nil
}

// bootstrapCluster performs full cluster bootstrap
func bootstrapCluster(state *ClusterState, config *Config) error {
	// Select bootstrap node
	bootstrapIP, method, err := selectBootstrapNode(state)
	if err != nil {
		return fmt.Errorf("failed to select bootstrap node: %v", err)
	}

	logReport("📋 Bootstrap node selection method: %s", method)

	if !askUserPermission(fmt.Sprintf("Bootstrap the cluster using node %s", bootstrapIP)) {
		return fmt.Errorf("user declined cluster bootstrap")
	}

	// Bootstrap the selected node
	logReport("🚀 Bootstrapping cluster on node %s...", bootstrapIP)
	err = bootstrapNode(bootstrapIP, config)
	if err != nil {
		return fmt.Errorf("failed to bootstrap node %s: %v", bootstrapIP, err)
	}

	logReport("✅ Successfully bootstrapped cluster on node %s", bootstrapIP)

	// Wait a bit for the bootstrap node to stabilize
	logReport("⏳ Waiting for bootstrap node to stabilize...")
	time.Sleep(5 * time.Second)

	// Start other nodes
	for _, node := range state.Nodes {
		if node.IP != bootstrapIP && !node.IsUp {
			logReport("🔄 Attempting to start MySQL/MariaDB on node %s...", node.IP)

			if !askUserPermission(fmt.Sprintf("Start MySQL/MariaDB service on node %s", node.IP)) {
				logReport("⏭️ Skipping node %s (user declined)", node.IP)
				continue
			}

			err := startMySQLService(node.IP, config)
			if err != nil {
				logReport("❌ Failed to start MySQL/MariaDB on node %s: %v", node.IP, err)
			} else {
				logReport("✅ Successfully started MySQL/MariaDB on node %s", node.IP)
				// Wait a bit between node starts
				time.Sleep(2 * time.Second)
			}
		}
	}

	return nil
}

// bootstrapNode performs galera_new_cluster on the specified node
func bootstrapNode(ip string, config *Config) error {
	cmd := "galera_new_cluster"
	_, err := executeCommandOnNode(ip, cmd, config)
	return err
}

// startMySQLService starts MySQL/MariaDB service on the specified node
func startMySQLService(ip string, config *Config) error {
	services := []string{"mariadb", "mysql", "mysqld"}

	for _, service := range services {
		cmd := fmt.Sprintf("systemctl start %s", service)
		_, err := executeCommandOnNode(ip, cmd, config)
		if err == nil {
			return nil // Successfully started with this service name
		}
	}

	return fmt.Errorf("failed to start MySQL/MariaDB with any service name")
}

// askUserPermission asks the user for permission to perform an action
// Recovery actions always require explicit user confirmation, even in -y mode
func askUserPermission(action string) bool {
	// For recovery actions, we ALWAYS ask for permission, even with -y flag
	// This is because recovery actions can be destructive and should be explicitly confirmed
	fmt.Printf("❓ Do you want to %s? (y/N): ", action)

	// Read input directly using bufio scanner
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	// Default to "no" if empty response
	if response == "" {
		response = "n"
	}

	return response == "y" || response == "yes"
}

// executeCommandOnNode executes a command on a specific node using proper SSH authentication
func executeCommandOnNode(ip string, command string, config *Config) (string, error) {
	// Check if this is a local command
	if ip == "localhost" || ip == "127.0.0.1" {
		return executeLocalCommand(command)
	}

	// For remote nodes, use the existing SSH connection logic
	sshClient, _, err := createSSHConnectionWithNodeCredentials(ip, config)
	if err != nil {
		return "", fmt.Errorf("failed to establish SSH connection to %s: %v", ip, err)
	}
	defer sshClient.Close()

	// Execute the command using the SSH client
	output, err := sshClient.executeCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command on %s: %v", ip, err)
	}

	return output, nil
}

// buildSSHCommand builds the SSH command for a remote node (legacy function, kept for compatibility)
func buildSSHCommand(ip string, config *Config) string {
	creds := config.getNodeCredentials(ip)
	if creds == nil {
		return fmt.Sprintf("ssh root@%s", ip)
	}

	return fmt.Sprintf("ssh %s@%s", creds.SSHUsername, ip)
}

// executeCommand executes a command either locally or remotely (legacy function)
func executeCommand(cmd string) (string, error) {
	// Check if this is a local command (doesn't start with ssh)
	if !strings.HasPrefix(cmd, "ssh ") {
		return executeLocalCommand(cmd)
	}

	// For SSH commands, use executeLocalCommand as well since we're building the full ssh command
	return executeLocalCommand(cmd)
}
