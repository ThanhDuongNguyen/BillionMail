#!/bin/bash
set -e

# ============================================================
# Build BillionMail Core Docker image with RBAC
# Usage: bash build-core-image.sh [DOCKER_USER] [VERSION]
# Example: bash build-core-image.sh thanhduongnguyen 4.8.1-rbac
# ============================================================

DOCKER_USER=${1:-"duongnt151099"}
VERSION=${2:-"4.8.1-rbac"}
IMAGE_NAME="${DOCKER_USER}/billionmail-core:${VERSION}"

echo "=========================================="
echo "Building BillionMail Core: ${IMAGE_NAME}"
echo "=========================================="

# Step 1: Build frontend
echo ""
echo "[1/4] Building frontend..."
cd core/frontend
npm install --silent
npx rsbuild build
cd ../..

# Step 2: Copy frontend dist to core/public/dist
echo ""
echo "[2/4] Copying frontend build to core/public/dist..."
rm -rf core/public/dist
cp -r core/frontend/dist core/public/dist

# Step 3: Build Go binaries (both amd64 and arm64)
echo ""
echo "[3/4] Building Go binaries..."

# Use Docker to cross-compile
docker run --rm \
  -v "$(pwd)/core":/opt/core \
  -w /opt/core \
  golang:1.22-alpine sh -c "
    apk add --no-cache file
    echo 'Building amd64...'
    GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o billionmail-amd64 main.go
    echo 'Building arm64...'
    GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o billionmail-arm64 main.go
    ls -la billionmail-*
  "

# Step 4: Build Docker image (multi-platform)
echo ""
echo "[4/4] Building Docker image..."

# Build context is Dockerfiles/core, but we need core/ files
# Copy necessary files to build context
BUILD_CTX=$(mktemp -d)
cp -r Dockerfiles/core/* "$BUILD_CTX/"
mkdir -p "$BUILD_CTX/core"
cp core/billionmail-amd64 "$BUILD_CTX/core/"
cp core/billionmail-arm64 "$BUILD_CTX/core/"
cp -r core/manifest "$BUILD_CTX/core/"
cp -r core/languages "$BUILD_CTX/core/"
cp -r core/public "$BUILD_CTX/core/"
cp -r core/resource "$BUILD_CTX/core/"
cp -r core/template "$BUILD_CTX/core/"

# Build and push multi-platform image
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t "${IMAGE_NAME}" \
  -t "${DOCKER_USER}/billionmail-core:latest" \
  --push \
  "$BUILD_CTX"

# Cleanup
rm -rf "$BUILD_CTX"
rm -f core/billionmail-amd64 core/billionmail-arm64

echo ""
echo "=========================================="
echo "✅ Image pushed: ${IMAGE_NAME}"
echo "✅ Image pushed: ${DOCKER_USER}/billionmail-core:latest"
echo "=========================================="
echo ""
echo "To update production, change docker-compose.yml:"
echo "  core-billionmail:"
echo "    image: ${IMAGE_NAME}"
echo ""
echo "Then run: docker compose pull core-billionmail && docker compose up -d core-billionmail"
