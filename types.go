package main

import "golang.org/x/crypto/ssh"

// GaleraClusterInfo contains information about a Galera cluster node
type GaleraClusterInfo struct {
	ClusterName    string
	ClusterAddress string
	NodeName       string
	NodeAddress    string
	NodeIP         string
	// MySQL/MariaDB status information
	ClusterSize       int
	ClusterStatus     string
	IsReady           bool
	LocalStateComment string
	MySQLResponding   bool
	StatusError       string
}

// ClusterAnalysis contains the results of analyzing cluster coherence
type ClusterAnalysis struct {
	InitialNode  *GaleraClusterInfo
	AllNodes     []*GaleraClusterInfo
	ClusterNodes []string
	ConfigErrors []string
	IsCoherent   bool
}

// SSHConnectionInfo holds information about SSH connection credentials and methods
type SSHConnectionInfo struct {
	Username    string
	Password    string
	HasPassword bool
	UsedKeys    bool
}

// MySQLConnectionInfo holds MySQL/MariaDB connection credentials
type MySQLConnectionInfo struct {
	Username string
	Password string
}

// SSHClient wraps the SSH client connection
type SSHClient struct {
	client *ssh.Client
}
