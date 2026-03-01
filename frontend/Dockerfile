# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy nginx configuration files
COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf

# Copy built app from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Create dynamic config script from environment variables
RUN echo '#!/bin/sh' > /usr/local/bin/generate-config.sh && \
    echo 'cat > /usr/share/nginx/html/config.js << EOF' >> /usr/local/bin/generate-config.sh && \
    echo 'window.__APP_CONFIG__ = {' >> /usr/local/bin/generate-config.sh && \
    echo '  apiUrl: "${VITE_API_URL}" || "http://localhost:8080/api/v1",' >> /usr/local/bin/generate-config.sh && \
  echo '  supabaseUrl: "${VITE_SUPABASE_URL:-$SUPABASE_URL}" || "",' >> /usr/local/bin/generate-config.sh && \
  echo '  supabaseAnonKey: "${VITE_SUPABASE_ANON_KEY:-$SUPABASE_ANON_KEY}" || "",' >> /usr/local/bin/generate-config.sh && \
    echo '}' >> /usr/local/bin/generate-config.sh && \
    echo 'EOF' >> /usr/local/bin/generate-config.sh && \
    chmod +x /usr/local/bin/generate-config.sh

# Create entrypoint script that generates config and starts nginx
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo 'echo "Génération de la configuration dynamique..."' >> /entrypoint.sh && \
    echo 'echo "VITE_API_URL=${VITE_API_URL}"' >> /entrypoint.sh && \
    echo 'echo "VITE_SUPABASE_URL=${VITE_SUPABASE_URL}"' >> /entrypoint.sh && \
    echo '/usr/local/bin/generate-config.sh' >> /entrypoint.sh && \
    echo 'echo "Configuration générée avec succès"' >> /entrypoint.sh && \
    echo 'exec "$@"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

# Run with entrypoint
ENTRYPOINT ["/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
