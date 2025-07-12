# Use the official PocketBase image as the base
FROM ghcr.io/pocketbase/pocketbase:latest AS base

# Copy your custom application files into the image
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./suaobra-app.go ./
COPY ./server/ ./server
COPY ./store/ ./store
COPY ./templates/ ./templates

# The official image already has an entrypoint that will run the application.
# It will automatically use the files we copied.
