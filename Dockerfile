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

# Copy the entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# Default command (can be overridden)
CMD [] 