package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// External variable for -y option (defined in main.go)
var useDefaults bool

// External variable for -s option (summary mode - only show summary)
var reportMode bool

// NodeCredentials holds SSH and MySQL credentials for a specific node
type NodeCredentials struct {
	NodeIP                 string `json:"node_ip"`
	SSHUsername            string `json:"ssh_username"`
	MySQLUsername          string `json:"mysql_username"`
	EncryptedSSHPassword   string `json:"encrypted_ssh_password,omitempty"`
	EncryptedMySQLPassword string `json:"encrypted_mysql_password,omitempty"`
	HasSSHPassword         bool   `json:"has_ssh_password"`
	HasMySQLPassword       bool   `json:"has_mysql_password"`
	UsesSSHKeys            bool   `json:"uses_ssh_keys"`
}

// Config represents the application configuration
type Config struct {
	LastNodeIP             string            `json:"last_node_ip"`
	LastSSHUsername        string            `json:"last_ssh_username"`
	LastMySQLUsername      string            `json:"last_mysql_username"`
	LastCheckCoherence     bool              `json:"last_check_coherence"`
	LastCheckMySQL         bool              `json:"last_check_mysql"`
	EncryptedMySQLPassword string            `json:"encrypted_mysql_password,omitempty"` // Deprecated, kept for backward compatibility
	HasSavedPassword       bool              `json:"has_saved_password"`                 // Deprecated, kept for backward compatibility
	NodeCredentials        []NodeCredentials `json:"node_credentials"`                   // New: per-node credentials
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".galerahealth")
}

// generateKey generates a key from the node IP for encryption
func generateKey(nodeIP string) []byte {
	hash := sha256.Sum256([]byte("galerahealth:" + nodeIP))
	return hash[:]
}

// encryptPassword encrypts a password using AES-GCM
func encryptPassword(password, nodeIP string) (string, error) {
	if password == "" {
		return "", nil
	}

	key := generateKey(nodeIP)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts a password using AES-GCM
func decryptPassword(encryptedPassword, nodeIP string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	key := generateKey(nodeIP)
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}

// loadConfig loads configuration from ~/.galerahealth
func loadConfig() *Config {
	configPath := getConfigPath()
	if configPath == "" {
		return &Config{} // Return empty config if can't get home dir
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{} // Return empty config if file doesn't exist
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		logNormal("Warning: Could not read config file: %v", err)
		return &Config{}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		logNormal("Warning: Could not parse config file: %v", err)
		return &Config{}
	}

	return &config
}

// saveConfig saves configuration to ~/.galerahealth
func saveConfig(config *Config) error {
	configPath := getConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Convert config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("could not write config file: %v", err)
	}

	return nil
}

// clearConfig removes the configuration file
func clearConfig() error {
	configPath := getConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	// Remove the file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("could not remove config file: %v", err)
	}

	return nil
}

// promptForInputWithDefault prompts for input with a default value
func promptForInputWithDefault(message, defaultValue string) string {
	// If -y flag is set and we have a default value, use it without prompting
	if useDefaults && defaultValue != "" {
		if !reportMode {
			logVerbose("Using default value for '%s': %s", message, defaultValue)
		}
		return defaultValue
	}

	// If -y flag is set but no default value, return empty (caller should handle this)
	if useDefaults && defaultValue == "" {
		if !reportMode {
			logVerbose("No default value available for '%s' in -y mode", message)
		}
		return ""
	}

	if defaultValue != "" {
		fmt.Printf("%s (default: %s): ", message, defaultValue)
	} else {
		fmt.Print(message + ": ")
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// promptForBoolWithDefault prompts for a boolean input with a default value
func promptForBoolWithDefault(message string, defaultValue bool) bool {
	// If -y flag is set, use default value without prompting
	if useDefaults {
		if !reportMode {
			logVerbose("Using default value for '%s': %t", message, defaultValue)
		}
		return defaultValue
	}

	defaultStr := "N"
	promptSuffix := "(y/N)"
	if defaultValue {
		defaultStr = "Y"
		promptSuffix = "(Y/n)"
	}

	response := promptForInputWithDefault(fmt.Sprintf("%s %s", message, promptSuffix), defaultStr)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// getNodeCredentials retrieves credentials for a specific node
func (c *Config) getNodeCredentials(nodeIP string) *NodeCredentials {
	for i := range c.NodeCredentials {
		if c.NodeCredentials[i].NodeIP == nodeIP {
			return &c.NodeCredentials[i]
		}
	}
	return nil
}

// setNodeCredentials saves or updates credentials for a specific node
func (c *Config) setNodeCredentials(nodeIP string, sshUsername, mysqlUsername, sshPassword, mysqlPassword string, usesSSHKeys bool) error {
	// Find existing credentials or create new ones
	var creds *NodeCredentials
	for i := range c.NodeCredentials {
		if c.NodeCredentials[i].NodeIP == nodeIP {
			creds = &c.NodeCredentials[i]
			break
		}
	}

	if creds == nil {
		// Create new credentials entry
		c.NodeCredentials = append(c.NodeCredentials, NodeCredentials{NodeIP: nodeIP})
		creds = &c.NodeCredentials[len(c.NodeCredentials)-1]
	}

	// Update credentials
	creds.SSHUsername = sshUsername
	creds.MySQLUsername = mysqlUsername
	creds.UsesSSHKeys = usesSSHKeys

	// Encrypt and store SSH password if provided
	if sshPassword != "" {
		encryptedSSH, err := encryptPassword(sshPassword, nodeIP)
		if err != nil {
			return fmt.Errorf("failed to encrypt SSH password: %v", err)
		}
		creds.EncryptedSSHPassword = encryptedSSH
		creds.HasSSHPassword = true
	}

	// Encrypt and store MySQL password if provided
	if mysqlPassword != "" {
		encryptedMySQL, err := encryptPassword(mysqlPassword, nodeIP)
		if err != nil {
			return fmt.Errorf("failed to encrypt MySQL password: %v", err)
		}
		creds.EncryptedMySQLPassword = encryptedMySQL
		creds.HasMySQLPassword = true
	}

	return nil
}

// getNodeSSHPassword retrieves and decrypts SSH password for a specific node
func (c *Config) getNodeSSHPassword(nodeIP string) (string, error) {
	creds := c.getNodeCredentials(nodeIP)
	if creds == nil || !creds.HasSSHPassword {
		return "", nil
	}

	return decryptPassword(creds.EncryptedSSHPassword, nodeIP)
}

// getNodeMySQLPassword retrieves and decrypts MySQL password for a specific node
func (c *Config) getNodeMySQLPassword(nodeIP string) (string, error) {
	creds := c.getNodeCredentials(nodeIP)
	if creds == nil || !creds.HasMySQLPassword {
		return "", nil
	}

	return decryptPassword(creds.EncryptedMySQLPassword, nodeIP)
}
