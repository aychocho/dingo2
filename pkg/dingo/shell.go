package dingo

import (
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

// remoteShell implements the Shell interface
type remoteShell struct {
	client         *ssh.Client
	shellType      ShellType
	requestPty     bool
	terminalConfig *TerminalConfig

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

/*
* Initiates the shell session, sets up streams, requests PTY if needed, and waits for completion
* Inputs: command - A command to run interactively instead of the shell, if command is "" the regular shell is run
* Outputs: error if session creation, PTY request, or shell startup fails, nil on successful completion
 */
func (rs *remoteShell) Start(command string) error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Set up input/output streams
	rs.setupStreams(session)

	// Request PTY if needed (for interactive shells)
	if rs.requestPty {
		if err := rs.requestPseudoTerminal(session); err != nil {
			return err
		}
	}

	// If command is provided start the command
	if command != "" {
		session.Run(command)
	}

	// Start the shell
	if err := session.Shell(); err != nil {
		return err
	}

	// Wait for the session to complete
	return session.Wait()
}

/*
* Configures custom input/output streams for the shell session
* Inputs: stdin (io.Reader) - input stream, stdout (io.Writer) - output stream, stderr (io.Writer) - error stream
* Outputs: Shell interface for method chaining
 */
func (rs *remoteShell) SetStdio(stdin io.Reader, stdout, stderr io.Writer) Shell {
	rs.stdin = stdin
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}

/*
* Internal helper that configures session streams using either custom streams or OS defaults
* Inputs: session (*ssh.Session) - SSH session to configure
* Outputs: none (modifies session streams)
 */
func (rs *remoteShell) setupStreams(session *ssh.Session) {
	// Set up stdin
	if rs.stdin == nil {
		session.Stdin = os.Stdin
	} else {
		session.Stdin = rs.stdin
	}

	// Set up stdout
	if rs.stdout == nil {
		session.Stdout = os.Stdout
	} else {
		session.Stdout = rs.stdout
	}

	// Set up stderr
	if rs.stderr == nil {
		session.Stderr = os.Stderr
	} else {
		session.Stderr = rs.stderr
	}
}

/*
* Internal helper that requests a pseudo-terminal for interactive shell sessions
* Inputs: session (*ssh.Session) - SSH session to request PTY for
* Outputs: error if PTY request fails, nil on success
 */
func (rs *remoteShell) requestPseudoTerminal(session *ssh.Session) error {
	tc := rs.terminalConfig
	if tc == nil {
		tc = DefaultTerminalConfig
	}

	return session.RequestPty(tc.Term, tc.Height, tc.Width, tc.Modes)
}

/*
* Creates an interactive shell with custom terminal configuration and dimensions
* Inputs: term (string) - terminal type, width (int) - terminal width, height (int) - terminal height, modes (ssh.TerminalModes) - terminal modes
* Outputs: Shell interface configured for interactive use with custom terminal settings
 */
func (c *client) ShellWithCustomTerminal(term string, width, height int, modes ssh.TerminalModes) Shell {
	config := &TerminalConfig{
		Term:   term,
		Width:  width,
		Height: height,
		Modes:  modes,
	}

	return &remoteShell{
		client:         c.sshClient,
		shellType:      InteractiveShell,
		requestPty:     true,
		terminalConfig: config,
	}
}

/*
* Creates a non-interactive shell with minimal configuration for quick operations
* Inputs: sshClient (*ssh.Client) - established SSH connection
* Outputs: Shell interface configured for non-interactive use
 */
func QuickShell(sshClient *ssh.Client) Shell {
	return &remoteShell{
		client:     sshClient,
		shellType:  NonInteractiveShell,
		requestPty: false,
	}
}

/*
* Creates an interactive shell with default terminal settings for standard interactive sessions
* Inputs: sshClient (*ssh.Client) - established SSH connection
* Outputs: Shell interface configured for interactive use with default terminal settings
 */
func InteractiveShellWithDefaults(sshClient *ssh.Client) Shell {
	return &remoteShell{
		client:         sshClient,
		shellType:      InteractiveShell,
		requestPty:     true,
		terminalConfig: DefaultTerminalConfig,
	}
}
