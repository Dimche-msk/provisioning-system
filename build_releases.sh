#!/bin/bash
set -e

# Create releases directory
mkdir -p release

echo "Starting build process..."

# 0. Build Frontend
echo "Building Frontend..."
cd frontend
pnpm install
pnpm build
cd ..

echo "Updating Backend static files..."
rm -rf backend/cmd/server/static/*
cp -r frontend/build/* backend/cmd/server/static/

# macOS (Darwin)
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Building for macOS Intel (AMD64)..."
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -C backend -o ../release/provisioning-system-darwin-amd64 cmd/server/main.go
    
    echo "Building for macOS Apple Silicon (ARM64)..."
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -C backend -o ../release/provisioning-system-darwin-arm64 cmd/server/main.go
else
    echo "Skipping macOS build (not running on macOS)."
fi

echo "Starting build process using Docker (Alpine) for other platforms..."

# Linux AMD64
echo "Building for Linux AMD64 (Static)..."
# Added CGO_CFLAGS="-D_LARGEFILE64_SOURCE" to fix off64_t/pread64 errors in go-sqlite3 on Alpine
docker pull --platform linux/amd64 golang:1.23-alpine
docker run --rm --platform linux/amd64 -v "$PWD":/app -w /app/backend -e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=amd64 -e CGO_CFLAGS="-D_LARGEFILE64_SOURCE" golang:1.23-alpine sh -c "uname -m && apk add --no-cache build-base && go build -ldflags '-linkmode external -extldflags \"-static\"' -o ../release/provisioning-system-linux-amd64 cmd/server/main.go"

# Windows AMD64
echo "Building for Windows AMD64..."
docker run --rm --platform linux/amd64 -v "$PWD":/app -w /app/backend -e CGO_ENABLED=1 -e GOOS=windows -e GOARCH=amd64 golang:1.23-alpine sh -c "apk add --no-cache mingw-w64-gcc && CC=x86_64-w64-mingw32-gcc go build -o ../release/provisioning-system-windows-amd64.exe cmd/server/main.go"

# Linux 386 - SKIPPED due to persistent Alpine compatibility issues
echo "Skipping Linux 386 build due to Alpine compatibility issues."
# echo "Building for Linux 386 (Static)..."
# docker pull --platform linux/386 golang:1.23-alpine
# docker run --rm --platform linux/386 -v "$PWD":/app -w /app/backend -e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=386 -e CGO_CFLAGS="-D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64" golang:1.23-alpine sh -c "uname -m && apk add --no-cache build-base && go build -ldflags '-linkmode external -extldflags \"-static\"' -o ../releases/provisioning-system-linux-386 cmd/server/main.go"

echo "Build complete. Binaries are located in release/"
ls -lh release/
