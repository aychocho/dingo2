# Dingo - SSH Client for Go

SSH client library and command-line tool with command execution, file operations, and interactive shells.

## Install

```bash
go get github.com/Quok-it/dingo
```

## Build Binary

```bash
go build -o dingo ./cmd/dingo
```

## Command-Line Usage

### Authentication
```bash
# SSH key (defaults to ~/.ssh/id_rsa)
./dingo -ip 192.168.1.100 -user root -cmd "uptime"

# Specific key
./dingo -host user@server:22 -key /path/to/key -cmd "ps aux"

# Password
./dingo -ip server -user admin -password secret -cmd "df -h"
```

### Operations
```bash
# Execute command
./dingo -ip server -user root -cmd "systemctl status nginx"

# Stream command output 
./dingo -ip server -user root -cmd "find /var -name '*.log'" -stream

# Upload/download files
./dingo -ip server -user root -upload "local.txt:/tmp/remote.txt"
./dingo -ip server -user root -download "/var/log/app.log:./app.log"

# Run script file
./dingo -ip server -user root -script "./deploy.sh"

# Interactive shell
./dingo -ip server -user root -shell

# Tail files
./dingo -ip server -user root -tail "/var/log/syslog" -lines 20
./dingo -ip server -user root -tail "/var/log/app.log" -follow

# Leave execution trace
./dingo -ip server -user root -footprint -hostname "my-workstation"

# Persistent mode (container-friendly)
./dingo -ip server -user root -cmd "uptime" -persistent -interval 60s
```

### Command-Line Flags
```
-ip string        Target IP address
-host string      SSH host (user@hostname:port format)
-port string      SSH port (default "22")
-user string      Username (default "user") 
-password string  Password authentication
-key string       SSH private key path

-cmd string       Command to execute
-script string    Script file to execute  
-upload string    Upload file (local:remote)
-download string  Download file (remote:local)
-shell            Interactive shell
-stream           Stream command output
-tail string      Tail file path
-follow           Follow file changes (-f)
-lines int        Lines to show (default 10)
-footprint        Upload execution trace
-hostname string  Source hostname for footprint
-restore          Restore a session, requires ip to be set, only supported in shell or script mode
-no_restore       Disables the use of the session restore functionality, automatically enabled if screen is not installed

-persistent       Keep connection alive
-interval duration Interval for persistent mode (default 30s)
```

## API Usage

### Connect
```go
import "github.com/Quok-it/dingo/pkg/dingo"

// Password auth
client, err := dingo.ConnectWithPassword("server:22", "user", "pass")

// SSH key auth  
client, err := dingo.ConnectWithKey("server:22", "user", "/path/to/key")

// SSH key with passphrase
client, err := dingo.ConnectWithKeyAndPassphrase("server:22", "user", "/path/to/key", "phrase")

defer client.Close()
```

### Execute Commands
```go
// Single command
output, err := client.Command("uptime").Output()

// Multiple commands
cmd := client.Command("cd /tmp").Cmd("ls -la").Cmd("pwd")
err := cmd.Run()

// Raw script
script := `#!/bin/bash
echo "Starting backup..."
tar -czf backup.tar.gz /data
echo "Backup complete"`
err := client.Script(script).Run()

// Script file
err := client.ScriptFile("./deploy.sh").Run()
```

### File Operations
```go
fs := client.FileSystem()
defer fs.Close()

// Upload/download
err := fs.Upload("local.txt", "/remote/path.txt")
err := fs.Download("/remote/file.txt", "./local.txt")

// File manipulation
data, err := fs.ReadFile("/remote/config.txt")
err = fs.WriteFile("/remote/config.txt", []byte("data"), 0644)
err = fs.Chmod("/remote/script.sh", 0755)
err = fs.Remove("/remote/temp.txt")

// Directory operations  
err = fs.Mkdir("/remote/newdir")
files, err := fs.ReadDir("/remote/path")
```

### Interactive Shell
```go
// Non-interactive
shell := client.Shell()
err := shell.Start()

// Interactive with PTY
shell := client.InteractiveShell(nil)
err := shell.Start()
```

## Architecture

```
cmd/dingo/          Command-line application
pkg/dingo/          Library code
├── types.go        Interfaces
├── auth.go         Authentication  
├── client.go       SSH client
├── command.go      Command execution
├── shell.go        Interactive shells
├── filesystem.go   SFTP operations
└── options.go      Configuration
```
