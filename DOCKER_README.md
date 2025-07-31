# Dingo Docker Setup

This Docker setup allows you to run the dingo SSH client in a containerized environment with runtime arguments.

## Quick Start

### Build the container once:
```bash
./build.sh build
```

### Run with your parameters:
```bash
# Your exact example
./build.sh run 147.185.41.233 20046 /path/to/ssh/key

# Or just with IP (uses default port 22)
./build.sh run 147.185.41.233

# Or with IP and port
./build.sh run 147.185.41.233 20046
```

### Or use Docker directly:
```bash
# Build
docker build -t dingo-ssh .

# Run with arguments
docker run --rm dingo-ssh 147.185.41.233 20046 /path/to/ssh/key

# Mount SSH key as volume
docker run --rm \
  -v /path/to/ssh/key:/path/to/ssh/key \
  dingo-ssh 147.185.41.233 20046 /path/to/ssh/key
```

## Parameters

The Docker container accepts the following runtime arguments:

1. **IP Address** (required): Target IP address to connect to
2. **Port** (optional, default: 22): SSH port number  
3. **SSH Key Path** (optional): Path to SSH private key file

## Usage Examples

### Your exact example:
```bash
./build.sh run 147.185.41.233 20046 /path/to/your/ssh/key
```

This will run:
```bash
./dingo -ip 147.185.41.233 -port 20046 -user user -footprint -key /path/to/your/ssh/key
```

### Without SSH key (password auth):
```bash
./build.sh run 147.185.41.233 20046
```

### Default port (22):
```bash
./build.sh run 147.185.41.233
```

## Build Script Commands

```bash
# Build the image (do this once)
./build.sh build

# Run with parameters
./build.sh run <ip> [port] [ssh_key_path]

# Show help
./build.sh
```

## Docker Image Details

- **Base Image**: Alpine Linux (minimal footprint)
- **Size**: ~10MB base + dingo binary
- **SSH Client**: Includes openssh-client for SSH operations
- **Working Directory**: `/app`
- **Entrypoint**: Script that takes IP, port, and SSH key as arguments

## Security Considerations

1. **SSH Keys**: When mounting SSH keys, ensure proper file permissions (600)
2. **Network Access**: Container needs network access to target SSH servers
3. **Key Storage**: Consider using Docker secrets or Kubernetes secrets for production

## Troubleshooting

### SSH Key Issues
```bash
# Check if key file exists and has correct permissions
ls -la /path/to/ssh/key
chmod 600 /path/to/ssh/key
```

### Network Connectivity
```bash
# Test connectivity from host
ssh -p 20046 user@147.185.41.233
```

### Container Debugging
```bash
# Run container interactively
docker run -it --rm dingo-ssh /bin/sh

# Check dingo binary
docker run --rm dingo-ssh ls -la /app/dingo
```

## Customization

### Modify the command
Edit the `Dockerfile` and change the `DINGO_CMD` line in the entrypoint script:

```dockerfile
# Change this line in the entrypoint script
echo 'DINGO_CMD="./dingo -ip $IP -port $PORT -user user -footprint"' >> /app/entrypoint.sh
```

### Add additional parameters
You can extend the entrypoint script to accept more parameters by adding more command-line arguments. 