package main

import (
	"os/exec"
	"strings"
)

// isLocalhost checks if the given IP address refers to the local machine
func isLocalhost(ip string) bool {
	localhostPatterns := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
	}

	ip = strings.TrimSpace(strings.ToLower(ip))
	for _, pattern := range localhostPatterns {
		if ip == pattern {
			return true
		}
	}
	return false
}

// executeLocalCommand executes a command locally without SSH
func executeLocalCommand(command string) (string, error) {
	logDebug("Executing local command: %s", command)
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.Output()
	return string(output), err
}
