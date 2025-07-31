package dingo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// remoteScript implements the CommandExecutor interface
type remoteScript struct {
	client     *ssh.Client
	scriptType ScriptType
	script     string
	scriptFile string
	err        error

	stdout io.Writer
	stderr io.Writer
}

/*
* Executes the script on the remote server based on the script type (command, raw script, or script file)
* Inputs: none (uses internal script configuration)
* Outputs: error if execution fails, nil on success
 */
func (rs *remoteScript) Run() error {
	if rs.err != nil {
		fmt.Println(rs.err) // TODO: Use proper logging
		return rs.err
	}

	switch rs.scriptType {
	case CommandLine:
		return rs.runCommands()
	case RawScript:
		return rs.runScript()
	case ScriptFile:
		return rs.runScriptFile()
	default:
		return errors.New("unsupported script type")
	}
}

/*
* Executes the script and captures its standard output as bytes
* This optimized version preserves original state and supports multiple calls
* Inputs: none (uses internal script configuration)
* Outputs: []byte containing stdout, error if execution fails
 */
func (rs *remoteScript) Output() ([]byte, error) {
	// Early return if there's already an error
	if rs.err != nil {
		return nil, rs.err
	}

	// Save original stdout state
	originalStdout := rs.stdout

	// Ensure we restore original state regardless of how this function exits
	defer func() {
		rs.stdout = originalStdout
	}()

	// Use local buffer for this execution
	var out bytes.Buffer
	rs.stdout = &out

	// Execute the script
	err := rs.Run()
	return out.Bytes(), err
}

/*
* Executes the script and returns stdout on success or stderr on error for intelligent output handling
* This optimized version preserves original state and supports multiple calls
* Inputs: none (uses internal script configuration)
* Outputs: []byte containing stdout (success) or stderr (error), error if execution fails
 */
func (rs *remoteScript) SmartOutput() ([]byte, error) {
	// Early return if there's already an error
	if rs.err != nil {
		return nil, rs.err
	}

	// Save original stdout/stderr state
	originalStdout := rs.stdout
	originalStderr := rs.stderr

	// Ensure we restore original state regardless of how this function exits
	defer func() {
		rs.stdout = originalStdout
		rs.stderr = originalStderr
	}()

	// Use local buffers for this execution
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	// Temporarily set the buffers
	rs.stdout = &stdout
	rs.stderr = &stderr

	// Execute the script
	err := rs.Run()

	// Return appropriate output based on success/failure
	if err != nil {
		// On failure, return stderr (error details)
		return stderr.Bytes(), err
	}
	// On success, return stdout (normal output)
	return stdout.Bytes(), nil
}

/*
* Sets custom output streams for stdout and stderr capture during script execution
* Inputs: stdout (io.Writer) - writer for standard output, stderr (io.Writer) - writer for standard error
* Outputs: CommandExecutor interface for method chaining
 */
func (rs *remoteScript) SetStdio(stdout, stderr io.Writer) CommandExecutor {
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}

/*
* Appends a command to the script for sequential execution (only available for CommandLine script type)
* Inputs: cmd (string) - command to append to the execution sequence
* Outputs: CommandExecutor interface for method chaining
 */
func (rs *remoteScript) Cmd(cmd string) CommandExecutor {
	if rs.scriptType == CommandLine {
		if rs.script == "" {
			rs.script = cmd
		} else {
			rs.script += "\n" + cmd
		}
	} else {
		rs.err = errors.New("Cmd() can only be used with CommandLine script type")
	}
	return rs
}

/*
* Internal helper that executes multiple commands sequentially, one per line
* Inputs: none (uses internal script string)
* Outputs: error if any command fails, nil if all succeed
 */
func (rs *remoteScript) runCommands() error {
	commands := strings.Split(rs.script, "\n")

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		if err := rs.runSingleCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}

/*
* Internal helper that executes a single command in a new SSH session
* Inputs: cmd (string) - single command to execute
* Outputs: error if command execution fails, nil on success
 */
func (rs *remoteScript) runSingleCommand(cmd string) error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = rs.stdout
	session.Stderr = rs.stderr

	return session.Run(cmd)
}

/*
* Internal helper that executes a raw script by starting a shell and piping the script content
* Inputs: none (uses internal script string)
* Outputs: error if script execution fails, nil on success
 */
func (rs *remoteScript) runScript() error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = strings.NewReader(rs.script)
	session.Stdout = rs.stdout
	session.Stderr = rs.stderr

	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}

/*
* Internal helper that reads a local script file and executes its content remotely
* Inputs: none (uses internal scriptFile path)
* Outputs: error if file reading or script execution fails, nil on success
 */
func (rs *remoteScript) runScriptFile() error {
	file, err := os.Open(rs.scriptFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return err
	}

	// Temporarily change to RawScript mode to execute the file content
	originalScript := rs.script
	originalType := rs.scriptType

	rs.script = buffer.String()
	rs.scriptType = RawScript

	err = rs.runScript()

	// Restore original values
	rs.script = originalScript
	rs.scriptType = originalType

	return err
}
