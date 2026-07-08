# Stage 1: Dev environment
FROM node:20-alpine AS dev
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm npm install
COPY . .
CMD ["npm", "run", "dev", "--", "--host", "--port", "80"]

# Stage 2: Build the static assets
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./

# Use a cache mount for the npm cache directory
RUN --mount=type=cache,target=/root/.npm npm ci

COPY . .

# Evita falha de build no Dokploy quando upload de source maps ao Sentry nao esta disponivel.
ENV DISABLE_SENTRY_UPLOAD=1
RUN npm run build

# Stage 2: Serve the static assets with Caddy
FROM caddy:2-alpine

WORKDIR /app

# Copy the built assets from the builder stage
COPY --from=builder /app/dist .

# Copy the dedicated Caddyfile for the frontend
COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 8091

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8091/ || exit 1

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
