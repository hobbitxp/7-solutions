#!/bin/bash

# Generate Protocol Buffer code for gRPC

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print info message
print_info() {
  echo -e "${YELLOW}ℹ $1${NC}"
}

# Print success message
print_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

# Print error message
print_error() {
  echo -e "${RED}✗ $1${NC}"
}

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
  print_error "protoc is not installed. Please install it to run this script."
  echo "  macOS: brew install protobuf"
  echo "  Linux: apt-get install protobuf-compiler"
  exit 1
fi

# Check if protoc plugins are installed
print_info "Checking and installing required protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# Generate code from user.proto
print_info "Generating code from user.proto..."
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/user.proto

# Check if files were generated
if [ -f api/proto/user.pb.go ] && [ -f api/proto/user_grpc.pb.go ]; then
  print_success "Code generation successful!"
else
  print_error "Failed to generate code. Check for errors above."
  exit 1
fi

print_info "Generated files:"
ls -la api/proto/*.pb.go