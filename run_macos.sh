#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "Starting Provisioning System Build & Dev process..."

# 1. Build Frontend Production (to be embedded in Go binary)
echo "--- Step 1: Building Frontend Production Bundle ---"
cd frontend
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi
echo "Running npm run build..."
npm run build
cd ..

# 2. Prepare Static Files for Embedding
echo "--- Step 2: Syncing Static Files for Go Embed ---"
mkdir -p backend/cmd/server/static
# Clear previous static files to ensure a clean embed
rm -rf backend/cmd/server/static/*
# Copy production build to backend static folder
cp -R frontend/build/* backend/cmd/server/static/

# 3. Build Backend using Docker (Cross-compilation for macOS Intel)
echo "--- Step 3: Building Backend via Docker (target: darwin/amd64) ---"
docker run --rm \
    -v "$(pwd)/backend":/app \
    -w /app \
    -e GOOS=darwin \
    -e GOARCH=amd64 \
    -e CGO_ENABLED=0 \
    golang:1.23.4 \
    sh -c "go mod tidy && go build -v -o provisioning-system-macos cmd/server/main.go"

echo "Backend binary built with embedded frontend."

# 4. Prepare for running
mv backend/provisioning-system-macos ./provisioning-system
chmod +x provisioning-system

# 5. Start Backend in background
echo "--- Step 4: Starting Backend ---"
# Using -config-dir conf to find yaml and vendors
./provisioning-system -config-dir conf &
BACKEND_PID=$!

# Function to kill background processes on exit
cleanup() {
    echo ""
    echo "Shutting down backend (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null || true
    exit
}
trap cleanup SIGINT SIGTERM

# 6. Start Frontend Dev Server
echo "--- Step 5: Starting Frontend Dev Server (Vite) ---"
echo "The application is available at:"
echo "  - Embedded UI (Static): http://localhost:8090"
echo "  - Dev UI (Hot-reload):  http://localhost:5173"
echo "------------------------------------------------"

cd frontend
# Use npx to ensure vite is found even if not in global PATH
npx vite dev
