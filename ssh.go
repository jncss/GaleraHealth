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

// createSSHConnectionWithFallback creates SSH connection with fallback from keys to password
func createSSHConnectionWithFallback(host, username string) (*SSHClient, error) {
	client, _, err := createSSHConnectionWithFallbackAndInfo(host, username)
	return client, err
}

// createSSHConnectionWithFallbackAndInfo creates SSH connection and returns connection info
func createSSHConnectionWithFallbackAndInfo(host, username string) (*SSHClient, *SSHConnectionInfo, error) {
	connInfo := &SSHConnectionInfo{
		Username:    username,
		HasPassword: false,
		UsedKeys:    false,
	}

	// First attempt: connection without password (SSH keys)
	logVerbose("🔑 Attempting SSH connection without password to node %s", host)

	sshClient, err := createSSHConnectionWithKeys(host, username)
	if err == nil {
		logNormal("✓ SSH connection successful using keys!")
		connInfo.UsedKeys = true
		return sshClient, connInfo, nil
	}

	logNormal("⚠️  Connection with keys failed: %v", err)
	logNormal("🔐 Attempting connection with password...")

	// Second attempt: ask for password
	fmt.Print("Enter SSH password: ")
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

// createSSHConnectionWithInfo creates SSH connection using saved connection info
func createSSHConnectionWithInfo(host string, connInfo *SSHConnectionInfo) (*SSHClient, error) {
	if connInfo.UsedKeys {
		// Try keys first
		logVerbose("      🔑 Trying SSH keys...")
		client, err := createSSHConnectionWithKeys(host, connInfo.Username)
		if err == nil {
			logVerbose("      ✓ Connected using SSH keys")
			return client, nil
		}

		// If keys fail and we have password, try password
		if connInfo.HasPassword {
			logVerbose("      🔐 Keys failed, trying saved password...")
			client, err := createSSHConnectionWithPassword(host, connInfo.Username, connInfo.Password)
			if err == nil {
				logVerbose("      ✓ Connected using saved password")
			}
			return client, err
		}

		return nil, fmt.Errorf("SSH keys failed and no password available")
	}

	if connInfo.HasPassword {
		logVerbose("      🔐 Using saved password...")
		client, err := createSSHConnectionWithPassword(host, connInfo.Username, connInfo.Password)
		if err == nil {
			logVerbose("      ✓ Connected using saved password")
		}
		return client, err
	}

	return nil, fmt.Errorf("no authentication method available")
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
