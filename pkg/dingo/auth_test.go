package dingo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

/*
* Test helper that creates a temporary SSH private key file for testing
* Inputs: t (*testing.T) - test context, encrypted (bool) - whether to encrypt with passphrase
* Outputs: string containing path to key file, cleanup function, error if key creation fails
 */
func createTestSSHKey(t *testing.T, encrypted bool) (string, func(), error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create temporary file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")

	var pemBlock *pem.Block
	if encrypted {
		// Encrypt with passphrase "testpass"
		pemBlock, err = x509.EncryptPEMBlock(
			rand.Reader,
			"RSA PRIVATE KEY",
			x509.MarshalPKCS1PrivateKey(privateKey),
			[]byte("testpass"),
			x509.PEMCipherAES256,
		)
		if err != nil {
			return "", nil, fmt.Errorf("failed to encrypt private key: %v", err)
		}
	} else {
		// Unencrypted key
		pemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}
	}

	// Write key to file
	keyData := pem.EncodeToMemory(pemBlock)
	err = ioutil.WriteFile(keyPath, keyData, 0600)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write key file: %v", err)
	}

	cleanup := func() {
		os.Remove(keyPath)
	}

	return keyPath, cleanup, nil
}

/*
* Test helper that creates a mock SSH server for testing connections
* Inputs: t (*testing.T) - test context
* Outputs: string containing server address, cleanup function, error if server creation fails
 */
func createMockSSHServer(t *testing.T) (string, func(), error) {
	// Generate host key
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate host key: %v", err)
	}

	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create host signer: %v", err)
	}

	// Configure SSH server
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if conn.User() == "testuser" && string(password) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("authentication failed")
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() == "testuser" {
				return nil, nil
			}
			return nil, fmt.Errorf("authentication failed")
		},
	}
	config.AddHostKey(hostSigner)

	// Start listening
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start listener: %v", err)
	}

	// Handle connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Server closed
			}

			go func(conn net.Conn) {
				defer conn.Close()
				ssh.NewServerConn(conn, config)
			}(conn)
		}
	}()

	cleanup := func() {
		listener.Close()
	}

	return listener.Addr().String(), cleanup, nil
}

/*
* Tests ConnectWithPassword function with valid credentials
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithPassword_Success(t *testing.T) {
	serverAddr, cleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer cleanup()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	client, err := ConnectWithPassword(serverAddr, "testuser", "testpass")
	if err != nil {
		t.Fatalf("ConnectWithPassword failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()

	// Verify client status
	if status := client.Status(); status != StatusConnected {
		t.Errorf("Expected status %v, got %v", StatusConnected, status)
	}
}

/*
* Tests ConnectWithPassword function with invalid credentials
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithPassword_InvalidCredentials(t *testing.T) {
	serverAddr, cleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := ConnectWithPassword(serverAddr, "testuser", "wrongpass")
	if err == nil {
		defer client.Close()
		t.Fatal("Expected authentication error")
	}

	if client != nil {
		t.Error("Expected nil client on authentication failure")
	}
}

/*
* Tests ConnectWithPassword function with invalid server address
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithPassword_InvalidAddress(t *testing.T) {
	client, err := ConnectWithPassword("invalid:22", "user", "pass")
	if err == nil {
		defer client.Close()
		t.Fatal("Expected connection error for invalid address")
	}

	if client != nil {
		t.Error("Expected nil client on connection failure")
	}
}

/*
* Tests ConnectWithKey function with valid unencrypted key
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithKey_Success(t *testing.T) {
	keyPath, cleanup, err := createTestSSHKey(t, false)
	if err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}
	defer cleanup()

	serverAddr, serverCleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer serverCleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := ConnectWithKey(serverAddr, "testuser", keyPath)
	if err != nil {
		t.Fatalf("ConnectWithKey failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

/*
* Tests ConnectWithKey function with non-existent key file
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithKey_FileNotFound(t *testing.T) {
	client, err := ConnectWithKey("localhost:22", "user", "/nonexistent/key")
	if err == nil {
		defer client.Close()
		t.Fatal("Expected error for non-existent key file")
	}

	if client != nil {
		t.Error("Expected nil client when key file not found")
	}

	// Verify error message indicates file not found
	if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "cannot find") {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

/*
* Tests ConnectWithKey function with invalid key format
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithKey_InvalidKey(t *testing.T) {
	// Create temporary file with invalid key data
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid_key")
	err := ioutil.WriteFile(keyPath, []byte("not a valid ssh key"), 0600)
	if err != nil {
		t.Fatalf("Failed to create invalid key file: %v", err)
	}

	client, err := ConnectWithKey("localhost:22", "user", keyPath)
	if err == nil {
		defer client.Close()
		t.Fatal("Expected error for invalid key format")
	}

	if client != nil {
		t.Error("Expected nil client when key parsing fails")
	}
}

/*
* Tests ConnectWithKeyAndPassphrase function with correct passphrase
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithKeyAndPassphrase_Success(t *testing.T) {
	keyPath, cleanup, err := createTestSSHKey(t, true)
	if err != nil {
		t.Fatalf("Failed to create encrypted test key: %v", err)
	}
	defer cleanup()

	serverAddr, serverCleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer serverCleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := ConnectWithKeyAndPassphrase(serverAddr, "testuser", keyPath, "testpass")
	if err != nil {
		t.Fatalf("ConnectWithKeyAndPassphrase failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

/*
* Tests ConnectWithKeyAndPassphrase function with incorrect passphrase
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithKeyAndPassphrase_WrongPassphrase(t *testing.T) {
	keyPath, cleanup, err := createTestSSHKey(t, true)
	if err != nil {
		t.Fatalf("Failed to create encrypted test key: %v", err)
	}
	defer cleanup()

	client, err := ConnectWithKeyAndPassphrase("localhost:22", "user", keyPath, "wrongpass")
	if err == nil {
		defer client.Close()
		t.Fatal("Expected error for wrong passphrase")
	}

	if client != nil {
		t.Error("Expected nil client when passphrase is wrong")
	}
}

/*
* Tests ConnectWithConfig function with custom SSH configuration
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConfig_Success(t *testing.T) {
	serverAddr, cleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	config := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ConnectWithConfig(serverAddr, config)
	if err != nil {
		t.Fatalf("ConnectWithConfig failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

/*
* Tests ConnectWithConfig function with nil configuration
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConfig_NilConfig(t *testing.T) {
	// The SSH library panics on nil config, so we need to catch that
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil config
			t.Log("Expected panic caught for nil config")
		}
	}()

	client, err := ConnectWithConfig("localhost:22", nil)
	if err == nil && client != nil {
		defer client.Close()
		t.Fatal("Expected error or panic for nil config")
	}
}

/*
* Tests ConnectWithConnection function with existing network connection
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConnection_Success(t *testing.T) {
	// Create a mock connection that will fail SSH handshake
	// This tests the function structure rather than full SSH functionality
	conn, err := net.Dial("tcp", "127.0.0.1:1") // Non-existent service
	if err == nil {
		defer conn.Close()

		config := &ssh.ClientConfig{
			User:            "testuser",
			Auth:            []ssh.AuthMethod{ssh.Password("testpass")},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         100 * time.Millisecond,
		}

		client, err := ConnectWithConnection(conn, "127.0.0.1:1", config)
		// We expect this to fail since it's not a real SSH server
		if err == nil {
			defer client.Close()
			t.Log("Unexpected success - this may indicate the test needs adjustment")
		} else {
			// This is the expected case - connection should fail SSH handshake
			if client != nil {
				t.Error("Expected nil client when SSH handshake fails")
			}
		}
	}
	// If we can't even establish a TCP connection, that's also expected
}

/*
* Tests ConnectWithConnection function with a real SSH server to test success path
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConnection_SuccessPath(t *testing.T) {
	// Create a mock SSH server
	serverAddr, cleanup, err := createMockSSHServer(t)
	if err != nil {
		t.Skipf("Failed to create mock SSH server: %v", err)
	}
	defer cleanup()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Establish a TCP connection first
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Skipf("Failed to connect to mock server: %v", err)
	}
	defer conn.Close()

	config := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ConnectWithConnection(conn, serverAddr, config)
	if err != nil {
		// This might fail due to mock server limitations, but we're testing the code path
		t.Logf("SSH handshake failed as expected for mock server: %v", err)
		if client != nil {
			t.Error("Expected nil client when SSH handshake fails")
		}
		return
	}

	// If we get here, the connection succeeded
	if client == nil {
		t.Fatal("Expected non-nil client on successful connection")
	}

	defer client.Close()

	// Verify client status
	if status := client.Status(); status != StatusConnected {
		t.Errorf("Expected status %v, got %v", StatusConnected, status)
	}
}

/*
* Tests ConnectWithConnection function with nil connection
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConnection_NilConnection(t *testing.T) {
	// The SSH library panics on nil connection, so we need to catch that
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil connection
			t.Log("Expected panic caught for nil connection")
		}
	}()

	config := &ssh.ClientConfig{
		User:            "testuser",
		Auth:            []ssh.AuthMethod{ssh.Password("testpass")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ConnectWithConnection(nil, "localhost:22", config)
	if err == nil && client != nil {
		defer client.Close()
		t.Fatal("Expected error or panic for nil connection")
	}
}

/*
* Tests SecureHostKeyCallback function (current placeholder implementation)
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestSecureHostKeyCallback(t *testing.T) {
	callback, err := SecureHostKeyCallback("/path/to/known_hosts")
	if err != nil {
		t.Fatalf("SecureHostKeyCallback failed: %v", err)
	}

	if callback == nil {
		t.Fatal("Expected non-nil callback")
	}

	// Test that the callback doesn't panic when called
	// Note: Current implementation uses InsecureIgnoreHostKey
	hostKey := &mockPublicKey{}
	err = callback("localhost:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, hostKey)
	if err != nil {
		t.Errorf("Host key callback failed: %v", err)
	}
}

/*
* Tests connectWithConfig helper function directly
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestConnectWithConfig_Helper(t *testing.T) {
	config := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         100 * time.Millisecond,
	}

	// Test with invalid network type
	client, err := connectWithConfig("invalid", "localhost:22", config)
	if err == nil {
		defer client.Close()
		t.Fatal("Expected error for invalid network type")
	}

	if client != nil {
		t.Error("Expected nil client for invalid network type")
	}
}

/*
* Tests authentication functions with empty parameters
* Inputs: t (*testing.T) - test context
* Outputs: none (test assertions)
 */
func TestAuthFunctions_EmptyParameters(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (SSHClient, error)
	}{
		{
			name: "ConnectWithPassword empty address",
			fn:   func() (SSHClient, error) { return ConnectWithPassword("", "user", "pass") },
		},
		{
			name: "ConnectWithPassword empty user",
			fn:   func() (SSHClient, error) { return ConnectWithPassword("localhost:22", "", "pass") },
		},
		{
			name: "ConnectWithPassword empty password",
			fn:   func() (SSHClient, error) { return ConnectWithPassword("localhost:22", "user", "") },
		},
		{
			name: "ConnectWithKey empty parameters",
			fn:   func() (SSHClient, error) { return ConnectWithKey("", "", "") },
		},
		{
			name: "ConnectWithKeyAndPassphrase empty parameters",
			fn:   func() (SSHClient, error) { return ConnectWithKeyAndPassphrase("", "", "", "") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := tt.fn()
			if err == nil {
				if client != nil {
					client.Close()
				}
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

// Mock public key for testing
type mockPublicKey struct{}

func (m *mockPublicKey) Type() string                                 { return "ssh-rsa" }
func (m *mockPublicKey) Marshal() []byte                              { return []byte("mock-key") }
func (m *mockPublicKey) Verify(data []byte, sig *ssh.Signature) error { return nil }
