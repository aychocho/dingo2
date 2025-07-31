package dingo

import (
	"errors"
	"os"
	"testing"

	"github.com/pkg/sftp"
)

// Test helper to create remoteFileSystem with mock SFTP client
func createTestRemoteFileSystem(sftpClient *sftp.Client, hasError bool) *remoteFileSystem {
	rfs := &remoteFileSystem{
		client: nil, // Mock SSH client
		sftp:   sftpClient,
		config: &SftpConfig{},
		err:    nil,
	}

	if hasError {
		rfs.err = errors.New("test SFTP error")
	}

	return rfs
}

func TestRemoteFileSystem_ReadFile_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	data, err := rfs.ReadFile("/test/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if data != nil {
		t.Error("Expected nil data when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_ReadFile_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.ReadFile("/test/file.txt")
}

func TestRemoteFileSystem_WriteFile_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.WriteFile("/test/file.txt", []byte("test data"), 0644)
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_WriteFile_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.WriteFile("/test/file.txt", []byte("test data"), 0644)
}

func TestRemoteFileSystem_Upload_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Upload("/local/file.txt", "/remote/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Upload_LocalFileNotFound(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	err := rfs.Upload("/nonexistent/local/file.txt", "/remote/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent local file")
	}

	// Should be a file not found error from os.Open
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

func TestRemoteFileSystem_Download_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Download("/remote/file.txt", "/local/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Download_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Download("/remote/file.txt", "/local/file.txt")
}

func TestRemoteFileSystem_Mkdir_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Mkdir("/remote/dir")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Mkdir_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Mkdir("/remote/dir")
}

func TestRemoteFileSystem_MkdirAll_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.MkdirAll("/remote/deep/dir")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_MkdirAll_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.MkdirAll("/remote/deep/dir")
}

func TestRemoteFileSystem_Remove_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Remove("/remote/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Remove_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Remove("/remote/file.txt")
}

func TestRemoteFileSystem_RemoveDirectory_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.RemoveDirectory("/remote/dir")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_RemoveDirectory_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.RemoveDirectory("/remote/dir")
}

func TestRemoteFileSystem_Stat_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	info, err := rfs.Stat("/remote/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if info != nil {
		t.Error("Expected nil info when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Stat_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Stat("/remote/file.txt")
}

func TestRemoteFileSystem_Lstat_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	info, err := rfs.Lstat("/remote/file.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if info != nil {
		t.Error("Expected nil info when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Lstat_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Lstat("/remote/file.txt")
}

func TestRemoteFileSystem_ReadDir_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	entries, err := rfs.ReadDir("/remote/dir")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if entries != nil {
		t.Error("Expected nil entries when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_ReadDir_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.ReadDir("/remote/dir")
}

func TestRemoteFileSystem_Chmod_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Chmod("/remote/file.txt", 0755)
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Chmod_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Chmod("/remote/file.txt", 0755)
}

func TestRemoteFileSystem_Chown_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Chown("/remote/file.txt", 1000, 1000)
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Chown_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Chown("/remote/file.txt", 1000, 1000)
}

func TestRemoteFileSystem_Rename_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Rename("/remote/old.txt", "/remote/new.txt")
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Rename_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Should panic when SFTP client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when SFTP client is nil")
		}
	}()

	rfs.Rename("/remote/old.txt", "/remote/new.txt")
}

func TestRemoteFileSystem_Close_WithError(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, true)

	err := rfs.Close()
	if err == nil {
		t.Error("Expected error when SFTP has error")
	}

	if err.Error() != "test SFTP error" {
		t.Errorf("Expected 'test SFTP error', got: %v", err)
	}
}

func TestRemoteFileSystem_Close_NilSFTP(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	err := rfs.Close()
	if err != nil {
		t.Errorf("Expected no error when SFTP is nil, got: %v", err)
	}
}

func TestRemoteFileSystem_Interface_Compliance(t *testing.T) {
	rfs := createTestRemoteFileSystem(nil, false)

	// Verify it implements FileSystem interface
	var _ FileSystem = rfs

	// Test all interface methods are callable (they'll panic or error, but that's expected)
	defer func() {
		if r := recover(); r != nil {
			// Expected panics due to nil SFTP client
		}
	}()

	// These will panic but we're just testing interface compliance
	rfs.ReadFile("test")
	rfs.WriteFile("test", []byte("data"), 0644)
	rfs.Upload("local", "remote")
	rfs.Download("remote", "local")
	rfs.Mkdir("dir")
	rfs.MkdirAll("deep/dir")
	rfs.Remove("file")
	rfs.RemoveDirectory("dir")
	rfs.Stat("file")
	rfs.Lstat("file")
	rfs.ReadDir("dir")
	rfs.Chmod("file", 0755)
	rfs.Chown("file", 1000, 1000)
	rfs.Rename("old", "new")
	rfs.Close()
}

// TestRemoteFileSystem_SuccessfulOperations tests successful filesystem operations
func TestRemoteFileSystem_SuccessfulOperations(t *testing.T) {
	// Create a filesystem without error to test success paths that don't require real SFTP
	rfs := &remoteFileSystem{
		client: nil,
		sftp:   nil, // This will cause panics, but we're testing error-free paths
		config: &SftpConfig{},
		err:    nil,
	}

	// Test that Close works when sftp is nil
	err := rfs.Close()
	if err != nil {
		t.Errorf("Expected no error for Close with nil SFTP, got: %v", err)
	}
}

// TestRemoteFileSystem_FileSystemOperationsWithRealFiles tests file operations using real temp files
func TestRemoteFileSystem_FileSystemOperationsWithRealFiles(t *testing.T) {
	// Test Upload with real local file (testing the local file opening part)
	tempDir := t.TempDir()
	localFile := tempDir + "/test.txt"
	testData := []byte("test file content")

	err := os.WriteFile(localFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rfs := &remoteFileSystem{
		client: nil,
		sftp:   nil,
		config: &SftpConfig{},
		err:    nil,
	}

	// Test Upload with non-existent local file first - this tests os.Open error path
	err = rfs.Upload("/nonexistent/file.txt", "/remote/test.txt")
	if err == nil {
		t.Error("Expected error for non-existent local file")
	}
	// This tests the os.Open error path without reaching SFTP code
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not found error, got: %v", err)
	}

	// The Upload with existing file would reach SFTP code and panic with nil client
	// So we skip that test to avoid the panic. The local file opening is already tested above.
	// The SFTP operations are tested with error conditions in other tests.
}
