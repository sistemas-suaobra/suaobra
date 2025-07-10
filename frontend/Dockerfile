# Stage 1: Build the static assets
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./

# Use a cache mount for the npm cache directory
RUN --mount=type=cache,target=/root/.npm npm install

COPY . .
RUN npm run build

# Stage 2: Serve the static assets with Caddy
FROM caddy:2-alpine

WORKDIR /app

# Copy the built assets from the builder stage
COPY --from=builder /app/dist .

# Copy the dedicated Caddyfile for the frontend
COPY Caddyfile /etc/caddy/Caddyfile

# Caddy will automatically use the Caddyfile in /etc/caddy
# and listen on port 80.
