#!/bin/bash

set -eu

COMPOSE="${COMPOSE:-insolvency-api}"
SEARCH_ROOT=$(pwd)
MAX_DEPTH=5
DEPTH=0
COMPOSE_DIR=""

#Confirm Bash
[ -n "$BASH_VERSION" ] || { echo "Run with bash"; exit 1; }

#Search for docker-chs-development
while [[ $DEPTH -le $MAX_DEPTH ]]; do
  CANDIDATE=$(find "$SEARCH_ROOT" -maxdepth 2 -type d -name "docker-chs-development" 2>/dev/null | head -n 1 || true)
  if [[ -n "$CANDIDATE" ]]; then
    COMPOSE_DIR="$CANDIDATE"
    break
  fi
  SEARCH_ROOT=$(dirname "$SEARCH_ROOT")
  DEPTH=$((DEPTH + 1))
done

#Fallback to search through common and root folders
if [[ -z "$COMPOSE_DIR" ]]; then
  echo "Scanning home directories"
  COMPOSE_DIR=$(find ~/Documents ~/Dev ~/git ~/workspace ~ -maxdepth 5 -type d -name "docker-chs-development" 2>/dev/null | head -n 1)
fi

[ -n "$COMPOSE_DIR" ] || { echo "Could not locate docker-chs-development"; exit 1; }

#Find docker compose file
COMPOSE_FILE=$(find "$COMPOSE_DIR" -name "${COMPOSE}.docker-compose.yaml" | head -n 1 || true)
[ -f "$COMPOSE_FILE" ] || { echo "Compose file not found for $COMPOSE"; exit 1; }

#Extract image name from docker compose file
IMAGE_NAME=$(awk -F'image: ' '/image:/ {print $2}' "$COMPOSE_FILE" | xargs)
[ -n "$IMAGE_NAME" ] || { echo "No image name found in compose"; exit 1; }

#Build image
echo "Building $IMAGE_NAME"
docker build -t "$IMAGE_NAME" ecs-image-build
echo "Build complete: $IMAGE_NAME"

#Optional cleanup
rm -rf ecs-image-build/app
