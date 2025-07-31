#!/bin/bash

# Simple build script for dingo Docker container
set -e

IMAGE_NAME="dingo-ssh"

# Function to show usage
show_usage() {
    echo "Usage:"
    echo "  $0 build                    # Build the Docker image"
    echo "  $0 run <ip> [port] [key]    # Run the container with arguments"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 run 147.185.41.233 20046 /path/to/ssh/key"
    echo "  $0 run 147.185.41.233 20046"
    echo "  $0 run 147.185.41.233"
    echo ""
    echo "Or use docker directly:"
    echo "  docker run --rm dingo-ssh 147.185.41.233 20046 /path/to/ssh/key"
}

# Check if we have any arguments
if [ $# -eq 0 ]; then
    show_usage
    exit 1
fi

case "$1" in
    build)
        echo "Building Docker image..."
        docker build -t "$IMAGE_NAME" .
        echo "Build complete. Image: $IMAGE_NAME"
        ;;
    run)
        shift  # Remove 'run' from arguments
        
        # Check if we have the required arguments
        if [ $# -lt 1 ]; then
            echo "Error: IP address is required"
            show_usage
            exit 1
        fi

        IP="$1"
        PORT="${2:-22}"
        SSH_KEY="${3:-}"

        echo "Running dingo container..."
        echo "IP: $IP"
        echo "Port: $PORT"
        if [ -n "$SSH_KEY" ]; then
            echo "SSH Key: $SSH_KEY"
        fi

        # Build the docker run command
        DOCKER_CMD="docker run --rm"

        # Add SSH key volume mount if provided
        if [ -n "$SSH_KEY" ] && [ -f "$SSH_KEY" ]; then
            DOCKER_CMD="$DOCKER_CMD -v $(realpath "$SSH_KEY"):$SSH_KEY"
        fi

        # Add the container run
        DOCKER_CMD="$DOCKER_CMD $IMAGE_NAME $IP $PORT $SSH_KEY"

        echo "Executing: $DOCKER_CMD"
        eval $DOCKER_CMD
        ;;
    *)
        echo "Unknown command: $1"
        show_usage
        exit 1
        ;;
esac 