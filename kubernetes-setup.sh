#!/bin/bash

set -e
set -o pipefail

ARCH=${ARCH:-amd64}
echo "Using architecture: $ARCH"

detect_host_arch() {
  local uname_arch
  uname_arch="$(uname -m)"
  case "$uname_arch" in
  x86_64|amd64) echo "amd64" ;; 
  aarch64|arm64) echo "arm64" ;;
  *)
  echo "Unsupported host architecture: $uname_arch" >&2
  exit 1
  ;;
  esac
}

install_minikube() {
  local host_arch bin_name url
  host_arch="$(detect_host_arch)"
  bin_name="minikube-linux-${host_arch}"
  url="https://storage.googleapis.com/minikube/releases/latest/${bin_name}"
  echo "Minikube not found. Installing for host architecture: ${host_arch}."
  curl -LO "${url}"
  sudo install "${bin_name}" /usr/local/bin/minikube
  rm "${bin_name}"
  echo "Minikube installed."
}

echo "Looking for minikube installation."
if ! command -v minikube >/dev/null 2>&1; then
  install_minikube
else
  echo "Minikube is already installed."
fi

echo "Checking Minikube."
if ! minikube status >/dev/null 2>&1; then
  echo "Starting Minikube."
  minikube start
else
  echo "Minikube is already started."
fi

echo "Enabling required Minikube addons..."
minikube addons enable ingress
minikube addons enable metrics-server
sleep 30

eval "$(minikube docker-env)"

function image_exists() {
  docker image inspect "$1" > /dev/null 2>&1
}


FRONTEND_IMAGE="frontend:${ARCH}-local"
BACKEND_IMAGE="backend:${ARCH}-local"
POSTGRES_IMAGE="postgres:17.5"

if image_exists "$FRONTEND_IMAGE"; then
  echo "Image '$FRONTEND_IMAGE' already exists."
else
  echo "Building frontend image: $FRONTEND_IMAGE"
  docker build \
    --platform "linux/${ARCH}" \
    --build-arg NEXT_PUBLIC_BACKEND_URL=http://backend.local \
    -t "$FRONTEND_IMAGE" \
    -t "frontend:local" \
    ./frontend
fi

if image_exists "$BACKEND_IMAGE"; then
  echo "Image '$BACKEND_IMAGE' already exists."
else
  echo "Building backend image: $BACKEND_IMAGE"
  docker build \
    --platform "linux/${ARCH}" \
    -t "$BACKEND_IMAGE" \
    -t "backend:local" \
    ./backend
fi

if ! image_exists "$POSTGRES_IMAGE"; then
  echo "Pulling postgres image: $POSTGRES_IMAGE"
  docker pull "$POSTGRES_IMAGE"
else
  echo "Postgres image already there."
fi


kubectl create configmap db-init-scripts \
  --from-file=init.sql=./db/init.sql \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f kubernetes/backend
kubectl apply -f kubernetes/db
kubectl apply -f kubernetes/frontend

kubectl apply -f kubernetes/frontend/ingress.yaml
kubectl apply -f kubernetes/backend/ingress.yaml

kubectl autoscale deployment backend --cpu-percent=50 --min=2 --max=5
kubectl autoscale deployment frontend --cpu-percent=50 --min=1 --max=3

echo " Setup complete."

