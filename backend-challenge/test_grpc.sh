#!/bin/bash

# Test gRPC Endpoints

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Define gRPC URL
GRPC_URL="localhost:50051"
REST_URL="http://localhost:8080"

# Print section header
print_header() {
  echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Print success message
print_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

# Print error message
print_error() {
  echo -e "${RED}✗ $1${NC}"
}

# Print info message
print_info() {
  echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
  print_error "grpcurl is not installed. Please install it to run this script."
  echo "  macOS: brew install grpcurl"
  echo "  Linux: https://github.com/fullstorydev/grpcurl/releases"
  exit 1
fi

# Check if jq is installed (for token generation)
if ! command -v jq &> /dev/null; then
  print_error "jq is not installed. It's needed for token extraction."
  echo "  macOS: brew install jq"
  echo "  Linux: apt-get install jq"
  exit 1
fi

print_header "Starting gRPC Tests"

# 1. Get authentication token via REST API (needed for authenticated gRPC calls)
print_header "1. Getting authentication token"
LOGIN_RESPONSE=$(curl -s -X POST \
  "${REST_URL}/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

# Extract token from login response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')

if [[ -z "$TOKEN" ]]; then
  print_error "Failed to get authentication token. Some tests will fail."
  
  # Try registering a new user
  print_info "Trying to register a new user..."
  EMAIL="test_$(date +%s)@example.com"
  
  REGISTER_RESPONSE=$(curl -s -X POST \
    "${REST_URL}/api/auth/register" \
    -H 'Content-Type: application/json' \
    -d "{
      \"name\": \"Test User\",
      \"email\": \"$EMAIL\",
      \"password\": \"password123\"
    }")
  
  TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token // empty')
  
  if [[ -z "$TOKEN" ]]; then
    print_error "Could not get a token. Some tests will fail."
  else
    print_success "Got token from new user registration"
  fi
else
  print_success "Authentication successful"
  print_info "Token: ${TOKEN}"
fi

# 2. List all available gRPC services
print_header "2. Listing available gRPC services"
grpcurl -plaintext $GRPC_URL list

# 3. List methods for UserService
print_header "3. Listing methods for UserService"
grpcurl -plaintext $GRPC_URL list backend.UserService

# 4. Test GetUser method (with mock ID)
print_header "4. Testing GetUser method"
if [[ ! -z "$TOKEN" ]]; then
  grpcurl -plaintext -H "Authorization: Bearer ${TOKEN}" \
    -d '{"id": "some-user-id"}' \
    $GRPC_URL backend.UserService/GetUser
else
  print_error "Cannot test GetUser - no token available"
fi

# 5. Test CreateUser method
print_header "5. Testing CreateUser method"
if [[ ! -z "$TOKEN" ]]; then
  NEW_EMAIL="create_user_$(date +%s)@example.com"
  grpcurl -plaintext -H "Authorization: Bearer ${TOKEN}" \
    -d "{
      \"name\": \"Created via gRPC\",
      \"email\": \"$NEW_EMAIL\",
      \"password\": \"grpc_password\"
    }" \
    $GRPC_URL backend.UserService/CreateUser
else
  print_error "Cannot test CreateUser - no token available"
fi

# 6. Test ListUsers method
print_header "6. Testing ListUsers method"
if [[ ! -z "$TOKEN" ]]; then
  grpcurl -plaintext -H "Authorization: Bearer ${TOKEN}" \
    -d '{"page": 1, "page_size": 10}' \
    $GRPC_URL backend.UserService/ListUsers
else
  print_error "Cannot test ListUsers - no token available"
fi

print_header "gRPC Tests Completed"