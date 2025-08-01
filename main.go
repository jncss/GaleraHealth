package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

// VerbosityLevel represents the logging verbosity level
type VerbosityLevel int

const (
	VerbosityMinimal VerbosityLevel = 0 // Default: Only essential output
	VerbosityNormal  VerbosityLevel = 1 // -v: Normal operations + warnings
	VerbosityVerbose VerbosityLevel = 2 // -vv: Detailed operations + debug info
	VerbosityDebug   VerbosityLevel = 3 // -vvv: Full debug output + raw data
)

var currentVerbosity VerbosityLevel = VerbosityMinimal

// logMinimal prints essential messages (always shown, unless in report mode)
func logMinimal(format string, args ...interface{}) {
	if !reportMode {
		fmt.Printf(format+"\n", args...)
	}
}

// logNormal prints normal operational messages (-v and above, suppressed in report mode)
func logNormal(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityNormal && !reportMode {
		fmt.Printf("📋 "+format+"\n", args...)
	}
}

// logVerbose prints detailed operational messages (-vv and above, suppressed in report mode)
func logVerbose(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityVerbose && !reportMode {
		fmt.Printf("🔍 "+format+"\n", args...)
	}
}

// logDebug prints debug messages (-vvv only, suppressed in report mode)
func logDebug(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityDebug && !reportMode {
		fmt.Printf("🐛 "+format+"\n", args...)
	}
}

// logReport prints messages always, even in report mode (for final summary)
func logReport(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func main() {
	// Parse command line arguments for verbosity and other options
	var args []string
	verbosityCount := 0

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "-v":
			verbosityCount = 1
		case arg == "-vv":
			verbosityCount = 2
		case arg == "-vvv":
			verbosityCount = 3
		case strings.HasPrefix(arg, "-v"):
			// Count consecutive 'v's for -vvv style
			verbosityCount = len(arg) - 1
		case arg == "-y", arg == "--yes":
			useDefaults = true
		case arg == "-s", arg == "--summary":
			reportMode = true
		case arg == "-r", arg == "--recovery":
			runMode = true
		default:
			args = append(args, arg)
		}
	}

	// Set verbosity level
	currentVerbosity = VerbosityLevel(verbosityCount)

	// Validate report mode requirements
	if reportMode && !useDefaults {
		fmt.Println("Error: -s (summary mode) can only be used with -y (automated mode)")
		fmt.Println("Usage: galerahealth -y -s")
		os.Exit(1)
	}

	logDebug("Verbosity level set to: %d", currentVerbosity)

	// Check for other command line arguments
	if len(args) > 0 {
		switch args[0] {
		case "--clear-config", "-c":
			logMinimal("🗑️  Clearing saved configuration...")
			if err := clearConfig(); err != nil {
				log.Fatalf("Error clearing configuration: %v", err)
			}
			logMinimal("✓ Configuration file removed: %s", getConfigPath())
			return
		case "--help", "-h":
			fmt.Println("GaleraHealth - Galera Cluster Monitor")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  galerahealth                      Run the cluster monitor")
			fmt.Println("  galerahealth -y                   Run using saved defaults without prompts")
			fmt.Println("  galerahealth -y -s                Run automated with summary only")
			fmt.Println("  galerahealth -r                   Monitor and attempt cluster recovery if needed")
			fmt.Println("  galerahealth -v                   Run with normal verbosity")
			fmt.Println("  galerahealth -vv                  Run with verbose output")
			fmt.Println("  galerahealth -vvv                 Run with debug output")
			fmt.Println("  galerahealth --clear-config       Clear saved configuration")
			fmt.Println("  galerahealth --help               Show this help")
			fmt.Println()
			fmt.Printf("Configuration file: %s\n", getConfigPath())
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -y, --yes     - Use saved defaults without prompting")
			fmt.Println("  -s, --summary - Show only final summary (requires -y)")
			fmt.Println("  -r, --recovery - Attempt cluster recovery if nodes are down")
			fmt.Println()
			fmt.Println("Verbosity levels:")
			fmt.Println("  (none) - Minimal output (default)")
			fmt.Println("  -v     - Normal operations + warnings")
			fmt.Println("  -vv    - Detailed operations + debug info")
			fmt.Println("  -vvv   - Full debug output + raw data")
			return
		default:
			fmt.Printf("Unknown option: %s\n", args[0])
			fmt.Println("Use --help for available options")
			return
		}
	}

	logMinimal("=== GaleraHealth - Galera Cluster Monitor ===")
	if useDefaults {
		logMinimal("🚀 Running in automatic mode (-y) - using saved defaults")
	}
	logDebug("Application started with verbosity level %d", currentVerbosity)

	// Load saved configuration
	config := loadConfig()
	if config.LastNodeIP != "" {
		logNormal("💾 Loaded saved configuration from %s", getConfigPath())
		logVerbose("   Last used: Node IP: %s, SSH User: %s, MySQL User: %s",
			config.LastNodeIP, config.LastSSHUsername, config.LastMySQLUsername)
	} else if useDefaults {
		logMinimal("⚠️  Using -y flag but no saved configuration found.")
		logMinimal("    The application will use built-in defaults where possible.")
		logMinimal("    Configuration file: %s", getConfigPath())
	}

	// Ask for node IP with default
	nodeIP := promptForInputWithDefault("Enter the Galera cluster node IP", config.LastNodeIP)
	if nodeIP == "" {
		if useDefaults {
			log.Fatal("Node IP is required but no saved configuration found. Run without -y to configure.")
		} else {
			log.Fatal("Node IP is required")
		}
	}

	// Ask for SSH username with default (only if not localhost)
	var username string
	var sshClient *SSHClient
	var connInfo *SSHConnectionInfo
	var err error

	isLocal := isLocalhost(nodeIP)

	if isLocal {
		logMinimal("🏠 Local connection detected - skipping SSH authentication")
		username = "local"
		// Create a dummy connection info for localhost
		connInfo = &SSHConnectionInfo{
			Username:    "local",
			Password:    "",
			HasPassword: false,
			UsedKeys:    false,
		}
	} else {
		defaultUsername := config.LastSSHUsername
		if defaultUsername == "" {
			defaultUsername = "root"
		}
		username = promptForInputWithDefault("Enter SSH username", defaultUsername)
		if username == "" {
			username = "root" // fallback default
		}

		logVerbose("Attempting SSH connection to %s@%s", username, nodeIP)
		// Try SSH connection using per-node credentials
		sshClient, connInfo, err = createSSHConnectionWithNodeCredentials(nodeIP, config)
		if err != nil {
			log.Fatal("Error connecting via SSH:", err)
		}
		defer sshClient.Close()

		// Save the connection info for this node
		if connInfo != nil {
			sshPassword := ""
			if connInfo.HasPassword {
				sshPassword = connInfo.Password
			}
			err = config.setNodeCredentials(nodeIP, connInfo.Username, "", sshPassword, "", connInfo.UsedKeys)
			if err != nil {
				logVerbose("Warning: Could not save node credentials: %v", err)
			} else {
				if connInfo.HasPassword {
					logVerbose("✓ SSH password saved for node %s", nodeIP)
				}
				logVerbose("✓ SSH credentials saved for node %s", nodeIP)
			}
		}

		logMinimal("✓ Successfully connected to node %s", nodeIP)
	}

	logVerbose("Gathering cluster information from initial node")
	// Get cluster information from initial node
	var initialClusterInfo *GaleraClusterInfo

	if isLocal {
		logMinimal("🔍 Analyzing local Galera configuration...")
		initialClusterInfo, err = getGaleraClusterInfoLocal(nodeIP)
	} else {
		initialClusterInfo, err = getGaleraClusterInfo(sshClient, nodeIP)
	}

	if err != nil {
		log.Fatal("Error obtaining cluster information:", err)
	}

	// Close SSH connection if it was used
	if sshClient != nil {
		sshClient.Close()
	}

	// Display initial node information (skip in report mode)
	if !reportMode {
		displayClusterInfo(initialClusterInfo)
	}

	// Ask if user wants to check cluster coherence with intelligent default
	logMinimal("")
	defaultCoherence := config.LastCheckCoherence

	// Smart default detection for cluster coherence
	isMultiNode := len(initialClusterInfo.ClusterAddress) > 0 && strings.Contains(initialClusterInfo.ClusterAddress, ",")

	if isMultiNode {
		// For multi-node clusters, default to checking coherence
		defaultCoherence = true
		if isLocalhost(nodeIP) {
			logVerbose("🔍 Multi-node cluster detected with localhost, defaulting to cluster analysis")
		} else {
			logVerbose("🔍 Multi-node cluster detected, defaulting to cluster analysis")
		}

		// If using -y and we detected a multi-node cluster, override the saved preference
		if useDefaults && !config.LastCheckCoherence {
			logVerbose("🔍 Overriding saved preference due to multi-node cluster detection")
		}
	}

	checkCoherence := promptForBoolWithDefault("Do you want to check cluster configuration coherence across all nodes?", defaultCoherence)

	// Update config with current values
	config.LastNodeIP = nodeIP
	config.LastSSHUsername = username
	config.LastCheckCoherence = checkCoherence

	logDebug("Updated configuration: NodeIP=%s, Username=%s, CheckCoherence=%t", nodeIP, username, checkCoherence)

	if checkCoherence {
		logMinimal("")
		logMinimal("🔍 Performing cluster coherence analysis...")

		analysis, localhostNodeIP, err := performClusterAnalysis(initialClusterInfo, connInfo, config)
		if err != nil {
			log.Fatal("Error performing cluster analysis:", err)
		}

		if !reportMode {
			displayClusterAnalysis(analysis)
		}

		// Ask if user wants to check MySQL/MariaDB status with default
		logMinimal("")
		checkMySQL := promptForBoolWithDefault("Do you want to check MySQL/MariaDB cluster status on all nodes?", config.LastCheckMySQL)
		config.LastCheckMySQL = checkMySQL

		logDebug("CheckMySQL set to: %t", checkMySQL)

		if checkMySQL {
			logVerbose("Gathering MySQL credentials")
			// Get MySQL credentials with default
			mysqlCreds := getMySQLCredentialsWithDefault(config.LastMySQLUsername, config, nodeIP)
			config.LastMySQLUsername = mysqlCreds.Username

			logMinimal("")
			logMinimal("🔍 Checking MySQL/MariaDB status on all nodes...")

			err := checkMySQLStatusOnAllNodes(analysis, connInfo, mysqlCreds, config, localhostNodeIP)
			if err != nil {
				log.Printf("Error checking MySQL status: %v", err)
			}

			// Display results with MySQL status (skip in report mode)
			if !reportMode {
				displayClusterAnalysisWithMySQL(analysis)
			}

		}

		// Display final cluster summary
		displayClusterSummary(analysis)

		// If recovery mode (-r) is enabled, attempt cluster recovery AFTER showing the summary
		if runMode {
			logMinimal("")
			logMinimal("🔧 Recovery mode enabled - checking if cluster recovery is needed...")

			if checkMySQL {
				// Recovery after MySQL check was done
				err := attemptClusterRecoveryAfterMySQLCheck(analysis, config)
				if err != nil {
					logMinimal("❌ Cluster recovery failed: %v", err)
				}
			} else {
				// Recovery without MySQL check - use basic analysis
				err := attemptClusterRecoveryWithAnalysis(analysis, config)
				if err != nil {
					logMinimal("❌ Cluster recovery failed: %v", err)
				}
			}
		}
	} else {
		// Create a basic analysis with just the initial node for summary
		analysis := &ClusterAnalysis{
			InitialNode:  initialClusterInfo,
			AllNodes:     []*GaleraClusterInfo{initialClusterInfo},
			ClusterNodes: []string{initialClusterInfo.NodeIP}, // Only current node analyzed
			IsCoherent:   true,                                // Single node analysis is always coherent for configuration
			ConfigErrors: []string{},
		}

		logDebug("Creating single-node analysis for summary")
		// Display basic summary for single node
		displayClusterSummary(analysis)

		// If recovery mode (-r) is enabled, attempt recovery for single node AFTER showing summary
		if runMode {
			logMinimal("")
			logMinimal("🔧 Recovery mode enabled - checking if single node recovery is needed...")

			err := attemptClusterRecoveryWithAnalysis(analysis, config)
			if err != nil {
				logMinimal("❌ Node recovery failed: %v", err)
			}
		}
	}

	logVerbose("Saving configuration for next time")
	// Save configuration for next time
	if err := saveConfig(config); err != nil {
		logNormal("Warning: Could not save configuration: %v", err)
	} else {
		logNormal("")
		logNormal("💾 Configuration saved for next time")
	}
}

// getMySQLCredentials prompts for MySQL/MariaDB credentials
func getMySQLCredentials() *MySQLConnectionInfo {
	return getMySQLCredentialsWithDefault("", nil, "")
}

// getMySQLCredentialsWithDefault prompts for MySQL/MariaDB credentials with default username
func getMySQLCredentialsWithDefault(defaultUsername string, config *Config, nodeIP string) *MySQLConnectionInfo {
	if !reportMode {
		fmt.Println()
		fmt.Println("Enter MySQL/MariaDB credentials:")
	}

	if defaultUsername == "" {
		defaultUsername = "root"
	}
	username := promptForInputWithDefault("MySQL username", defaultUsername)
	if username == "" {
		username = "root"
	}

	var password string

	// Check if we have a saved encrypted password
	if config != nil && config.HasSavedPassword && nodeIP != "" {
		logVerbose("Found saved encrypted password")
		useStored := promptForBoolWithDefault("Use stored password?", true)
		if useStored {
			logDebug("Attempting to decrypt stored password")
			decryptedPassword, err := decryptPassword(config.EncryptedMySQLPassword, nodeIP)
			if err != nil {
				logNormal("Warning: Could not decrypt stored password: %v", err)
				logMinimal("Please enter password manually:")
			} else {
				password = decryptedPassword
				logMinimal("✓ Using stored password")
				logDebug("Password successfully decrypted")
			}
		}
	}

	// If we don't have a password yet, prompt for it
	if password == "" {
		// If -y flag is used but no saved password, we can't proceed with MySQL check
		if useDefaults {
			logMinimal("⚠️  No saved MySQL password found with -y flag. Skipping MySQL operations.")
			return &MySQLConnectionInfo{Username: username, Password: ""}
		}

		fmt.Print("MySQL password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			log.Printf("Error reading password: %v", err)
			return &MySQLConnectionInfo{Username: username, Password: ""}
		}
		password = string(passwordBytes)

		// Ask if user wants to save the password
		if config != nil && nodeIP != "" && password != "" {
			savePassword := promptForBoolWithDefault("Save this password for next time? (encrypted)", true)
			if savePassword {
				logDebug("Attempting to encrypt password for storage")
				encryptedPassword, err := encryptPassword(password, nodeIP)
				if err != nil {
					logNormal("Warning: Could not encrypt password: %v", err)
				} else {
					config.EncryptedMySQLPassword = encryptedPassword
					config.HasSavedPassword = true
					logMinimal("✓ Password will be saved (encrypted)")
					logDebug("Password successfully encrypted and marked for storage")
				}
			}
		}
	}

	return &MySQLConnectionInfo{
		Username: username,
		Password: password,
	}
}

// attemptClusterRecoveryAfterMySQLCheck attempts recovery after MySQL analysis has been done
func attemptClusterRecoveryAfterMySQLCheck(analysis *ClusterAnalysis, config *Config) error {
	// Since MySQL check has been done, we should have information about node health
	// Check if we need to look at the analysis to determine if recovery is needed

	// TODO: For now, we don't have direct access to MySQL status results
	// So we still need to do a basic recovery check, but we can be smarter about it

	// If cluster configuration is coherent and no errors, likely MySQL check passed
	// and nodes are healthy - do minimal recovery check
	if analysis.IsCoherent && len(analysis.ConfigErrors) == 0 {
		logVerbose("Cluster appears healthy from analysis, doing minimal recovery verification...")

		// Quick check on primary node to avoid unnecessary SSH connections
		primaryNode := analysis.InitialNode.NodeIP
		if primaryNode == "localhost" || primaryNode == "127.0.0.1" {
			cmd := "systemctl is-active mariadb mysqld mysql 2>/dev/null | head -1"
			output, err := executeLocalCommand(cmd)
			if err == nil && strings.TrimSpace(output) == "active" {
				logMinimal("✅ Primary node MySQL/MariaDB is running and cluster appears healthy - no recovery needed")
				return nil
			}
		}
	}

	// If we get here, either cluster has issues or we need to do full recovery analysis
	logVerbose("Proceeding with detailed cluster state analysis for recovery...")
	state, err := analyzeClusterState(analysis.ClusterNodes, config)
	if err != nil {
		return fmt.Errorf("failed to analyze cluster state: %v", err)
	}

	return performClusterRecovery(state, config)
}

// attemptClusterRecoveryWithAnalysis attempts to recover the cluster using existing analysis
func attemptClusterRecoveryWithAnalysis(analysis *ClusterAnalysis, config *Config) error {
	// If cluster analysis shows everything is coherent and we have no config errors,
	// do a minimal check before proceeding with full recovery analysis
	if analysis.IsCoherent && len(analysis.ConfigErrors) == 0 {
		logVerbose("Cluster configuration appears healthy, doing minimal recovery check...")

		// For single-node clusters or localhost, do a quick local check
		if len(analysis.ClusterNodes) == 1 {
			nodeIP := analysis.ClusterNodes[0]
			if nodeIP == "localhost" || nodeIP == "127.0.0.1" {
				cmd := "systemctl is-active mariadb mysqld mysql 2>/dev/null | head -1"
				output, err := executeLocalCommand(cmd)
				if err == nil && strings.TrimSpace(output) == "active" {
					logMinimal("✅ Local MySQL/MariaDB is running - no recovery needed")
					return nil
				}
			}
		}
	}

	// Proceed with detailed recovery analysis
	logVerbose("Proceeding with detailed cluster state analysis for recovery...")
	state, err := analyzeClusterState(analysis.ClusterNodes, config)
	if err != nil {
		return fmt.Errorf("failed to analyze cluster state: %v", err)
	}

	// Attempt recovery based on cluster state
	return performClusterRecovery(state, config)
}

// attemptClusterRecovery attempts to recover the cluster if needed
func attemptClusterRecovery(clusterIPs []string, config *Config) error {
	// First, do a quick check to see if we even need to attempt recovery
	// If we have MySQL status information from previous checks, use that
	quickCheck := true
	for _, ip := range clusterIPs {
		var cmd string
		if ip == "localhost" || ip == "127.0.0.1" {
			cmd = "systemctl is-active mariadb mysqld mysql 2>/dev/null | head -1"
		} else {
			// For remote nodes, we'll need to do a more detailed analysis
			// But first let's see if we can avoid SSH by checking if localhost is part of cluster
			quickCheck = false
			break
		}

		if quickCheck {
			output, err := executeLocalCommand(cmd)
			if err != nil || strings.TrimSpace(output) != "active" {
				quickCheck = false
				break
			}
		}
	}

	// If quick check suggests all is well and we only have localhost, no need for detailed analysis
	if quickCheck && len(clusterIPs) == 1 && (clusterIPs[0] == "localhost" || clusterIPs[0] == "127.0.0.1") {
		logMinimal("✅ Local MySQL/MariaDB is running - no recovery needed")
		return nil
	}

	// Need detailed analysis - analyze current cluster state
	state, err := analyzeClusterState(clusterIPs, config)
	if err != nil {
		return fmt.Errorf("failed to analyze cluster state: %v", err)
	}

	// Attempt recovery based on cluster state
	return performClusterRecovery(state, config)
}
