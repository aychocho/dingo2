package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Quok-it/dingo/pkg/dingo"
)

/*
* Main entry point for the dingo CLI application that handles SSH connections and operations
* Inputs: none (reads from command line flags)
* Outputs: none (exits with status code 0 on success, 1 on error)
 */
func main() {
	var (
		host       = flag.String("host", "", "SSH host (e.g., user@hostname:port)")
		ip         = flag.String("ip", "", "Target IP address")
		port       = flag.String("port", "22", "SSH port")
		username   = flag.String("user", "user", "SSH username")
		password   = flag.String("password", "", "SSH password")
		keyFile    = flag.String("key", "", "SSH private key file (defaults to ~/.ssh/id_rsa)")
		command    = flag.String("cmd", "", "Command to execute")
		upload     = flag.String("upload", "", "Upload file (format: local:remote)")
		download   = flag.String("download", "", "Download file (format: remote:local)")
		persistent = flag.Bool("persistent", false, "Keep connection alive for continuous operation")
		interval   = flag.Duration("interval", 30*time.Second, "Interval between operations in persistent mode")
		script     = flag.String("script", "", "Script file to execute")
		shell      = flag.Bool("shell", false, "Start interactive shell")
		footprint  = flag.Bool("footprint", false, "Upload and execute footprint script")
		hostname   = flag.String("hostname", "", "Hostname to include in footprint (defaults to current hostname)")
		tail       = flag.String("tail", "", "Tail a file (e.g., /var/log/syslog)")
		follow     = flag.Bool("follow", false, "Follow file changes (like tail -f)")
		restore    = flag.Bool("restore", false, "Restore a previously started session, needs ip-address")
		no_restore = flag.Bool("no_restore", false, "Don't enable session restoration, automatically set to true when screen is not installed")
		lines      = flag.Int("lines", 10, "Number of lines to show initially when tailing")
		stream     = flag.Bool("stream", false, "Stream command output in real-time with separate stdout/stderr")
	)
	flag.Parse()

	// Handle new IP/port style or traditional host style
	var hostAddr string
	if *ip != "" {
		hostAddr = fmt.Sprintf("%s@%s:%s", *username, *ip, *port)
	} else if *host != "" {
		hostAddr = *host
	} else {
		fmt.Fprintf(os.Stderr, "Error: either -ip or -host is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Default to ~/.ssh/id_rsa if no key specified
	keyPath := *keyFile
	if keyPath == "" && *password == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		keyPath = filepath.Join(homeDir, ".ssh", "id_rsa")

		// Check if the default key exists
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Default key %s not found and no password provided\n", keyPath)
			fmt.Fprintf(os.Stderr, "Use -key or -password to specify authentication\n")
			os.Exit(1)
		}
		fmt.Printf("Using default SSH key: %s\n", keyPath)
	}

	// Connect to SSH server
	client, err := connect(hostAddr, *password, keyPath)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	if (*restore || !*no_restore) && *ip == "" {
		fmt.Fprintf(os.Stderr, "Warning: Session restoration functionality requires the -ip flag. Disabling.\n")
		*no_restore = true
	}

	var sessionName string
	useScreen := !*no_restore

	if useScreen {
		// Check if screen is available on the remote host.
		if !isScreenInstalled(client) {
			fmt.Println("Warning: 'screen' is not installed on the remote host. Disabling session restore")
			useScreen = false
		} else {
			sessionName = "dingo"
		}
	}

	if *restore {
		if !useScreen {
			log.Fatalf("Cannot restore session: 'screen' is unavailable or session features were disabled.")
		}
		err = handleRestore(client, sessionName)
		if err != nil {
			log.Fatalf("Failed to restore session '%s': %v", sessionName, err)
		}
		fmt.Printf("Detached from session '%s'.\n", sessionName)
		return // Exit after the restore attempt.
	}

	// Handle footprint mode
	if *footprint {
		err = handleFootprint(client, *hostname)
		if err != nil {
			log.Fatalf("Footprint operation failed: %v", err)
		}
		return
	}

	// Handle tail mode
	if *tail != "" {
		err = handleTail(client, *tail, *follow, *lines)
		if err != nil {
			log.Fatalf("Tail operation failed: %v", err)
		}
		return
	}

	// Execute based on mode
	if *persistent {
		err = runPersistentMode(client, *command, *interval)
	} else {
		err = runSingleMode(client, *command, *upload, *download, *script, *shell, *stream, useScreen, sessionName)
	}

	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}

/*
* Handles footprint operation - uploads and executes a script that leaves a trace file
* Inputs: client (dingo.SSHClient) - established SSH connection, hostname (string) - hostname to include in footprint
* Outputs: error if footprint operation fails, nil on success
 */
func handleFootprint(client dingo.SSHClient, hostname string) error {
	// Get hostname if not provided
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
	}

	// Create footprint script content
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	scriptContent := fmt.Sprintf(`#!/bin/bash

# Dingo SSH Footprint Script
echo "=== Dingo SSH Footprint ==="
echo "Executed by: %s"
echo "Timestamp: %s"
echo "Target system: $(hostname)"
echo "Current user: $(whoami)"
echo "Working directory: $(pwd)"
echo "System info: $(uname -a)"

# Create footprint file
FOOTPRINT_FILE="/tmp/dingo_footprint_$(date +%%s).txt"
cat > "$FOOTPRINT_FILE" << EOF
Dingo SSH Client Footprint
==========================
Source hostname: %s
Execution time: %s
Target hostname: $(hostname)
Target user: $(whoami)
Target system: $(uname -a)
Working directory: $(pwd)
Process ID: $$
Random ID: $RANDOM
EOF

echo ""
echo "Footprint file created: $FOOTPRINT_FILE"
echo "File contents:"
cat "$FOOTPRINT_FILE"
echo ""
echo "File permissions:"
ls -la "$FOOTPRINT_FILE"
echo ""
echo "✓ Footprint operation completed successfully!"
`, hostname, timestamp, hostname, timestamp)

	fmt.Printf("Creating footprint script for hostname: %s\n", hostname)

	// Write script to temporary file
	tmpScript := "/tmp/dingo_footprint_script.sh"
	fs := client.FileSystem()
	defer fs.Close()

	// Upload the script
	err := fs.WriteFile(tmpScript, []byte(scriptContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to upload footprint script: %v", err)
	}

	fmt.Printf("✓ Footprint script uploaded to: %s\n", tmpScript)

	// Explicitly set executable permissions
	err = fs.Chmod(tmpScript, 0755)
	if err != nil {
		return fmt.Errorf("failed to set executable permissions: %v", err)
	}

	fmt.Println("✓ Executable permissions set")

	// Verify the script exists and has correct permissions
	fileInfo, err := fs.Stat(tmpScript)
	if err != nil {
		return fmt.Errorf("failed to verify script file: %v", err)
	}
	fmt.Printf("✓ Script verified: %s (permissions: %s)\n", fileInfo.Name(), fileInfo.Mode())

	// Execute the script using bash explicitly
	fmt.Println("Executing footprint script...")
	cmd := client.Command(fmt.Sprintf("bash %s", tmpScript))
	output, err := cmd.Output()
	if err != nil {
		// Try alternative execution methods
		fmt.Println("Direct bash execution failed, trying alternative...")

		// Try making it executable and running directly
		err2 := client.Command(fmt.Sprintf("chmod +x %s && %s", tmpScript, tmpScript)).Run()
		if err2 != nil {
			return fmt.Errorf("failed to execute footprint script: %v (alternative method also failed: %v)", err, err2)
		}

		// Get output from successful alternative execution
		output, err = client.Command(tmpScript).Output()
		if err != nil {
			return fmt.Errorf("script executed but failed to get output: %v", err)
		}
	}

	fmt.Printf("Script output:\n%s\n", string(output))

	return nil
}

/*
* Establishes SSH connection using provided credentials (password or key-based authentication)
* Inputs: host (string) - SSH server address, password (string) - password or empty, keyFile (string) - private key path or empty
* Outputs: dingo.SSHClient interface, error if connection fails
 */
func connect(host, password, keyFile string) (dingo.SSHClient, error) {
	username := extractUser(host)
	cleanHost := extractHostAddr(host)

	if keyFile != "" {
		return dingo.ConnectWithKey(cleanHost, username, keyFile)
	}
	if password != "" {
		return dingo.ConnectWithPassword(cleanHost, username, password)
	}
	return nil, fmt.Errorf("either password or key file must be provided")
}

/*
* Extracts username from SSH host string format (user@hostname:port)
* Inputs: host (string) - SSH host string in format user@hostname:port
* Outputs: string containing extracted username
 */
func extractUser(host string) string {
	parts := strings.Split(host, "@")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root" // Default fallback
}

/*
* Extracts clean host address from SSH host string format (user@hostname:port)
* Inputs: host (string) - SSH host string in format user@hostname:port or hostname:port
* Outputs: string containing clean host address (hostname:port)
 */
func extractHostAddr(host string) string {
	// If there's an @ symbol, extract everything after it
	if strings.Contains(host, "@") {
		parts := strings.Split(host, "@")
		if len(parts) > 1 {
			return parts[1] // Return hostname:port part
		}
	}
	// If no @ symbol, return the host as-is
	return host
}

/*
* Runs the application in persistent mode, executing commands at regular intervals
* Inputs: client (dingo.SSHClient) - established SSH connection, command (string) - command to execute, interval (time.Duration) - time between executions
* Outputs: error if persistent operations fail, nil on graceful shutdown
 */
func runPersistentMode(client dingo.SSHClient, command string, interval time.Duration) error {
	fmt.Printf("Running in persistent mode (interval: %v)\n", interval)

	// If no command specified, just keep connection alive
	if command == "" {
		fmt.Println("Keeping connection alive... Press Ctrl+C to exit")
		for {
			time.Sleep(interval)
			// Check connection status
			if status := client.Status(); status != dingo.StatusConnected {
				return fmt.Errorf("connection lost: %v", status)
			}
		}
	}

	// Execute command periodically
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("Executing command: %s\n", command)
			output, err := client.Command(command).Output()
			if err != nil {
				log.Printf("Command failed: %v", err)
				continue
			}
			fmt.Printf("Output: %s\n", string(output))
		}
	}
}

/*
* Runs the application in single-operation mode, executing one task and exiting
* Inputs: client (dingo.SSHClient) - established SSH connection, command (string) - command to execute, upload (string) - upload spec, download (string) - download spec, script (string) - script file path, shell (bool) - whether to start interactive shell, stream (bool) - whether to stream output
* Outputs: error if any operation fails, nil on successful completion
 */
func runSingleMode(client dingo.SSHClient, command, upload, download, script string, shell bool, stream bool, useScreen bool, sessionName string) error {
	// Handle file operations
	if upload != "" {
		return handleUpload(client, upload)
	}
	if download != "" {
		return handleDownload(client, download)
	}

	// Handle script execution
	if script != "" {
		return handleScript(client, script)
	}

	// Handle interactive shell
	if shell {
		return handleShell(client, useScreen, sessionName)
	}

	// Handle command execution
	if command != "" {
		if stream {
			return handleStreamCommand(client, command)
		}
		return handleCommand(client, command)
	}

	return fmt.Errorf("no operation specified")
}

/*
* Handles file upload operation from local to remote server
* Inputs: client (dingo.SSHClient) - established SSH connection, upload (string) - upload specification in format "local:remote"
* Outputs: error if upload fails, nil on successful upload
 */
func handleUpload(client dingo.SSHClient, upload string) error {
	parts := strings.Split(upload, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid upload format, use: local:remote")
	}

	fmt.Printf("Uploading %s to %s\n", parts[0], parts[1])
	fs := client.FileSystem()
	defer fs.Close()

	return fs.Upload(parts[0], parts[1])
}

/*
* Handles file download operation from remote server to local filesystem
* Inputs: client (dingo.SSHClient) - established SSH connection, download (string) - download specification in format "remote:local"
* Outputs: error if download fails, nil on successful download
 */
func handleDownload(client dingo.SSHClient, download string) error {
	parts := strings.Split(download, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid download format, use: remote:local")
	}

	fmt.Printf("Downloading %s to %s\n", parts[0], parts[1])
	fs := client.FileSystem()
	defer fs.Close()

	return fs.Download(parts[0], parts[1])
}

/*
* Handles script file execution on the remote server
* Inputs: client (dingo.SSHClient) - established SSH connection, script (string) - path to local script file
* Outputs: error if script execution fails, nil on successful execution
 */
func handleScript(client dingo.SSHClient, script string) error {
	fmt.Printf("Executing script: %s\n", script)
	return client.ScriptFile(script).Run()
}

/*
* Handles interactive shell session with the remote server
* Inputs: client (dingo.SSHClient) - established SSH connection
* Outputs: error if shell startup fails, nil on successful shell session completion
 */
func handleShell(client dingo.SSHClient, useScreen bool, sessionName string) error {
	fmt.Println("Starting interactive shell...")
	shell := client.InteractiveShell(nil)
	command := ""
	if useScreen {
		command = fmt.Sprintf("screen -mS %s", sessionName)
	}
	return shell.Start(command)
}

/*
* Handles single command execution on the remote server
* Inputs: client (dingo.SSHClient) - established SSH connection, command (string) - command to execute
* Outputs: error if command execution fails, nil on successful execution
 */
func handleCommand(client dingo.SSHClient, command string) error {
	fmt.Printf("Executing command: %s\n", command)
	output, err := client.Command(command).SmartOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Error output: %s\n", string(output))
		return err
	}
	fmt.Printf("Output: %s\n", string(output))
	return nil
}

/*
* Handles file tailing operation - monitors a file for changes and displays new content
* Inputs: client (dingo.SSHClient) - established SSH connection, filename (string) - file to tail, follow (bool) - whether to follow changes, lines (int) - initial lines to show
* Outputs: error if tail operation fails, nil on completion
 */
func handleTail(client dingo.SSHClient, filename string, follow bool, lines int) error {
	fmt.Printf("Tailing file: %s\n", filename)

	if follow {
		fmt.Printf("Following changes (Ctrl+C to stop)...\n")
		cmd := client.Command(fmt.Sprintf("tail -f -n %d %s", lines, filename))
		cmd.SetStdio(os.Stdout, os.Stderr)
		return cmd.Run()
	} else {
		fmt.Printf("Showing last %d lines:\n", lines)
		cmd := client.Command(fmt.Sprintf("tail -n %d %s", lines, filename))
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to tail file: %v", err)
		}
		fmt.Printf("%s", string(output))
		return nil
	}
}

/*
* Handles streaming command execution with separate stdout/stderr display
* Inputs: client (dingo.SSHClient) - established SSH connection, command (string) - command to execute
* Outputs: error if command execution fails, nil on successful execution
* Notes: sessionName for all sessions is currently just dingo
 */
func handleStreamCommand(client dingo.SSHClient, command string) error {
	fmt.Printf("Streaming command: %s\n", command)
	fmt.Println("--- STDOUT ---")

	cmd := client.Command(command)
	cmd.SetStdio(os.Stdout, os.Stderr)

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nCommand failed: %v\n", err)
		return err
	}

	fmt.Println("\n--- Command completed ---")
	return nil
}

/*
* Checks if the screen utilit is installed on the system
* Inputs: client (dingo.SSHClient) - established SSH connection
* Outputs: Boolean - True if screen is installed False if not
 */

func isScreenInstalled(client dingo.SSHClient) bool {
	// command -v since which is not installed by default
	output, err := client.Command("command -v screen").Output()
	return err == nil && len(strings.TrimSpace(string(output))) > 0
}

/*
* Handles restoring (attaching to) a named screen session.
* Inputs: client (dingo.SSHClient) - established SSH connection, sessionName (string) - the name of the screen session to restore
* Outputs: error if attaching fails, nil on success
 */
func handleRestore(client dingo.SSHClient, sessionName string) error {
	fmt.Printf("Attempting to restore screen session: %s\n", sessionName)
	fmt.Println("This will start an interactive session. Press Ctrl+A then D to detach.")

	// Use an interactive shell to run 'screen -r' which re-attaches to a session.
	screenCmd := fmt.Sprintf("screen -r %s", sessionName)
	shell := client.InteractiveShell(nil)
	return shell.Start(screenCmd)
}
