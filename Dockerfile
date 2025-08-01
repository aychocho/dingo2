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
    echo '' >> /app/entrypoint.sh && \
    echo '# Validate required parameters' >> /app/entrypoint.sh && \
    echo 'if [ -z "$IP" ]; then' >> /app/entrypoint.sh && \
    echo '    echo "Error: IP address is required"' >> /app/entrypoint.sh && \
    echo '    echo "Usage: docker run -v /path/to/ssh/key:/app/.ssh/your_key_name <image> <ip> <port>"' >> /app/entrypoint.sh && \
    echo '    exit 1' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Set default port if not provided' >> /app/entrypoint.sh && \
    echo 'if [ -z "$PORT" ]; then' >> /app/entrypoint.sh && \
    echo '    PORT="22"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Find any SSH key (non-.pub files) in mounted directory' >> /app/entrypoint.sh && \
    echo 'SSH_KEY=""' >> /app/entrypoint.sh && \
    echo 'for key in /app/.ssh/*; do' >> /app/entrypoint.sh && \
    echo '    if [ -f "$key" ] && [ "${key##*.}" != "pub" ]; then' >> /app/entrypoint.sh && \
    echo '        SSH_KEY="$key"' >> /app/entrypoint.sh && \
    echo '        break' >> /app/entrypoint.sh && \
    echo '    fi' >> /app/entrypoint.sh && \
    echo 'done' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Check if any SSH key was found' >> /app/entrypoint.sh && \
    echo 'if [ -z "$SSH_KEY" ]; then' >> /app/entrypoint.sh && \
    echo '    echo "Error: No SSH key found in /app/.ssh/"' >> /app/entrypoint.sh && \
    echo '    echo "Mount your SSH key with: -v /path/to/ssh/key:/app/.ssh/your_key_name"' >> /app/entrypoint.sh && \
    echo '    echo "Any non-.pub file will be used as the SSH key"' >> /app/entrypoint.sh && \
    echo '    exit 1' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Set proper permissions on SSH key' >> /app/entrypoint.sh && \
    echo 'chmod 600 "$SSH_KEY"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Build dingo command' >> /app/entrypoint.sh && \
    echo 'DINGO_CMD="./dingo -ip $IP -port $PORT -user user -footprint -key $SSH_KEY"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo 'echo "Running: $DINGO_CMD"' >> /app/entrypoint.sh && \
    echo 'exec $DINGO_CMD' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command (can be overridden)
CMD [] 