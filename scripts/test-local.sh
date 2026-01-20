#!/bin/bash
# test-local.sh - Run local tests with mock API

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

MOCK_API_PID=""

cleanup() {
    if [ -n "$MOCK_API_PID" ]; then
        echo "Stopping mock API..."
        kill "$MOCK_API_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT

echo "=== Building jprobe ==="
cd "$PROJECT_ROOT"
make build

echo ""
echo "=== Building mock API ==="
go build -o "$PROJECT_ROOT/test/mock-api/mock-api" "$PROJECT_ROOT/test/mock-api/main.go"

echo ""
echo "=== Starting mock API ==="
"$PROJECT_ROOT/test/mock-api/mock-api" &
MOCK_API_PID=$!
sleep 2

echo ""
echo "=== Running unit tests ==="
go test ./... -v

echo ""
echo "=== Running jprobe health checks ==="
./jprobe run --config configs/ --tags api --verbose

echo ""
echo "=== All tests passed! ==="
