package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// Test program to create a realistic configuration with encrypted password
func main() {
	nodeIP := "10.1.1.91"
	password := "test123"
	
	// Encrypt the password like the real application does
	key := sha256.Sum256([]byte(nodeIP))
	
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	encryptedPassword := base64.StdEncoding.EncodeToString(ciphertext)
	
	// Create configuration
	config := map[string]interface{}{
		"last_node_ip": nodeIP,
		"last_ssh_username": "root",
		"last_mysql_username": "",
		"last_check_coherence": false,
		"last_check_mysql": false,
		"has_saved_password": false,
		"node_credentials": []map[string]interface{}{
			{
				"node_ip": nodeIP,
				"ssh_username": "root",
				"mysql_username": "",
				"encrypted_ssh_password": encryptedPassword,
				"has_ssh_password": true,
				"has_mysql_password": false,
				"uses_ssh_keys": false,
			},
		},
	}
	
	// Write to file
	homeDir, _ := os.UserHomeDir()
	configFile := homeDir + "/.galerahealth"
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(err)
	}
	
	err = os.WriteFile(configFile, data, 0600)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("✓ Created test configuration with encrypted password for %s\n", nodeIP)
	fmt.Printf("✓ Password 'test123' encrypted and saved\n")
	fmt.Printf("✓ Configuration written to %s\n", configFile)
}
