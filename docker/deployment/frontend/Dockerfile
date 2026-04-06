# =================================
# Kagibi Frontend - Multi-Stage Dockerfile
# Vue.js 3 + Vite built with nginx
# =================================

# Stage 1: Build
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --silent

# Copy source code
COPY . .

# Build arguments for environment variables
ARG VITE_API_URL
ARG VITE_SUPABASE_URL
ARG VITE_SUPABASE_KEY

# Set build-time environment variables
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_SUPABASE_URL=$VITE_SUPABASE_URL
ENV VITE_SUPABASE_KEY=$VITE_SUPABASE_KEY

# Build the application
RUN npm run build

# Stage 2: Runtime with nginx
FROM nginx:1.27-alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Create env.js template for runtime environment variables
RUN echo 'window.ENV = { \
    API_URL: "${VITE_API_URL}", \
    SUPABASE_URL: "${VITE_SUPABASE_URL}", \
    SUPABASE_KEY: "${VITE_SUPABASE_KEY}" \
}' > /usr/share/nginx/html/env.js.template

# Copy entrypoint script for runtime env injection
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Use custom entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
