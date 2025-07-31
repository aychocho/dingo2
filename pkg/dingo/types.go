package dingo

import (
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient represents the main interface for SSH operations
type SSHClient interface {
	// Command execution
	Command(cmd string) CommandExecutor
	Script(script string) CommandExecutor
	ScriptFile(path string) CommandExecutor

	// Shell operations
	Shell() Shell
	InteractiveShell(config *TerminalConfig) Shell

	// File operations
	FileSystem(opts ...SftpOption) FileSystem

	// Connection management
	Close() error
	Status() ConnectionStatus
}

// CommandExecutor represents an interface for executing commands remotely
type CommandExecutor interface {
	Run() error
	Output() ([]byte, error)
	SmartOutput() ([]byte, error)
	SetStdio(stdout, stderr io.Writer) CommandExecutor
	Cmd(cmd string) CommandExecutor
}

// Shell represents an interface for interactive shell sessions
type Shell interface {
	Start(command string) error
	SetStdio(stdin io.Reader, stdout, stderr io.Writer) Shell
}

// FileSystem represents an interface for remote file operations
type FileSystem interface {
	// File operations
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Upload(localPath, remotePath string) error
	Download(remotePath, localPath string) error

	// Directory operations
	Mkdir(path string) error
	MkdirAll(path string) error
	Remove(path string) error
	RemoveDirectory(path string) error

	// File info operations
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)

	// File manipulation
	Chmod(path string, mode os.FileMode) error
	Chown(path string, uid, gid int) error
	Rename(oldname, newname string) error

	// Cleanup
	Close() error
}

// ConnectionStatus represents the status of an SSH connection
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusError        ConnectionStatus = "error"
)

// ClientConfig represents configuration for the SSH client
type ClientConfig struct {
	Timeout       time.Duration
	KeepAlive     time.Duration
	MaxSessions   int
	RetryAttempts int
	RetryDelay    time.Duration
}

// TerminalConfig represents configuration for interactive shell sessions
type TerminalConfig struct {
	Term   string
	Width  int
	Height int
	Modes  ssh.TerminalModes
}

// ScriptType represents the type of script execution
type ScriptType byte

const (
	CommandLine ScriptType = iota
	RawScript
	ScriptFile
)

// ShellType represents the type of shell session
type ShellType byte

const (
	InteractiveShell ShellType = iota
	NonInteractiveShell
)

// SftpOption represents a configuration option for SFTP operations
type SftpOption func(*SftpConfig)

// SftpConfig represents configuration for SFTP operations
type SftpConfig struct {
	MaxPacket int
	UseFstat  bool
}

// Default configurations
var (
	DefaultClientConfig = &ClientConfig{
		Timeout:       30 * time.Second,
		KeepAlive:     5 * time.Second,
		MaxSessions:   10,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	DefaultTerminalConfig = &TerminalConfig{
		Term:   "xterm",
		Width:  80,
		Height: 40,
		Modes:  ssh.TerminalModes{},
	}

	DefaultSftpConfig = &SftpConfig{
		MaxPacket: 32768,
		UseFstat:  true,
	}
)
