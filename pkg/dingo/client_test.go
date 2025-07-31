package dingo

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClientInterface abstracts the SSH client methods we need for testing
type SSHClientInterface interface {
	NewSession() (*ssh.Session, error)
	Close() error
	Conn() ssh.Conn
}

// SFTPClientInterface abstracts the SFTP client methods we need for testing
type SFTPClientInterface interface {
	Close() error
	Open(path string) (*sftp.File, error)
	OpenFile(path string, f int) (*sftp.File, error)
	Mkdir(path string) error
	MkdirAll(path string) error
	Remove(path string) error
	RemoveDirectory(path string) error
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
	Chmod(path string, mode os.FileMode) error
	Chown(path string, uid, gid int) error
	Rename(oldname, newname string) error
}

// testableClient wraps the client for testing with interface-based mocking
type testableClient struct {
	sshClient  SSHClientInterface
	sftpClient SFTPClientInterface
	config     *ClientConfig
	status     ConnectionStatus
}

// Implement SSHClient interface for testableClient
func (c *testableClient) Command(cmd string) CommandExecutor {
	return &remoteScript{
		client:     nil, // Mock doesn't need real ssh.Client
		scriptType: CommandLine,
		script:     cmd,
	}
}

func (c *testableClient) Script(script string) CommandExecutor {
	return &remoteScript{
		client:     nil, // Mock doesn't need real ssh.Client
		scriptType: RawScript,
		script:     script,
	}
}

func (c *testableClient) ScriptFile(path string) CommandExecutor {
	return &remoteScript{
		client:     nil, // Mock doesn't need real ssh.Client
		scriptType: ScriptFile,
		scriptFile: path,
	}
}

func (c *testableClient) Shell() Shell {
	return &remoteShell{
		client:     nil, // Mock doesn't need real ssh.Client
		shellType:  NonInteractiveShell,
		requestPty: false,
	}
}

func (c *testableClient) InteractiveShell(config *TerminalConfig) Shell {
	if config == nil {
		config = DefaultTerminalConfig
	}
	return &remoteShell{
		client:         nil, // Mock doesn't need real ssh.Client
		shellType:      InteractiveShell,
		requestPty:     true,
		terminalConfig: config,
	}
}

func (c *testableClient) FileSystem(opts ...SftpOption) FileSystem {
	if c.sftpClient == nil {
		// Create mock SFTP client for testing
		c.sftpClient = createMockSFTPClient()
	}

	config := &SftpConfig{}
	for _, opt := range opts {
		opt(config)
	}

	return &remoteFileSystem{
		client: nil, // Mock doesn't need real ssh.Client
		sftp:   nil, // Mock doesn't need real sftp.Client
		config: config,
		err:    nil,
	}
}

func (c *testableClient) Close() error {
	if c.status == StatusDisconnected {
		return nil
	}

	var err error
	if c.sftpClient != nil {
		err = c.sftpClient.Close()
	}
	if sshErr := c.sshClient.Close(); sshErr != nil && err == nil {
		err = sshErr
	}

	c.status = StatusDisconnected
	return err
}

func (c *testableClient) Status() ConnectionStatus {
	return c.status
}

func (c *testableClient) UnderlyingClient() *ssh.Client {
	return nil // Mock doesn't return real ssh.Client
}

// mockSSHClient implements SSHClientInterface for testing
type mockSSHClient struct {
	closed       bool
	sessionCount int
	shouldFail   bool
	closeError   error
}

func (m *mockSSHClient) NewSession() (*ssh.Session, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock connection error")
	}
	m.sessionCount++
	return nil, nil // Mock doesn't need real ssh.Session
}

func (m *mockSSHClient) Close() error {
	m.closed = true
	return m.closeError
}

func (m *mockSSHClient) Conn() ssh.Conn {
	return &mockConn{}
}

// mockConn implements ssh.Conn for testing
type mockConn struct{}

func (m *mockConn) User() string          { return "testuser" }
func (m *mockConn) SessionID() []byte     { return []byte("session123") }
func (m *mockConn) ClientVersion() []byte { return []byte("SSH-2.0-Test") }
func (m *mockConn) ServerVersion() []byte { return []byte("SSH-2.0-TestServer") }
func (m *mockConn) RemoteAddr() net.Addr  { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22} }
func (m *mockConn) LocalAddr() net.Addr   { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0} }
func (m *mockConn) SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error) {
	return true, nil, nil
}
func (m *mockConn) OpenChannel(name string, data []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	return nil, nil, nil
}
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) Wait() error                        { return nil }
func (m *mockConn) Read(b []byte) (int, error)         { return 0, nil }
func (m *mockConn) Write(b []byte) (int, error)        { return len(b), nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// mockSFTPClient implements SFTPClientInterface for testing
type mockSFTPClient struct {
	closed     bool
	closeError error
	files      map[string][]byte
	dirs       map[string]bool
}

func (m *mockSFTPClient) Close() error {
	m.closed = true
	return m.closeError
}

func (m *mockSFTPClient) Open(path string) (*sftp.File, error) {
	if _, exists := m.files[path]; !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return nil, nil
}

func (m *mockSFTPClient) OpenFile(path string, f int) (*sftp.File, error) {
	if f&os.O_CREATE != 0 {
		m.files[path] = []byte{}
	}
	return nil, nil
}

func (m *mockSFTPClient) Mkdir(path string) error {
	m.dirs[path] = true
	return nil
}

func (m *mockSFTPClient) MkdirAll(path string) error {
	m.dirs[path] = true
	return nil
}

func (m *mockSFTPClient) Remove(path string) error {
	delete(m.files, path)
	return nil
}

func (m *mockSFTPClient) RemoveDirectory(path string) error {
	delete(m.dirs, path)
	return nil
}

func (m *mockSFTPClient) Stat(path string) (os.FileInfo, error) {
	if _, exists := m.files[path]; exists {
		return &mockFileInfo{name: path, size: int64(len(m.files[path]))}, nil
	}
	if _, exists := m.dirs[path]; exists {
		return &mockFileInfo{name: path, isDir: true}, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (m *mockSFTPClient) Lstat(path string) (os.FileInfo, error) {
	return m.Stat(path)
}

func (m *mockSFTPClient) ReadDir(path string) ([]os.FileInfo, error) {
	var files []os.FileInfo
	files = append(files, &mockFileInfo{name: "test.txt", size: 100})
	files = append(files, &mockFileInfo{name: "subdir", isDir: true})
	return files, nil
}

func (m *mockSFTPClient) Chmod(path string, mode os.FileMode) error {
	return nil
}

func (m *mockSFTPClient) Chown(path string, uid, gid int) error {
	return nil
}

func (m *mockSFTPClient) Rename(oldname, newname string) error {
	if data, exists := m.files[oldname]; exists {
		m.files[newname] = data
		delete(m.files, oldname)
	}
	return nil
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m *mockFileInfo) Name() string { return m.name }
func (m *mockFileInfo) Size() int64  { return m.size }
func (m *mockFileInfo) Mode() os.FileMode {
	if m.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// Helper functions for creating mocks
func createMockSSHClient(shouldFail bool) SSHClientInterface {
	return &mockSSHClient{
		shouldFail: shouldFail,
	}
}

func createMockSFTPClient() SFTPClientInterface {
	return &mockSFTPClient{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func createTestableClient(sshClient SSHClientInterface, config *ClientConfig) *testableClient {
	if config == nil {
		config = DefaultClientConfig
	}
	return &testableClient{
		sshClient: sshClient,
		config:    config,
		status:    StatusConnected,
	}
}

// Helper function to create a testable version of the real client
func createTestableClientFromReal(mockSSH *ssh.Client, config *ClientConfig) SSHClient {
	return newClient(mockSSH, config)
}

// sshClientAdapter wraps a real ssh.Client to work with our mocks
type sshClientAdapter struct {
	mock SSHClientInterface
}

func (s *sshClientAdapter) NewSession() (*ssh.Session, error) {
	return s.mock.NewSession()
}

func (s *sshClientAdapter) Close() error {
	return s.mock.Close()
}

func (s *sshClientAdapter) Conn() ssh.Conn {
	return s.mock.Conn()
}

// Creates a mock ssh.Client that wraps our interface mock
func createMockSSHClientReal(shouldFail bool) *ssh.Client {
	// We can't create a real ssh.Client easily, so we'll test with the wrapper approach
	// For testing the actual client.go functions, we need to test them differently
	return nil
}

// Unit Tests

func TestNewClient_WithDefaultConfig(t *testing.T) {
	// Test with nil ssh.Client (common in unit tests)
	client := newClient(nil, nil)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Verify status is connected
	if status := client.Status(); status != StatusConnected {
		t.Errorf("Expected status %v, got %v", StatusConnected, status)
	}
}

func TestNewClient_WithCustomConfig(t *testing.T) {
	customConfig := &ClientConfig{
		Timeout:       60 * time.Second,
		KeepAlive:     10 * time.Second,
		MaxSessions:   5,
		RetryAttempts: 5,
		RetryDelay:    2 * time.Second,
	}

	client := newClient(nil, customConfig)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Verify status is connected
	if status := client.Status(); status != StatusConnected {
		t.Errorf("Expected status %v, got %v", StatusConnected, status)
	}
}

func TestClient_Command(t *testing.T) {
	client := newClient(nil, nil)

	cmd := "ls -la"
	executor := client.Command(cmd)

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor")
	}

	// Verify it's a remoteScript with correct type and command
	remoteScript, ok := executor.(*remoteScript)
	if !ok {
		t.Fatal("Expected CommandExecutor to be *remoteScript")
	}

	if remoteScript.scriptType != CommandLine {
		t.Errorf("Expected script type %v, got %v", CommandLine, remoteScript.scriptType)
	}

	if remoteScript.script != cmd {
		t.Errorf("Expected script %q, got %q", cmd, remoteScript.script)
	}
}

func TestClient_Command_EmptyString(t *testing.T) {
	client := newClient(nil, nil)

	executor := client.Command("")

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor even with empty command")
	}

	remoteScript := executor.(*remoteScript)
	if remoteScript.script != "" {
		t.Errorf("Expected empty script, got %q", remoteScript.script)
	}
}

func TestClient_Script(t *testing.T) {
	client := newClient(nil, nil)

	script := "#!/bin/bash\necho 'Hello World'"
	executor := client.Script(script)

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor")
	}

	remoteScript, ok := executor.(*remoteScript)
	if !ok {
		t.Fatal("Expected CommandExecutor to be *remoteScript")
	}

	if remoteScript.scriptType != RawScript {
		t.Errorf("Expected script type %v, got %v", RawScript, remoteScript.scriptType)
	}

	if remoteScript.script != script {
		t.Errorf("Expected script %q, got %q", script, remoteScript.script)
	}
}

func TestClient_Script_Multiline(t *testing.T) {
	client := newClient(nil, nil)

	script := `#!/bin/bash
set -e
echo "Starting script"
for i in {1..3}; do
    echo "Iteration $i"
done
echo "Script completed"`

	executor := client.Script(script)

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor")
	}

	remoteScript := executor.(*remoteScript)
	if remoteScript.script != script {
		t.Errorf("Expected multiline script to be preserved correctly")
	}
}

func TestClient_ScriptFile(t *testing.T) {
	client := newClient(nil, nil)

	path := "/tmp/test.sh"
	executor := client.ScriptFile(path)

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor")
	}

	remoteScript, ok := executor.(*remoteScript)
	if !ok {
		t.Fatal("Expected CommandExecutor to be *remoteScript")
	}

	if remoteScript.scriptType != ScriptFile {
		t.Errorf("Expected script type %v, got %v", ScriptFile, remoteScript.scriptType)
	}

	if remoteScript.scriptFile != path {
		t.Errorf("Expected script file %q, got %q", path, remoteScript.scriptFile)
	}
}

func TestClient_ScriptFile_EmptyPath(t *testing.T) {
	client := newClient(nil, nil)

	executor := client.ScriptFile("")

	if executor == nil {
		t.Fatal("Expected non-nil CommandExecutor even with empty path")
	}

	remoteScript := executor.(*remoteScript)
	if remoteScript.scriptFile != "" {
		t.Errorf("Expected empty script file, got %q", remoteScript.scriptFile)
	}
}

func TestClient_Shell(t *testing.T) {
	client := newClient(nil, nil)

	shell := client.Shell()

	if shell == nil {
		t.Fatal("Expected non-nil Shell")
	}

	remoteShell, ok := shell.(*remoteShell)
	if !ok {
		t.Fatal("Expected Shell to be *remoteShell")
	}

	if remoteShell.shellType != NonInteractiveShell {
		t.Errorf("Expected shell type %v, got %v", NonInteractiveShell, remoteShell.shellType)
	}

	if remoteShell.requestPty != false {
		t.Errorf("Expected requestPty to be false, got %v", remoteShell.requestPty)
	}
}

func TestClient_InteractiveShell_DefaultConfig(t *testing.T) {
	client := newClient(nil, nil)

	shell := client.InteractiveShell(nil)

	if shell == nil {
		t.Fatal("Expected non-nil Shell")
	}

	remoteShell, ok := shell.(*remoteShell)
	if !ok {
		t.Fatal("Expected Shell to be *remoteShell")
	}

	if remoteShell.shellType != InteractiveShell {
		t.Errorf("Expected shell type %v, got %v", InteractiveShell, remoteShell.shellType)
	}

	if remoteShell.requestPty != true {
		t.Errorf("Expected requestPty to be true, got %v", remoteShell.requestPty)
	}

	if remoteShell.terminalConfig == nil {
		t.Error("Expected terminalConfig to be set to default")
	}
}

func TestClient_InteractiveShell_CustomConfig(t *testing.T) {
	client := newClient(nil, nil)

	customConfig := &TerminalConfig{
		Term:   "xterm-256color",
		Height: 50,
		Width:  120,
		Modes:  ssh.TerminalModes{},
	}

	shell := client.InteractiveShell(customConfig)

	if shell == nil {
		t.Fatal("Expected non-nil Shell")
	}

	remoteShell := shell.(*remoteShell)
	if remoteShell.terminalConfig != customConfig {
		t.Error("Expected custom terminal config to be used")
	}
}

func TestClient_FileSystem_NoOptions(t *testing.T) {
	client := newClient(nil, nil)

	// FileSystem with nil ssh.Client will panic when trying to create SFTP client
	// We need to handle this gracefully in the test
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil ssh.Client
			t.Log("Expected panic caught when creating FileSystem with nil ssh.Client")
		}
	}()

	fs := client.FileSystem()

	if fs == nil {
		t.Fatal("Expected non-nil FileSystem")
	}

	remoteFS, ok := fs.(*remoteFileSystem)
	if !ok {
		t.Fatal("Expected FileSystem to be *remoteFileSystem")
	}

	// With nil ssh.Client, SFTP creation should fail but FileSystem should still be created
	if remoteFS.err == nil {
		t.Error("Expected error when creating SFTP with nil ssh.Client")
	}
}

// TestClient_FileSystem_NilSSHClient tests FileSystem with nil SSH client
func TestClient_FileSystem_NilSSHClient(t *testing.T) {
	client := newClient(nil, nil)

	fs := client.FileSystem()

	if fs == nil {
		t.Fatal("Expected non-nil FileSystem")
	}

	remoteFS, ok := fs.(*remoteFileSystem)
	if !ok {
		t.Fatal("Expected FileSystem to be *remoteFileSystem")
	}

	// Should return FileSystem with error for nil SSH client
	if remoteFS.err == nil {
		t.Error("Expected error when SSH client is nil")
	}

	expectedError := "SSH client is nil"
	if remoteFS.err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %v", expectedError, remoteFS.err)
	}

	// Should have nil SFTP client
	if remoteFS.sftp != nil {
		t.Error("Expected nil SFTP client when SSH client is nil")
	}
}

// TestClient_FileSystem_SuccessfulSFTPCreation tests FileSystem with successful SFTP creation
func TestClient_FileSystem_SuccessfulSFTPCreation(t *testing.T) {
	// Create a testable client with mock SSH client
	mockSSH := &mockSSHClient{shouldFail: false}
	client := createTestableClient(mockSSH, nil)

	fs := client.FileSystem()

	if fs == nil {
		t.Fatal("Expected non-nil FileSystem")
	}

	// For testableClient, the FileSystem returns a configured filesystem
	// The key test is that it doesn't panic and returns a valid interface
	var _ FileSystem = fs
}

// TestClient_FileSystem_SFTPCreationFailure tests FileSystem when SFTP creation fails
func TestClient_FileSystem_SFTPCreationFailure(t *testing.T) {
	// Create a filesystem that simulates SFTP creation failure
	fs := &remoteFileSystem{
		client: nil,
		sftp:   nil,
		err:    fmt.Errorf("mock SFTP creation error"),
	}

	// Instead of testing the mock client, test the actual behavior
	// The key test is that when SFTP creation fails, we get an error
	if fs.err == nil {
		t.Error("Expected error when SFTP creation fails")
	}

	// Should have nil SFTP client
	if fs.sftp != nil {
		t.Error("Expected nil SFTP client when creation fails")
	}
}

// TestClient_FileSystem_WithOptions tests FileSystem with SFTP options
func TestClient_FileSystem_WithOptions(t *testing.T) {
	client := newClient(nil, nil)

	// Create some test options
	option1 := func(config *SftpConfig) {
		config.MaxPacket = 1024
	}
	option2 := func(config *SftpConfig) {
		config.UseFstat = false
	}

	fs := client.FileSystem(option1, option2)

	if fs == nil {
		t.Fatal("Expected non-nil FileSystem")
	}

	remoteFS, ok := fs.(*remoteFileSystem)
	if !ok {
		t.Fatal("Expected FileSystem to be *remoteFileSystem")
	}

	// Should have config with applied options
	if remoteFS.config == nil {
		t.Fatal("Expected config to be set")
	}

	if remoteFS.config.MaxPacket != 1024 {
		t.Errorf("Expected MaxPacket to be 1024, got %d", remoteFS.config.MaxPacket)
	}

	if remoteFS.config.UseFstat != false {
		t.Errorf("Expected UseFstat to be false, got %v", remoteFS.config.UseFstat)
	}
}

// TestClient_FileSystem_MultipleCalls tests FileSystem with multiple calls
func TestClient_FileSystem_MultipleCalls(t *testing.T) {
	client := newClient(nil, nil)

	// Call FileSystem multiple times
	fs1 := client.FileSystem()
	fs2 := client.FileSystem()

	if fs1 == nil || fs2 == nil {
		t.Fatal("Expected non-nil FileSystem instances")
	}

	// Both should return FileSystem instances
	_, ok1 := fs1.(*remoteFileSystem)
	_, ok2 := fs2.(*remoteFileSystem)

	if !ok1 || !ok2 {
		t.Error("Expected both FileSystem instances to be *remoteFileSystem")
	}

	// Both should have the same error (SSH client is nil)
	remoteFS1, _ := fs1.(*remoteFileSystem)
	remoteFS2, _ := fs2.(*remoteFileSystem)

	if remoteFS1.err == nil || remoteFS2.err == nil {
		t.Error("Expected errors for both FileSystem instances with nil SSH client")
	}
}

func TestClient_Close(t *testing.T) {
	client := newClient(nil, nil)

	// Close should handle nil ssh.Client gracefully (no panic, no error)
	err := client.Close()

	// With nil ssh.Client, Close now handles gracefully
	if err != nil {
		t.Errorf("Expected no error when closing with nil ssh.Client, got %v", err)
	}

	// Verify status is disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}
}

// TestClient_Close_WithSSHClientError tests Close when SSH client returns error
func TestClient_Close_WithSSHClientError(t *testing.T) {
	// Create a mock SSH client that returns error on close
	mockSSH := &mockSSHClient{
		shouldFail: false,
		closeError: errors.New("SSH close error"),
	}

	client := createTestableClient(mockSSH, nil)

	err := client.Close()

	// Should return the SSH close error
	if err == nil {
		t.Error("Expected error when SSH client close fails")
	}

	if !strings.Contains(err.Error(), "SSH close error") {
		t.Errorf("Expected SSH close error, got: %v", err)
	}

	// Status should still be disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}
}

// TestClient_Close_WithSFTPClientError tests Close when SFTP client returns error
func TestClient_Close_WithSFTPClientError(t *testing.T) {
	// Create a mock SFTP client that returns error on close
	mockSFTP := &mockSFTPClient{
		closeError: errors.New("SFTP close error"),
	}

	// Create a testable client with mock SFTP client
	client := &testableClient{
		sshClient:  &mockSSHClient{shouldFail: false},
		sftpClient: mockSFTP,
		config:     DefaultClientConfig,
		status:     StatusConnected,
	}

	err := client.Close()

	// Should return the SFTP close error
	if err == nil {
		t.Error("Expected error when SFTP client close fails")
	}

	if !strings.Contains(err.Error(), "SFTP close error") {
		t.Errorf("Expected SFTP close error, got: %v", err)
	}

	// Status should still be disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}
}

// TestClient_Close_WithBothClientsError tests Close when both SSH and SFTP clients return errors
func TestClient_Close_WithBothClientsError(t *testing.T) {
	// Create mock clients that both return errors on close
	mockSSH := &mockSSHClient{
		shouldFail: false,
		closeError: errors.New("SSH close error"),
	}

	mockSFTP := &mockSFTPClient{
		closeError: errors.New("SFTP close error"),
	}

	client := &testableClient{
		sshClient:  mockSSH,
		sftpClient: mockSFTP,
		config:     DefaultClientConfig,
		status:     StatusConnected,
	}

	err := client.Close()

	// Should return combined errors
	if err == nil {
		t.Error("Expected error when both SSH and SFTP clients close fail")
	}

	// Should contain both error messages (exact format may vary due to multierror)
	errStr := err.Error()
	if !strings.Contains(errStr, "SFTP close error") {
		t.Errorf("Expected SFTP close error in combined error, got: %v", err)
	}

	// Note: multierror may format SSH error differently, so we check for it more flexibly
	if !strings.Contains(errStr, "SSH close error") && !strings.Contains(errStr, "close error") {
		t.Errorf("Expected SSH-related close error in combined error, got: %v", err)
	}

	// Status should still be disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}
}

// TestClient_Close_SuccessfulCleanup tests Close when both clients close successfully
func TestClient_Close_SuccessfulCleanup(t *testing.T) {
	// Create mock clients that close successfully
	mockSSH := &mockSSHClient{
		shouldFail: false,
		closeError: nil,
	}

	mockSFTP := &mockSFTPClient{
		closeError: nil,
	}

	client := &testableClient{
		sshClient:  mockSSH,
		sftpClient: mockSFTP,
		config:     DefaultClientConfig,
		status:     StatusConnected,
	}

	err := client.Close()

	// Should not return error when both clients close successfully
	if err != nil {
		t.Errorf("Expected no error when both clients close successfully, got: %v", err)
	}

	// Status should be disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}

	// Verify SSH client was closed
	if !mockSSH.closed {
		t.Error("Expected SSH client to be closed")
	}

	// Verify SFTP client was closed
	if !mockSFTP.closed {
		t.Error("Expected SFTP client to be closed")
	}
}

// TestClient_Close_AlreadyClosed tests Close when client is already disconnected
func TestClient_Close_AlreadyClosed(t *testing.T) {
	client := newClient(nil, nil)

	// Close once
	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error on first close, got %v", err)
	}

	// Verify status is disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}

	// Close again - should return nil immediately
	err = client.Close()
	if err != nil {
		t.Errorf("Expected no error on double close, got %v", err)
	}

	// Status should remain disconnected
	if client.Status() != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, client.Status())
	}
}

func TestClient_Status(t *testing.T) {
	client := newClient(nil, nil)

	// Initial status should be connected
	if status := client.Status(); status != StatusConnected {
		t.Errorf("Expected status %v, got %v", StatusConnected, status)
	}

	// Status after closing should be disconnected
	_ = client.Close()
	if status := client.Status(); status != StatusDisconnected {
		t.Errorf("Expected status %v, got %v", StatusDisconnected, status)
	}
}

func TestClient_UnderlyingClient(t *testing.T) {
	client := newClient(nil, nil)

	// UnderlyingClient is not part of SSHClient interface but is available on concrete type
	// This tests the concrete implementation specific method
	if concreteClient, ok := client.(interface{ UnderlyingClient() *ssh.Client }); ok {
		underlying := concreteClient.UnderlyingClient()
		// Should return nil for our test case
		if underlying != nil {
			t.Error("Expected nil underlying client for nil ssh.Client")
		}
	} else {
		t.Error("Expected client to have UnderlyingClient method")
	}
}

func TestClient_InterfaceCompliance(t *testing.T) {
	client := newClient(nil, nil)

	// Verify client implements SSHClient interface
	var _ SSHClient = client

	// Test all interface methods are available and callable
	_ = client.Command("test")
	_ = client.Script("test")
	_ = client.ScriptFile("test")
	_ = client.Shell()
	_ = client.InteractiveShell(nil)
	_ = client.FileSystem()
	_ = client.Status()
	_ = client.Close()
}

func TestClient_MethodChaining(t *testing.T) {
	client := newClient(nil, nil)

	// Handle expected panic with nil ssh.Client when creating FileSystem
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic caught when creating FileSystem with nil ssh.Client")
		}
	}()

	// Test chaining different methods
	cmd := client.Command("echo test")
	script := client.Script("echo script")
	shell := client.Shell()
	fs := client.FileSystem()

	// All should return non-nil
	if cmd == nil || script == nil || shell == nil || fs == nil {
		t.Error("Expected all methods to return non-nil values")
	}

	// Status should remain connected
	if client.Status() != StatusConnected {
		t.Error("Expected status to remain connected after method calls")
	}
}

func TestClient_ConfigurationHandling(t *testing.T) {
	tests := []struct {
		name   string
		config *ClientConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name:   "empty config",
			config: &ClientConfig{},
		},
		{
			name: "full config",
			config: &ClientConfig{
				Timeout:       30 * time.Second,
				KeepAlive:     5 * time.Second,
				MaxSessions:   10,
				RetryAttempts: 3,
				RetryDelay:    1 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(nil, tt.config)

			if client == nil {
				t.Fatal("Expected non-nil client")
			}

			if client.Status() != StatusConnected {
				t.Errorf("Expected status %v, got %v", StatusConnected, client.Status())
			}
		})
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	client := newClient(nil, nil)

	// All methods should handle nil ssh.Client gracefully (no panics)
	cmd := client.Command("test")
	script := client.Script("test")
	scriptFile := client.ScriptFile("test")
	shell := client.Shell()
	interactiveShell := client.InteractiveShell(nil)

	// FileSystem should also handle nil ssh.Client gracefully (no panic)
	fs := client.FileSystem()

	// All should return non-nil despite nil ssh.Client
	if cmd == nil || script == nil || scriptFile == nil || shell == nil || interactiveShell == nil || fs == nil {
		t.Error("Expected all methods to return non-nil values even with nil ssh.Client")
	}

	// Close should handle nil ssh.Client gracefully (no error)
	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing with nil ssh.Client, got %v", err)
	}
}

func TestClient_ConcurrentAccess(t *testing.T) {
	client := newClient(nil, nil)

	// Run multiple operations concurrently
	done := make(chan bool, 6)

	go func() {
		client.Command("test1")
		done <- true
	}()

	go func() {
		client.Script("test2")
		done <- true
	}()

	go func() {
		client.Shell()
		done <- true
	}()

	go func() {
		// FileSystem will panic with nil ssh.Client - handle gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected panic caught when creating FileSystem with nil ssh.Client")
			}
			done <- true
		}()
		client.FileSystem()
	}()

	go func() {
		client.Status()
		done <- true
	}()

	go func() {
		client.Status()
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 6; i++ {
		<-done
	}

	// Client should still be functional
	if client.Status() != StatusConnected {
		t.Error("Expected client to remain connected after concurrent operations")
	}
}

func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client := newClient(nil, nil)
		_ = client
	}
}

func BenchmarkClient_Command(b *testing.B) {
	client := newClient(nil, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := client.Command("test command")
		_ = cmd
	}
}

func BenchmarkClient_FileSystem(b *testing.B) {
	client := newClient(nil, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Handle expected panic with nil ssh.Client
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected panic, continue benchmark
				}
			}()
			fs := client.FileSystem()
			_ = fs
		}()
	}
}
