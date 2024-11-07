#!/bin/bash
echo "Creating build for Windows, Linux and MacOS"
BUILD_VERSION="v0.0.1-alpha"
echo "Alpha Version Release Building binary with new version $BUILD_VERSION"
# Build for Linux 32-bit
GOOS=linux GOARCH=386 go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-linux-386 ./../custom-mongo-exporter.go
echo "Built for Linux 32-bit"

# Build for Linux 64-bit
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-linux-amd64 ./../custom-mongo-exporter.go
echo "Built for Linux 64-bit"

# Build for macOS 32-bit
GOOS=darwin GOARCH=386 go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-macos-386 ./../custom-mongo-exporter.go
echo "Built for macOS 32-bit"

# Build for macOS 64-bit
go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-macos-amd64 custom-mongo-exporter.go
echo "Built for macOS 64-bit"

# Build for Windows 32-bit
GOOS=windows GOARCH=386 go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-windows-386.exe ./../custom-mongo-exporter.go
echo "Built for Windows 32-bit"

# Build for Windows 64-bit
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=${BUILD_VERSION}" -o ./../bin/custom-1203-mongo-exporter-windows-amd64.exe ./../custom-mongo-exporter.go
echo "Built for Windows 64-bit"
echo "Build completed in bin/ directory"