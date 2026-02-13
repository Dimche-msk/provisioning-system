#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "Starting Provisioning System in Development Mode..."

# 1. Build Backend using Docker (Cross-compilation for macOS Intel)
echo "Building Backend via Docker (target: darwin/amd64)..."
# We mount the backend directory and build the binary inside.
# Since macOS 10 is Intel-based, we use GOARCH=amd64.
docker run --rm \
    -v "$(pwd)/backend":/app \
    -w /app \
    -e GOOS=darwin \
    -e GOARCH=amd64 \
    golang:1.23.4 \
    go build -o provisioning-system-macos cmd/server/main.go

echo "Backend binary built: backend/provisioning-system-macos"

# 2. Prepare for running
mv backend/provisioning-system-macos ./provisioning-system
chmod +x provisioning-system

# 3. Start Backend in background
echo "Starting Backend on port 8090..."
./provisioning-system &
BACKEND_PID=$!

# Function to kill background processes on exit
cleanup() {
    echo "Shutting down..."
    kill $BACKEND_PID
    exit
}
trap cleanup SIGINT SIGTERM

# 4. Start Frontend Dev Server
echo "Starting Frontend Dev Server (npm run dev)..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi

# Run npm dev (this will stay in foreground)
npm run dev
