#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Rebuilding Docker images from $PROJECT_ROOT..."

# Build Backend
echo "Building Backend..."
docker build -t coinwave-backend:latest "$PROJECT_ROOT/backend"

# Build Frontend
echo "Building Frontend..."
docker build -t coinwave-frontend:latest "$PROJECT_ROOT/frontend"

echo "✅ Build complete."

# Check if running on K8s
if kubectl get deployments backend &> /dev/null; then
    echo "🔄 Restarting K8s deployments..."
    kubectl rollout restart deployment backend
    kubectl rollout restart deployment frontend
    echo "✅ K8s deployments restarted."
else
    echo "ℹ️  If you are using Docker Compose, run: docker-compose up -d --build"
fi
