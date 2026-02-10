#!/bin/sh

# SaferCloud Frontend - Runtime Environment Variable Injection
# This script generates env.js at container startup with actual environment variables

set -e

echo "Injecting runtime environment variables..."

# Generate env.js from template with actual environment variables
envsubst < /usr/share/nginx/html/env.js.template > /usr/share/nginx/html/env.js

echo "Environment variables injected successfully"
echo "API_URL: ${VITE_API_URL}"
echo "SUPABASE_URL: ${VITE_SUPABASE_URL}"

# Execute the original command (nginx)
exec "$@"
