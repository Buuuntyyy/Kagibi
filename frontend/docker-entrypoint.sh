#!/bin/sh

# Kagibi Frontend - Nginx Startup Script
# Simple entrypoint that passes arguments to nginx

set -e

# Log startup
echo "Starting Kagibi Frontend..."
echo "VITE_API_URL=${VITE_API_URL}"
echo "VITE_SUPABASE_URL=${VITE_SUPABASE_URL}"

# Execute the command passed to the container (typically nginx)
exec "$@"
