package dingo

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// client implements the SSHClient interface with single-threaded design
type client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client // Single SFTP session instead of sync.Map
	config     *ClientConfig
	status     ConnectionStatus
}

/*
* Creates a new single-threaded SSH client instance with the provided SSH connection
* Inputs: sshClient (*ssh.Client) - established SSH connection, config (*ClientConfig) - client configuration or nil for defaults
* Outputs: SSHClient interface implementation
 */
func newClient(sshClient *ssh.Client, config *ClientConfig) SSHClient {
	if config == nil {
		config = DefaultClientConfig
	}

	return &client{
		sshClient: sshClient,
		config:    config,
		status:    StatusConnected,
	}
}

/*
* Creates a CommandExecutor for executing a single command on the remote server
* Inputs: cmd (string) - the command to execute
* Outputs: CommandExecutor interface for running the command
 */
func (c *client) Command(cmd string) CommandExecutor {
	return &remoteScript{
		client:     c.sshClient,
		scriptType: CommandLine,
		script:     cmd,
	}
}

/*
* Creates a CommandExecutor for executing a raw shell script on the remote server
* Inputs: script (string) - the shell script content to execute
* Outputs: CommandExecutor interface for running the script
 */
func (c *client) Script(script string) CommandExecutor {
	return &remoteScript{
		client:     c.sshClient,
		scriptType: RawScript,
		script:     script,
	}
}

/*
* Creates a CommandExecutor for executing a script file on the remote server
* Inputs: path (string) - local path to the script file to execute
* Outputs: CommandExecutor interface for running the script file
 */
func (c *client) ScriptFile(path string) CommandExecutor {
	return &remoteScript{
		client:     c.sshClient,
		scriptType: ScriptFile,
		scriptFile: path,
	}
}

/*
* Creates a non-interactive shell session without PTY support
* Inputs: none
* Outputs: Shell interface for non-interactive shell operations
 */
func (c *client) Shell() Shell {
	return &remoteShell{
		client:     c.sshClient,
		shellType:  NonInteractiveShell,
		requestPty: false,
	}
}

/*
* Creates an interactive shell session with PTY support and terminal configuration
* Inputs: config (*TerminalConfig) - terminal configuration or nil for defaults
* Outputs: Shell interface for interactive shell operations
 */
func (c *client) InteractiveShell(config *TerminalConfig) Shell {
	if config == nil {
		config = DefaultTerminalConfig
	}

	return &remoteShell{
		client:         c.sshClient,
		shellType:      InteractiveShell,
		requestPty:     true,
		terminalConfig: config,
	}
}

/*
* Creates a FileSystem interface for SFTP operations with optional configuration
* Inputs: opts (...SftpOption) - variadic SFTP configuration options
* Outputs: FileSystem interface for remote file operations
 */
func (c *client) FileSystem(opts ...SftpOption) FileSystem {
	// Apply configuration options first
	config := &SftpConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Create SFTP client if not already created
	if c.sftpClient == nil {
		if c.sshClient == nil {
			return &remoteFileSystem{
				client: c.sshClient,
				sftp:   nil,
				config: config,
				err:    fmt.Errorf("SSH client is nil"),
			}
		}
		var err error
		c.sftpClient, err = sftp.NewClient(c.sshClient)
		if err != nil {
			return &remoteFileSystem{
				client: c.sshClient,
				sftp:   nil,
				config: config,
				err:    err,
			}
		}
	}

	return &remoteFileSystem{
		client: c.sshClient,
		sftp:   c.sftpClient,
		config: config,
		err:    nil,
	}
}

/*
* Closes the SSH connection and all associated resources including SFTP sessions
* Inputs: none
* Outputs: error if any issues during cleanup, nil on success
 */
func (c *client) Close() error {
	if c.status == StatusDisconnected {
		return nil
	}

	var merr *multierror.Error

	// Close SFTP client if exists
	if c.sftpClient != nil {
		merr = multierror.Append(merr, c.sftpClient.Close())
	}

	// Close the main SSH connection if exists
	if c.sshClient != nil {
		merr = multierror.Append(merr, c.sshClient.Close())
	}

	c.status = StatusDisconnected
	return merr.ErrorOrNil()
}

/*
* Returns the current connection status of the SSH client
* Inputs: none
* Outputs: ConnectionStatus enum (connected, disconnected, connecting, error)
 */
func (c *client) Status() ConnectionStatus {
	return c.status
}

/*
* Returns the underlying ssh.Client for advanced operations not covered by the interface
* Inputs: none
* Outputs: *ssh.Client - the raw SSH client connection
 */
func (c *client) UnderlyingClient() *ssh.Client {
	return c.sshClient
}
