FROM node:24-alpine AS builder

WORKDIR /app

# Copy package files first for better layer caching
COPY package*.json ./

# Reproducible install from committed package-lock.json
RUN npm ci

# Copy source code
COPY . .

# Build arguments
ARG VITE_API_URL

# Set environment variables for build
# (the Google Maps key is served at runtime via /api/config, not baked in)
ENV VITE_API_URL=$VITE_API_URL
ENV NODE_ENV=production

# Build the app with optimizations
RUN npm run build

# Production stage - minimal nginx image (pinned)
FROM nginx:1.30-alpine

# Remove default nginx configuration
RUN rm -rf /usr/share/nginx/html/*

# Copy built files with correct permissions
COPY --from=builder --chown=nginx:nginx /app/dist /usr/share/nginx/html

# Copy nginx config with optimizations
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create non-root user (nginx user already exists in nginx:alpine)
RUN touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid && \
    chown -R nginx:nginx /var/cache/nginx

USER nginx

EXPOSE 80

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
