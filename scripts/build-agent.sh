#!/bin/bash

# Build script for VPS Monitor Agent

set -e

echo "Building VPS Monitor Agent..."

# Build for Linux (default)
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o bin/agent-linux-amd64 cmd/agent/main.go

# Build for other common platforms
echo "Building for other platforms..."
GOOS=linux GOARCH=arm64 go build -o bin/agent-linux-arm64 cmd/agent/main.go
GOOS=windows GOARCH=amd64 go build -o bin/agent-windows-amd64.exe cmd/agent/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/agent-darwin-amd64 cmd/agent/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/agent-darwin-arm64 cmd/agent/main.go

echo "Build completed successfully!"
echo "Binaries are located in the bin/ directory"