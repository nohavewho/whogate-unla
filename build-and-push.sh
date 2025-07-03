#!/bin/bash

# Build and push custom UNLA image with English as default

# Registry and image name
REGISTRY="docker.io"
IMAGE_NAME="whogate/unla-english"
TAG="latest"

# Build the image
echo "Building custom UNLA image with English as default..."
cd /Users/vladislaviushchenko/civo/unla

docker build -f deploy/docker/allinone/Dockerfile -t ${IMAGE_NAME}:${TAG} .

# Tag for registry
docker tag ${IMAGE_NAME}:${TAG} ${REGISTRY}/${IMAGE_NAME}:${TAG}

echo "Build complete. Image: ${REGISTRY}/${IMAGE_NAME}:${TAG}"

# Note: To push to registry, you need to be logged in:
# docker login
# docker push ${REGISTRY}/${IMAGE_NAME}:${TAG}