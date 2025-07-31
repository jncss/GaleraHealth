package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Close closes the SSH connection
func (s *SSHClient) Close() error {
	return s.client.Close()
}

// executeCommand executes a command on the remote server via SSH
func (s *SSHClient) executeCommand(command string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// createSSHConnectionWithNodeCredentials creates SSH connection using node-specific credentials
func createSSHConnectionWithNodeCredentials(host string, config *Config) (*SSHClient, *SSHConnectionInfo, error) {
	creds := config.getNodeCredentials(host)

	if creds != nil {
		// We have saved credentials for this node
		logVerbose("      🔍 Found saved credentials for node %s", host)

		if creds.UsesSSHKeys {
			// Try SSH keys first
			logVerbose("      🔑 Trying SSH keys for %s...", host)
			client, err := createSSHConnectionWithKeys(host, creds.SSHUsername)
			if err == nil {
				logVerbose("      ✓ Connected to %s using SSH keys", host)
				return client, &SSHConnectionInfo{
					Username:    creds.SSHUsername,
					Password:    "",
					HasPassword: false,
					UsedKeys:    true,
				}, nil
			}
			logVerbose("      ⚠️  SSH keys failed for %s: %v", host, err)
		}

		if creds.HasSSHPassword {
			// Try saved password
			logVerbose("      🔐 Trying saved password for %s...", host)
			password, err := config.getNodeSSHPassword(host)
			if err != nil {
				logVerbose("      ❌ Failed to decrypt password for %s: %v", host, err)
			} else {
				client, err := createSSHConnectionWithPassword(host, creds.SSHUsername, password)
				if err == nil {
					logVerbose("      ✓ Connected to %s using saved password", host)
					return client, &SSHConnectionInfo{
						Username:    creds.SSHUsername,
						Password:    password,
						HasPassword: true,
						UsedKeys:    false,
					}, nil
				}
				logVerbose("      ⚠️  Saved password failed for %s: %v", host, err)
			}
		}

		// Use saved username but need new password
		logNormal("Saved credentials for %s failed, requesting new password...", host)
		return createSSHConnectionWithFallbackAndUsername(host, creds.SSHUsername)
	}

	// No saved credentials, use fallback
	logVerbose("      🆕 No saved credentials for %s, using fallback authentication", host)
	return createSSHConnectionWithFallbackAndUsername(host, "root")
}

// createSSHConnectionWithFallbackAndUsername creates SSH connection with specific username
func createSSHConnectionWithFallbackAndUsername(host, username string) (*SSHClient, *SSHConnectionInfo, error) {
	connInfo := &SSHConnectionInfo{
		Username:    username,
		HasPassword: false,
		UsedKeys:    false,
	}

	// First attempt: connection without password (SSH keys)
	logVerbose("🔑 Attempting SSH connection without password to node %s as %s", host, username)

	sshClient, err := createSSHConnectionWithKeys(host, username)
	if err == nil {
		logNormal("✓ SSH connection successful using keys!")
		connInfo.UsedKeys = true
		return sshClient, connInfo, nil
	}

	logNormal("⚠️  Connection with keys failed: %v", err)
	logNormal("🔐 Attempting connection with password...")

	// Second attempt: ask for password
	fmt.Printf("Enter SSH password for %s@%s: ", username, host)
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading password: %v", err)
	}
	fmt.Println() // new line after password

	connInfo.Password = string(password)
	connInfo.HasPassword = true

	client, err := createSSHConnectionWithPassword(host, username, string(password))
	return client, connInfo, err
}

// createSSHConnectionWithKeys creates SSH connection using SSH keys
func createSSHConnectionWithKeys(host, username string) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				// Try to load SSH keys from standard locations
				keyPaths := []string{
					os.Getenv("HOME") + "/.ssh/id_rsa",
					os.Getenv("HOME") + "/.ssh/id_ecdsa",
					os.Getenv("HOME") + "/.ssh/id_ed25519",
				}

				var signers []ssh.Signer
				for _, keyPath := range keyPaths {
					if key, err := loadPrivateKey(keyPath); err == nil {
						signers = append(signers, key)
					}
				}

				if len(signers) == 0 {
					return nil, fmt.Errorf("no valid SSH keys found")
				}

				return signers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second, // Shorter timeout for key attempt
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host += ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("error establishing SSH connection with keys: %v", err)
	}

	return &SSHClient{client: client}, nil
}

// loadPrivateKey loads a private key from file
func loadPrivateKey(keyPath string) (ssh.Signer, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

// createSSHConnectionWithPassword creates SSH connection using password
func createSSHConnectionWithPassword(host, username, password string) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: In production, verify host keys properly
		Timeout:         10 * time.Second,
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host += ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("error establishing SSH connection: %v", err)
	}

	return &SSHClient{client: client}, nil
}
