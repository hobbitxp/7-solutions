#!/bin/bash

# Test All Endpoints (REST API and gRPC)

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Define base URLs
REST_URL="http://localhost:8080"
GRPC_URL="localhost:50051"

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

# Check if jq is installed
if ! command -v jq &> /dev/null; then
  print_error "jq is not installed. Please install it to run this script."
  echo "  macOS: brew install jq"
  echo "  Linux: apt-get install jq"
  exit 1
fi

# Check if grpcurl is installed for gRPC tests
if ! command -v grpcurl &> /dev/null; then
  print_info "grpcurl is not installed. gRPC tests will be skipped."
  print_info "To install grpcurl:"
  echo "  macOS: brew install grpcurl"
  echo "  Linux: https://github.com/fullstorydev/grpcurl/releases"
  SKIP_GRPC=true
else
  SKIP_GRPC=false
fi

print_header "Starting API Tests"

# 1. Test Health Check
print_header "1. Testing Health Check"
HEALTH_RESPONSE=$(curl -s -X GET "${REST_URL}/api/health")
echo "$HEALTH_RESPONSE" | jq .
if [[ $(echo "$HEALTH_RESPONSE" | jq -r .status) == "ok" ]]; then
  print_success "Health check successful"
else
  print_error "Health check failed"
fi

# 2. Test User Registration
print_header "2. Testing User Registration"
EMAIL="test_$(date +%s)@example.com"
print_info "Registering user with email: $EMAIL"

REGISTER_RESPONSE=$(curl -s -X POST \
  "${REST_URL}/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{
    \"name\": \"Test User\",
    \"email\": \"$EMAIL\",
    \"password\": \"password123\"
  }")
echo "$REGISTER_RESPONSE" | jq .

# Try to extract token, but it might not be there if registration failed
REG_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token // empty')

# 3. Test User Login
print_header "3. Testing User Login"
LOGIN_RESPONSE=$(curl -s -X POST \
  "${REST_URL}/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"password123\"
  }")
echo "$LOGIN_RESPONSE" | jq .

# Extract token from login response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')

# If we didn't get a token from login (maybe it failed), try using the registration token
if [[ -z "$TOKEN" && ! -z "$REG_TOKEN" ]]; then
  TOKEN=$REG_TOKEN
  print_info "Using token from registration"
fi

if [[ -z "$TOKEN" ]]; then
  print_error "Failed to get authentication token. Using a previously saved token if available."
  # Try to get a token from a previously created user
  LOGIN_RESPONSE=$(curl -s -X POST \
    "${REST_URL}/api/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{
      "email": "test@example.com",
      "password": "password123"
    }')
  TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')
  
  if [[ -z "$TOKEN" ]]; then
    print_error "Could not get a token. Some tests will fail."
  else
    print_success "Got token from previous user"
  fi
else
  print_success "Authentication successful"
fi

# 4. Test List Users with Authentication
print_header "4. Testing List Users (Authenticated)"
USER_LIST_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/users" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$USER_LIST_RESPONSE" | jq .

# Check if we got a valid response
if [[ $(echo "$USER_LIST_RESPONSE" | jq -e '.data') ]]; then
  print_success "User list retrieved successfully"
else
  print_error "Failed to retrieve user list"
fi

# 5. Test Todo Functionality
print_header "5. Testing Todo Creation (Authenticated)"
TODO_RESPONSE=$(curl -s -X POST \
  "${REST_URL}/api/todos" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "Fruit",
    "name": "Apple"
  }')
echo "$TODO_RESPONSE" | jq .

# Extract todo ID from response
TODO_ID=$(echo "$TODO_RESPONSE" | jq -r '.id // empty')

if [[ ! -z "$TODO_ID" ]]; then
  print_success "Todo created successfully with ID: $TODO_ID"
else
  print_error "Failed to create todo"
fi

# 6. Test List Todos
print_header "6. Testing List Todos (Authenticated)"
TODO_LIST_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/todos" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$TODO_LIST_RESPONSE" | jq .

# Check if we got a valid response
if [[ $(echo "$TODO_LIST_RESPONSE" | jq -e '.main') ]]; then
  print_success "Todo list retrieved successfully"
else
  print_error "Failed to retrieve todo list"
fi

# 7. Test Click Todo
print_header "7. Testing Todo Click (Authenticated)"
if [[ ! -z "$TODO_ID" ]]; then
  CLICK_RESPONSE=$(curl -s -X POST \
    "${REST_URL}/api/todos/${TODO_ID}/click" \
    -H "Authorization: Bearer ${TOKEN}")
  echo "$CLICK_RESPONSE" | jq .
  
  # Check if status is now COLUMN
  if [[ $(echo "$CLICK_RESPONSE" | jq -r '.status // empty') == "COLUMN" ]]; then
    print_success "Todo clicked and moved to COLUMN status"
  else
    print_error "Failed to click todo"
  fi
else
  print_error "Cannot test click - no todo ID available"
fi

# 8. Wait for auto-return
print_info "Waiting 6 seconds for auto-return..."
sleep 6

# 9. Test List Todos Again to Verify Auto-Return
print_header "8. Testing List Todos Again to Verify Auto-Return"
TODO_LIST_AFTER_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/todos" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$TODO_LIST_AFTER_RESPONSE" | jq .

# Check if the todo is back in the main list
if [[ ! -z "$TODO_ID" ]]; then
  # Find the todo in the main list
  TODO_STATUS=$(echo "$TODO_LIST_AFTER_RESPONSE" | jq -r --arg id "$TODO_ID" '.main[] | select(.id==$id) | .status // empty')
  
  if [[ "$TODO_STATUS" == "MAIN" ]]; then
    print_success "Todo auto-returned to MAIN status successfully"
  else
    print_error "Todo did not auto-return to MAIN status"
  fi
else
  print_error "Cannot verify auto-return - no todo ID available"
fi

# 10. gRPC Tests
if [[ "$SKIP_GRPC" == false ]]; then
  print_header "9. Testing gRPC Services"
  
  # List available services
  print_info "Available gRPC services:"
  grpcurl -plaintext $GRPC_URL list
  
  # List methods for the user service
  print_info "Available methods for the User service:"
  grpcurl -plaintext $GRPC_URL list backend.UserService
  
  # Test GetUser method via gRPC
  if [[ ! -z "$TOKEN" ]]; then
    print_info "Testing GetUser via gRPC (with a mock ID):"
    grpcurl -plaintext -H "Authorization: Bearer ${TOKEN}" -d '{"id": "some-user-id"}' $GRPC_URL backend.UserService/GetUser
  else
    print_error "Cannot test GetUser via gRPC - no token available"
  fi
fi

print_header "Tests Completed"