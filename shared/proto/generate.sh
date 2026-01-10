#!/bin/bash

# Script to generate Go code from .proto files

set -e

echo "🔧 Generating Protocol Buffers code..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc is not installed"
    echo "Installation: brew install protobuf (macOS)"
    exit 1
fi

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Check if Go plugin is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "⚠️  protoc-gen-go is not installed"
    echo "Installation: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "⚠️  protoc-gen-go-grpc is not installed"
    echo "Installation: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Generate Go code
export PATH="$PATH:$(go env GOPATH)/bin"

# Generate device proto
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  device/device.proto

# Generate user proto
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  user/user.proto

echo "✅ Code generated successfully in shared/proto/"
