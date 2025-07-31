package dingo

import (
	"testing"
)

/*
* Tests that all required types and interfaces are properly defined and can be assigned
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if type definitions are incorrect)
 */
func TestTypes(t *testing.T) {
	// Test that all types are properly defined
	var client SSHClient
	var cmd CommandExecutor
	var shell Shell
	var fs FileSystem

	// These should not cause compilation errors
	_ = client
	_ = cmd
	_ = shell
	_ = fs
}

/*
* Tests that SFTP option functions work correctly and apply configurations as expected
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if SFTP options don't work correctly)
 */
func TestSftpOptions(t *testing.T) {
	// Test SFTP option functions for single-threaded operation
	config := &SftpConfig{}

	opts := []SftpOption{
		WithMaxPacket(32768),
		WithFstat(true),
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.MaxPacket != 32768 {
		t.Errorf("Expected MaxPacket to be 32768, got %d", config.MaxPacket)
	}

	if !config.UseFstat {
		t.Error("Expected UseFstat to be true")
	}
}

/*
* Tests that SFTP option preset functions return the correct configuration combinations
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if preset options don't contain expected values)
 */
func TestSftpPresets(t *testing.T) {
	// Test preset option functions
	fastOpts := FastSftpOptions()
	safeOpts := SafeSftpOptions()
	defaultOpts := DefaultSftpOptions()

	// All presets should return some options
	if len(fastOpts) == 0 || len(safeOpts) == 0 || len(defaultOpts) == 0 {
		t.Error("Preset options should not be empty")
	}

	// Test that presets can be applied
	config := &SftpConfig{}
	for _, opt := range fastOpts {
		opt(config)
	}
}

/*
* Tests that default configuration constants are properly defined and accessible
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if default configurations are not properly defined)
 */
func TestDefaultConfigs(t *testing.T) {
	// Test that default configurations exist
	if DefaultClientConfig == nil {
		t.Error("DefaultClientConfig should not be nil")
	}

	if DefaultTerminalConfig == nil {
		t.Error("DefaultTerminalConfig should not be nil")
	}

	// Test terminal config values
	if DefaultTerminalConfig.Width <= 0 {
		t.Error("DefaultTerminalConfig.Width should be positive")
	}

	if DefaultTerminalConfig.Height <= 0 {
		t.Error("DefaultTerminalConfig.Height should be positive")
	}

	if DefaultTerminalConfig.Term == "" {
		t.Error("DefaultTerminalConfig.Term should not be empty")
	}

	if DefaultTerminalConfig.Modes == nil {
		t.Error("DefaultTerminalConfig.Modes should not be nil")
	}

	// Test client config
	if DefaultClientConfig.Timeout <= 0 {
		t.Error("DefaultClientConfig.Timeout should be positive")
	}

	if DefaultClientConfig.KeepAlive <= 0 {
		t.Error("DefaultClientConfig.KeepAlive should be positive")
	}

	if DefaultClientConfig.MaxSessions <= 0 {
		t.Error("DefaultClientConfig.MaxSessions should be positive")
	}
}

/*
* Tests that connection status enumeration values are properly defined
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if connection status values are not properly defined)
 */
func TestConnectionStatus(t *testing.T) {
	// Test connection status values
	statuses := []ConnectionStatus{
		StatusConnected,
		StatusDisconnected,
		StatusConnecting,
		StatusError,
	}

	expectedValues := []string{
		"connected",
		"disconnected",
		"connecting",
		"error",
	}

	for i, status := range statuses {
		if string(status) != expectedValues[i] {
			t.Errorf("Expected status %d to be '%s', got '%s'", i, expectedValues[i], string(status))
		}
	}
}

/*
* Tests that script type enumeration values are properly defined
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if script type values are not properly defined)
 */
func TestScriptTypes(t *testing.T) {
	// Test script type values
	types := []ScriptType{
		CommandLine,
		RawScript,
		ScriptFile,
	}

	// Test that values are sequential starting from 0
	for i, scriptType := range types {
		if byte(scriptType) != byte(i) {
			t.Errorf("Expected ScriptType %d to have value %d, got %d", i, i, byte(scriptType))
		}
	}
}

/*
* Tests that shell type enumeration values are properly defined
* Inputs: t (*testing.T) - testing context
* Outputs: none (fails test if shell type values are not properly defined)
 */
func TestShellTypes(t *testing.T) {
	// Test shell type values
	types := []ShellType{
		InteractiveShell,
		NonInteractiveShell,
	}

	// Test that values are sequential starting from 0
	for i, shellType := range types {
		if byte(shellType) != byte(i) {
			t.Errorf("Expected ShellType %d to have value %d, got %d", i, i, byte(shellType))
		}
	}
}
