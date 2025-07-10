# Stage 1: Build the Go application
FROM golang:1.20-alpine AS builder

# Install build tools for CGO
RUN apk add --no-cache build-base

WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy only the necessary source code for the backend
COPY suaobra-app.go .
COPY server/ ./server
COPY store/ ./store
COPY templates/ ./templates

# Set CGO flags to fix sqlite3 build issues on Alpine
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

# Build the application with CGO enabled, using a cache mount for build artifacts
# and flags to create a smaller binary.
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 go build -ldflags="-s -w" -a -o suaobra-app .

# Stage 2: Create the final, minimal image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/suaobra-app .

# The core.db is now managed by the volume, so it's not needed here.

EXPOSE 8090

# Run the application
CMD ["./suaobra-app", "serve", "--http", "0.0.0.0:8090"]
