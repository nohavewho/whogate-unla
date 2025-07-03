#!/bin/bash

# Build and deploy updated WhoGate UNLA with Russian language

set -e

echo "Building Docker image with Russian language support..."

# Build the Docker image for x86_64 architecture
docker build --platform linux/amd64 -t nononohave/whogate-unla:latest .

echo "Pushing to Docker Hub..."
docker push nononohave/whogate-unla:latest

echo "Updating Kubernetes deployment..."
kubectl rollout restart deployment/octobrowser-mcp -n octobrowser-mcp

echo "Watching rollout status..."
kubectl rollout status deployment/octobrowser-mcp -n octobrowser-mcp

echo "Deployment complete!"
echo "Access WhoGate at: http://74.220.18.25"
echo "Russian language support is now available!"