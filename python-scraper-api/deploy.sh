#!/bin/bash

# Script para build e deploy da API RRT Scraper

set -e

echo "🐳 Building RRT Scraper API Docker image..."

# Build da imagem
docker build -t rrt-scraper-api:latest .

echo "✅ Build completed successfully!"

echo "🚀 Starting container..."

# Para o container se estiver rodando
docker-compose down

# Inicia o container
docker-compose up -d

echo "✅ Container started successfully!"

# Aguarda o health check
echo "🔍 Waiting for health check..."
sleep 10

# Verifica se está funcionando
if curl -f http://localhost:5000/health > /dev/null 2>&1; then
    echo "✅ API is healthy and running at http://localhost:5000"
    echo "📚 API Documentation:"
    echo "  - Health check: GET http://localhost:5000/health"
    echo "  - Get RRT data: GET http://localhost:5000/rrt/{numero}"
else
    echo "❌ API health check failed"
    echo "📋 Container logs:"
    docker-compose logs
    exit 1
fi
