# Use a compatible base image
FROM debian:bookworm-slim

# Install necessary packages for SSH operations and runtime libraries
RUN apt-get update && apt-get install -y \
    openssh-client \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy the dingo binary
COPY dingo /app/dingo

# Make the binary executable
RUN chmod +x /app/dingo

# Create directory for SSH keys
RUN mkdir -p /app/.ssh

# Create entrypoint script
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'set -e' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Get arguments' >> /app/entrypoint.sh && \
    echo 'IP="$1"' >> /app/entrypoint.sh && \
    echo 'PORT="$2"' >> /app/entrypoint.sh && \
    echo 'SSH_KEY_PATH="$3"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Validate required parameters' >> /app/entrypoint.sh && \
    echo 'if [ -z "$IP" ]; then' >> /app/entrypoint.sh && \
    echo '    echo "Error: IP address is required"' >> /app/entrypoint.sh && \
    echo '    echo "Usage: docker run <image> <ip> <port> <ssh_key_path>"' >> /app/entrypoint.sh && \
    echo '    exit 1' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Set default port if not provided' >> /app/entrypoint.sh && \
    echo 'if [ -z "$PORT" ]; then' >> /app/entrypoint.sh && \
    echo '    PORT="22"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Handle SSH key if provided' >> /app/entrypoint.sh && \
    echo 'if [ -n "$SSH_KEY_PATH" ]; then' >> /app/entrypoint.sh && \
    echo '    if [ -f "$SSH_KEY_PATH" ]; then' >> /app/entrypoint.sh && \
    echo '        # Copy key to container and set permissions' >> /app/entrypoint.sh && \
    echo '        cp "$SSH_KEY_PATH" /app/.ssh/id_rsa' >> /app/entrypoint.sh && \
    echo '        chmod 600 /app/.ssh/id_rsa' >> /app/entrypoint.sh && \
    echo '        SSH_KEY="/app/.ssh/id_rsa"' >> /app/entrypoint.sh && \
    echo '    else' >> /app/entrypoint.sh && \
    echo '        echo "Warning: SSH key file $SSH_KEY_PATH not found"' >> /app/entrypoint.sh && \
    echo '        SSH_KEY=""' >> /app/entrypoint.sh && \
    echo '    fi' >> /app/entrypoint.sh && \
    echo 'else' >> /app/entrypoint.sh && \
    echo '    SSH_KEY=""' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Build dingo command' >> /app/entrypoint.sh && \
    echo 'DINGO_CMD="./dingo -ip $IP -port $PORT -user user -footprint"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Add SSH key if provided' >> /app/entrypoint.sh && \
    echo 'if [ -n "$SSH_KEY" ]; then' >> /app/entrypoint.sh && \
    echo '    DINGO_CMD="$DINGO_CMD -key $SSH_KEY"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo 'echo "Running: $DINGO_CMD"' >> /app/entrypoint.sh && \
    echo 'exec $DINGO_CMD' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command (can be overridden)
CMD [] 