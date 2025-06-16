#!/bin/bash

set -e
set -o pipefail

echo "🧹 Cleaning old Docker images in Minikube..."

# Use Minikube's Docker daemon
eval $(minikube docker-env)

# Remove old images
docker rmi -f frontend:local || true
docker rmi -f backend:local || true

echo "🔁 Rebuilding Docker images..."

docker build \
  --build-arg NEXT_PUBLIC_BACKEND_URL=http://backend.local \
  -t frontend:local ./frontend

docker build -t backend:local ./backend

echo "Restarting Deployments to use fresh images..."

kubectl delete pods -l app=frontend
kubectl delete pods -l app=backend

echo "Restarting Kubernetes manifests."

kubectl apply -f kubernetes/backend/
kubectl apply -f kubernetes/db/
kubectl apply -f kubernetes/frontend/

echo "Completed."

