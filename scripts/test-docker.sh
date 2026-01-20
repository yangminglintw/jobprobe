#!/bin/bash
# test-docker.sh - Run tests using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== Building Docker images ==="
docker compose build

echo ""
echo "=== Starting mock API ==="
docker compose up -d mock-api

echo ""
echo "=== Waiting for services to be healthy ==="
sleep 5

echo ""
echo "=== Running jprobe tests ==="
docker compose run --rm jprobe run --config /etc/jprobe --verbose

echo ""
echo "=== Stopping services ==="
docker compose down

echo ""
echo "=== Docker tests completed! ==="
