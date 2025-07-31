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

// logMinimal prints essential messages (always shown)
func logMinimal(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// logNormal prints normal operational messages (-v and above)
func logNormal(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityNormal {
		fmt.Printf("📋 "+format+"\n", args...)
	}
}

// logVerbose prints detailed operational messages (-vv and above)
func logVerbose(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityVerbose {
		fmt.Printf("🔍 "+format+"\n", args...)
	}
}

// logDebug prints debug messages (-vvv only)
func logDebug(format string, args ...interface{}) {
	if currentVerbosity >= VerbosityDebug {
		fmt.Printf("🐛 "+format+"\n", args...)
	}
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
		default:
			args = append(args, arg)
		}
	}

	// Set verbosity level
	currentVerbosity = VerbosityLevel(verbosityCount)

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
			fmt.Println("  galerahealth -v                   Run with normal verbosity")
			fmt.Println("  galerahealth -vv                  Run with verbose output")
			fmt.Println("  galerahealth -vvv                 Run with debug output")
			fmt.Println("  galerahealth --clear-config       Clear saved configuration")
			fmt.Println("  galerahealth --help               Show this help")
			fmt.Println()
			fmt.Printf("Configuration file: %s\n", getConfigPath())
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
	logDebug("Application started with verbosity level %d", currentVerbosity)

	// Load saved configuration
	config := loadConfig()
	if config.LastNodeIP != "" {
		logNormal("💾 Loaded saved configuration from %s", getConfigPath())
		logVerbose("   Last used: Node IP: %s, SSH User: %s, MySQL User: %s",
			config.LastNodeIP, config.LastSSHUsername, config.LastMySQLUsername)
	}

	// Ask for node IP with default
	nodeIP := promptForInputWithDefault("Enter the Galera cluster node IP", config.LastNodeIP)
	if nodeIP == "" {
		log.Fatal("Node IP is required")
	}

	// Ask for SSH username with default
	defaultUsername := config.LastSSHUsername
	if defaultUsername == "" {
		defaultUsername = "root"
	}
	username := promptForInputWithDefault("Enter SSH username", defaultUsername)
	if username == "" {
		username = "root" // fallback default
	}

	logVerbose("Attempting SSH connection to %s@%s", username, nodeIP)
	// Try SSH connection (first without password, then with password if needed)
	sshClient, connInfo, err := createSSHConnectionWithFallbackAndInfo(nodeIP, username)
	if err != nil {
		log.Fatal("Error connecting via SSH:", err)
	}
	defer sshClient.Close()

	logMinimal("✓ Successfully connected to node %s", nodeIP)

	logVerbose("Gathering cluster information from initial node")
	// Get cluster information from initial node
	initialClusterInfo, err := getGaleraClusterInfo(sshClient, nodeIP)
	if err != nil {
		log.Fatal("Error obtaining cluster information:", err)
	}

	// Close initial connection
	sshClient.Close()

	// Display initial node information
	displayClusterInfo(initialClusterInfo)

	// Ask if user wants to check cluster coherence with default
	logMinimal("")
	checkCoherence := promptForBoolWithDefault("Do you want to check cluster configuration coherence across all nodes?", config.LastCheckCoherence)

	// Update config with current values
	config.LastNodeIP = nodeIP
	config.LastSSHUsername = username
	config.LastCheckCoherence = checkCoherence

	logDebug("Updated configuration: NodeIP=%s, Username=%s, CheckCoherence=%t", nodeIP, username, checkCoherence)

	if checkCoherence {
		logMinimal("")
		logMinimal("🔍 Performing cluster coherence analysis...")

		analysis, err := performClusterAnalysis(initialClusterInfo, connInfo)
		if err != nil {
			log.Fatal("Error performing cluster analysis:", err)
		}

		displayClusterAnalysis(analysis)

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

			err := checkMySQLStatusOnAllNodes(analysis, connInfo, mysqlCreds)
			if err != nil {
				log.Printf("Error checking MySQL status: %v", err)
			}

			// Display results with MySQL status
			displayClusterAnalysisWithMySQL(analysis)
		}

		// Display final cluster summary
		displayClusterSummary(analysis)
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
	fmt.Println()
	fmt.Println("Enter MySQL/MariaDB credentials:")

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
