package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// getGaleraClusterInfo retrieves Galera cluster configuration from a node
func getGaleraClusterInfo(sshClient *SSHClient, nodeIP string) (*GaleraClusterInfo, error) {
	clusterInfo := &GaleraClusterInfo{
		NodeIP: nodeIP,
	}

	logNormal("🔍 Searching for cluster information...")

	// Search recursively for all .cnf files in /etc/mysql and also check /etc/my.cnf
	logVerbose("📁 Searching for configuration files...")

	var foundConfigs []string

	// Find all .cnf files recursively in /etc/mysql
	output, err := sshClient.executeCommand("find /etc/mysql -name '*.cnf' -type f 2>/dev/null")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				foundConfigs = append(foundConfigs, line)
			}
		}
	}

	// Also check /etc/my.cnf
	output, err = sshClient.executeCommand("test -f /etc/my.cnf && echo '/etc/my.cnf' || echo ''")
	if err == nil {
		line := strings.TrimSpace(output)
		if line != "" {
			foundConfigs = append(foundConfigs, line)
		}
	}

	if len(foundConfigs) == 0 {
		return nil, fmt.Errorf("no MySQL configuration files found in /etc/mysql or /etc/my.cnf")
	}

	logVerbose("📁 Configuration files found: %d files", len(foundConfigs))
	for _, config := range foundConfigs {
		logDebug("   - %s", config)
	}

	// Search for wsrep variables in all found files
	for _, configPath := range foundConfigs {
		logVerbose("   Analyzing %s...", configPath)

		// Read file content
		content, err := sshClient.executeCommand(fmt.Sprintf("cat %s", configPath))
		if err != nil {
			logVerbose("   ⚠️  Error reading %s: %v", configPath, err)
			continue
		}

		// Search for wsrep_cluster_name
		if clusterInfo.ClusterName == "" {
			clusterName := extractConfigValue(content, "wsrep_cluster_name")
			if clusterName != "" {
				clusterInfo.ClusterName = clusterName
				logVerbose("   ✓ wsrep_cluster_name found in %s", configPath)
			}
		}

		// Search for wsrep_cluster_address
		if clusterInfo.ClusterAddress == "" {
			clusterAddress := extractConfigValue(content, "wsrep_cluster_address")
			if clusterAddress != "" {
				clusterInfo.ClusterAddress = clusterAddress
				logVerbose("   ✓ wsrep_cluster_address found in %s", configPath)
			}
		}

		// Search for wsrep_node_name
		if clusterInfo.NodeName == "" {
			nodeName := extractConfigValue(content, "wsrep_node_name")
			if nodeName != "" {
				clusterInfo.NodeName = nodeName
				logVerbose("   ✓ wsrep_node_name found in %s", configPath)
			}
		}

		// Search for wsrep_node_address
		if clusterInfo.NodeAddress == "" {
			nodeAddress := extractConfigValue(content, "wsrep_node_address")
			if nodeAddress != "" {
				clusterInfo.NodeAddress = nodeAddress
				logVerbose("   ✓ wsrep_node_address found in %s", configPath)
			}
		}
	}

	// Also try to get information from MySQL runtime variables
	logVerbose("🔍 Checking MySQL runtime variables...")
	runtimeInfo, err := getRuntimeMySQLInfo(sshClient)
	if err == nil {
		if clusterInfo.ClusterName == "" && runtimeInfo.ClusterName != "" {
			clusterInfo.ClusterName = runtimeInfo.ClusterName
			logVerbose("   ✓ wsrep_cluster_name obtained from runtime variables")
		}
		if clusterInfo.ClusterAddress == "" && runtimeInfo.ClusterAddress != "" {
			clusterInfo.ClusterAddress = runtimeInfo.ClusterAddress
			logVerbose("   ✓ wsrep_cluster_address obtained from runtime variables")
		}
		if clusterInfo.NodeName == "" && runtimeInfo.NodeName != "" {
			clusterInfo.NodeName = runtimeInfo.NodeName
			logVerbose("   ✓ wsrep_node_name obtained from runtime variables")
		}
		if clusterInfo.NodeAddress == "" && runtimeInfo.NodeAddress != "" {
			clusterInfo.NodeAddress = runtimeInfo.NodeAddress
			logVerbose("   ✓ wsrep_node_address obtained from runtime variables")
		}
	}

	return clusterInfo, nil
}

// extractConfigValue extracts a configuration value from file content
func extractConfigValue(content, key string) string {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Search for the key
		if strings.HasPrefix(line, key) {
			// Extract the value after = or space
			re := regexp.MustCompile(fmt.Sprintf(`%s\s*=\s*(.+)`, regexp.QuoteMeta(key)))
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				value := strings.TrimSpace(matches[1])
				// Remove quotes if present
				value = strings.Trim(value, "\"'")
				return value
			}
		}
	}

	return ""
}

// getRuntimeMySQLInfo gets Galera information from MySQL runtime variables
func getRuntimeMySQLInfo(sshClient *SSHClient) (*GaleraClusterInfo, error) {
	// Try to get information from runtime variables
	queries := []string{
		"mysql -e \"SHOW VARIABLES LIKE 'wsrep_cluster_name';\"",
		"mysql -e \"SHOW VARIABLES LIKE 'wsrep_cluster_address';\"",
		"mysql -e \"SHOW VARIABLES LIKE 'wsrep_node_name';\"",
		"mysql -e \"SHOW VARIABLES LIKE 'wsrep_node_address';\"",
	}

	info := &GaleraClusterInfo{}

	for _, query := range queries {
		output, err := sshClient.executeCommand(query)
		if err != nil {
			continue // If MySQL is not accessible, continue
		}

		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "wsrep_cluster_name") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					info.ClusterName = parts[1]
				}
			}
			if strings.Contains(line, "wsrep_cluster_address") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					info.ClusterAddress = strings.Join(parts[1:], " ")
				}
			}
			if strings.Contains(line, "wsrep_node_name") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					info.NodeName = parts[1]
				}
			}
			if strings.Contains(line, "wsrep_node_address") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					info.NodeAddress = parts[1]
				}
			}
		}
	}

	return info, nil
}

// checkMySQLStatus checks MySQL/MariaDB status on a node
func checkMySQLStatus(sshClient *SSHClient, nodeIP string, mysqlCreds *MySQLConnectionInfo, info *GaleraClusterInfo) {
	// First, check if MySQL/MariaDB service is running
	serviceCheck, _ := sshClient.executeCommand("systemctl is-active mysql mariadb 2>/dev/null || service mysql status 2>/dev/null || service mariadb status 2>/dev/null")
	if !strings.Contains(serviceCheck, "active") && !strings.Contains(serviceCheck, "running") {
		// Get detailed service status
		serviceStatus, _ := sshClient.executeCommand("systemctl status mysql mariadb 2>/dev/null | head -10")
		info.StatusError = fmt.Sprintf("MySQL/MariaDB service is not running. Status: %s", strings.TrimSpace(serviceStatus))

		// Provide suggestions for starting the service
		suggestions := getSuggestionsForInactiveService(sshClient)
		if suggestions != "" {
			info.StatusError += fmt.Sprintf(". Suggestions: %s", suggestions)
		}
		return
	}

	// Try different connection methods
	var mysqlCmd string
	if mysqlCreds.Password != "" {
		// Try TCP connection first (more reliable for remote checks)
		mysqlCmd = fmt.Sprintf("mysql -h 127.0.0.1 -P 3306 -u %s -p'%s'", mysqlCreds.Username, mysqlCreds.Password)
	} else {
		mysqlCmd = fmt.Sprintf("mysql -h 127.0.0.1 -P 3306 -u %s", mysqlCreds.Username)
	}

	// Test basic connectivity first
	testCmd := fmt.Sprintf("%s -e \"SELECT 1;\" 2>&1", mysqlCmd)
	output, err := sshClient.executeCommand(testCmd)
	if err != nil || strings.Contains(output, "ERROR") {
		// If TCP fails, try socket connection
		if mysqlCreds.Password != "" {
			mysqlCmd = fmt.Sprintf("mysql -u %s -p'%s'", mysqlCreds.Username, mysqlCreds.Password)
		} else {
			mysqlCmd = fmt.Sprintf("mysql -u %s", mysqlCreds.Username)
		}

		testCmd = fmt.Sprintf("%s -e \"SELECT 1;\" 2>&1", mysqlCmd)
		output, err = sshClient.executeCommand(testCmd)
		if err != nil || strings.Contains(output, "ERROR") {
			// Get diagnostic information
			diagnostic := diagnoseMySQL(sshClient, nodeIP)
			info.StatusError = fmt.Sprintf("MySQL connection failed. Error: %s. Diagnostic: %s", strings.TrimSpace(output), diagnostic)
			return
		}
	}

	// Check cluster size
	cmd := fmt.Sprintf("%s -e \"SHOW STATUS LIKE 'wsrep_cluster_size';\" 2>&1", mysqlCmd)
	output, err = sshClient.executeCommand(cmd)
	if err != nil || strings.Contains(output, "ERROR") {
		info.StatusError = fmt.Sprintf("Failed to get cluster size. Error: %s", strings.TrimSpace(output))
		return
	}

	if parseClusterSize(output, info) {
		info.MySQLResponding = true

		// Check cluster status
		cmd = fmt.Sprintf("%s -e \"SHOW STATUS LIKE 'wsrep_cluster_status';\" 2>&1", mysqlCmd)
		if output, err := sshClient.executeCommand(cmd); err == nil && !strings.Contains(output, "ERROR") {
			parseClusterStatus(output, info)
		}

		// Check if node is ready
		cmd = fmt.Sprintf("%s -e \"SHOW STATUS LIKE 'wsrep_ready';\" 2>&1", mysqlCmd)
		if output, err := sshClient.executeCommand(cmd); err == nil && !strings.Contains(output, "ERROR") {
			parseReadyStatus(output, info)
		}

		// Check local state comment
		cmd = fmt.Sprintf("%s -e \"SHOW STATUS LIKE 'wsrep_local_state_comment';\" 2>&1", mysqlCmd)
		if output, err := sshClient.executeCommand(cmd); err == nil && !strings.Contains(output, "ERROR") {
			parseLocalStateComment(output, info)
		}
	} else {
		info.StatusError = "Could not retrieve cluster size - node may not be part of Galera cluster"
	}
}

// diagnoseMySQL provides diagnostic information for MySQL connection issues
func diagnoseMySQL(sshClient *SSHClient, nodeIP string) string {
	var diagnostic []string

	// Check if MySQL/MariaDB is installed
	checkInstalled, _ := sshClient.executeCommand("which mysql mysqld mariadb 2>/dev/null")
	if checkInstalled == "" {
		diagnostic = append(diagnostic, "MySQL/MariaDB client not found - may not be installed")
	}

	// Check service status with more detail
	serviceStatus, _ := sshClient.executeCommand("systemctl status mysql mariadb 2>/dev/null | head -3")
	if serviceStatus != "" {
		diagnostic = append(diagnostic, fmt.Sprintf("Service status: %s", strings.TrimSpace(serviceStatus)))
	}

	// Check if socket file exists
	socketCheck, _ := sshClient.executeCommand("ls -la /run/mysqld/mysqld.sock /var/lib/mysql/mysql.sock /tmp/mysql.sock 2>/dev/null")
	if socketCheck != "" {
		diagnostic = append(diagnostic, fmt.Sprintf("Socket files found: %s", strings.TrimSpace(socketCheck)))
	} else {
		diagnostic = append(diagnostic, "No MySQL socket files found")
	}

	// Check if port 3306 is listening (try both netstat and ss)
	portCheck, _ := sshClient.executeCommand("ss -tlnp | grep 3306 2>/dev/null || netstat -tlnp | grep 3306 2>/dev/null")
	if portCheck != "" {
		diagnostic = append(diagnostic, fmt.Sprintf("Port 3306 status: %s", strings.TrimSpace(portCheck)))
	} else {
		diagnostic = append(diagnostic, "Port 3306 is not listening")
	}

	// Check for recent service errors in logs
	errorCheck, _ := sshClient.executeCommand("journalctl -u mysql -u mariadb --no-pager -n 3 --since '1 hour ago' 2>/dev/null | grep -i error | tail -1")
	if errorCheck != "" {
		diagnostic = append(diagnostic, fmt.Sprintf("Recent error: %s", strings.TrimSpace(errorCheck)))
	}

	return strings.Join(diagnostic, "; ")
}

// getSuggestionsForInactiveService provides suggestions for starting MySQL/MariaDB service
func getSuggestionsForInactiveService(sshClient *SSHClient) string {
	var suggestions []string

	// Check which service name to use
	mysqlCheck, _ := sshClient.executeCommand("systemctl list-unit-files | grep -E '^(mysql|mariadb)' | head -1")
	if strings.Contains(mysqlCheck, "mysql") {
		suggestions = append(suggestions, "Try: sudo systemctl start mysql && sudo systemctl enable mysql")
	} else if strings.Contains(mysqlCheck, "mariadb") {
		suggestions = append(suggestions, "Try: sudo systemctl start mariadb && sudo systemctl enable mariadb")
	} else {
		suggestions = append(suggestions, "Try: sudo systemctl start mysql/mariadb")
	}

	// Check for common config issues
	configCheck, _ := sshClient.executeCommand("systemctl status mysql mariadb 2>/dev/null | grep -i 'failed\\|error'")
	if configCheck != "" {
		suggestions = append(suggestions, "Check service logs: sudo journalctl -u mysql -u mariadb --no-pager -n 20")
	}

	// Check if it's a Galera specific issue
	galeraCheck, _ := sshClient.executeCommand("grep -r 'wsrep\\|galera' /etc/mysql/ 2>/dev/null | head -1")
	if galeraCheck != "" {
		suggestions = append(suggestions, "For Galera cluster startup, you may need to bootstrap: sudo galera_new_cluster")
	}

	return strings.Join(suggestions, "; ")
}

// parseClusterSize parses the cluster size from MySQL output
func parseClusterSize(output string, info *GaleraClusterInfo) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wsrep_cluster_size") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if size, err := strconv.Atoi(parts[1]); err == nil {
					info.ClusterSize = size
					return true
				}
			}
		}
	}
	return false
}

// parseClusterStatus parses the cluster status from MySQL output
func parseClusterStatus(output string, info *GaleraClusterInfo) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wsrep_cluster_status") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.ClusterStatus = parts[1]
				return
			}
		}
	}
}

// parseReadyStatus parses the ready status from MySQL output
func parseReadyStatus(output string, info *GaleraClusterInfo) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wsrep_ready") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.IsReady = strings.ToUpper(parts[1]) == "ON"
				return
			}
		}
	}
}

// parseLocalStateComment parses the local state comment from MySQL output
func parseLocalStateComment(output string, info *GaleraClusterInfo) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wsrep_local_state_comment") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.LocalStateComment = parts[1]
				return
			}
		}
	}
}
