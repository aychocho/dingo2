#!/usr/bin/env bash
set -euo pipefail

# Read index from environment variable, default to 0
INDEX="${JOB_COMPLETION_INDEX:-0}"

# Extract IP and PORT from mounted files using sed (convert 0-based to 1-based for sed)
IP=$(sed -n "$((INDEX + 1))p" /config/ip-list.txt)
PORT=$(sed -n "$((INDEX + 1))p" /config/port-list.txt)

# Validate extracted values
if [[ -z "$IP" ]] || [[ -z "$PORT" ]]; then
    echo "Error: Failed to extract IP or PORT at index $INDEX" >&2
    exit 1
fi

# Find SSH key in mounted secrets directory
SSH_KEY=""
for key in /secrets/*; do
    if [[ -f "$key" ]] && [[ "${key##*.}" != "pub" ]]; then
        SSH_KEY="$key"
        break
    fi
done

# Check if SSH key was found
if [[ -z "$SSH_KEY" ]]; then
    echo "Error: No SSH key found in /secrets/" >&2
    exit 1
fi

# Copy SSH key to writable location and set proper permissions
cp "$SSH_KEY" /tmp/ssh_key
chmod 600 /tmp/ssh_key

# Execute dingo with extracted IP and PORT
exec /app/dingo -ip "$IP" -port "$PORT" -user user -footprint -key /tmp/ssh_key "$@" 