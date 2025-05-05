#!/bin/bash

# Test REST API Endpoints

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Define base URL
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

# Check if jq is installed
if ! command -v jq &> /dev/null; then
  print_error "jq is not installed. Please install it to run this script."
  echo "  macOS: brew install jq"
  echo "  Linux: apt-get install jq"
  exit 1
fi

print_header "Starting REST API Tests"

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

# 4. Test User Management
print_header "4. User Management Tests"

# 4.1 List Users
print_info "4.1 List Users"
USER_LIST_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/users" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$USER_LIST_RESPONSE" | jq .

# Get user ID for other tests
USER_ID=$(echo "$USER_LIST_RESPONSE" | jq -r '.data[0].id // empty')

# 4.2 Get User by ID
if [[ ! -z "$USER_ID" ]]; then
  print_info "4.2 Get User by ID: $USER_ID"
  USER_GET_RESPONSE=$(curl -s -X GET \
    "${REST_URL}/api/users/${USER_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  echo "$USER_GET_RESPONSE" | jq .
else
  print_error "Cannot get user by ID - no user ID available"
fi

# 5. Todo Management Tests
print_header "5. Todo Management Tests"

# 5.1 Create Todo
print_info "5.1 Create Todo"
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

# 5.2 List Todos
print_info "5.2 List Todos"
TODO_LIST_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/todos" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$TODO_LIST_RESPONSE" | jq .

# 5.3 Get Todo by ID
if [[ ! -z "$TODO_ID" ]]; then
  print_info "5.3 Get Todo by ID: $TODO_ID"
  TODO_GET_RESPONSE=$(curl -s -X GET \
    "${REST_URL}/api/todos/${TODO_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  echo "$TODO_GET_RESPONSE" | jq .
else
  print_error "Cannot get todo by ID - no todo ID available"
fi

# 5.4 Update Todo
if [[ ! -z "$TODO_ID" ]]; then
  print_info "5.4 Update Todo"
  TODO_UPDATE_RESPONSE=$(curl -s -X PUT \
    "${REST_URL}/api/todos/${TODO_ID}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H 'Content-Type: application/json' \
    -d '{
      "name": "Updated Apple"
    }')
  echo "$TODO_UPDATE_RESPONSE" | jq .
else
  print_error "Cannot update todo - no todo ID available"
fi

# 5.5 Click Todo
if [[ ! -z "$TODO_ID" ]]; then
  print_info "5.5 Click Todo"
  CLICK_RESPONSE=$(curl -s -X POST \
    "${REST_URL}/api/todos/${TODO_ID}/click" \
    -H "Authorization: Bearer ${TOKEN}")
  echo "$CLICK_RESPONSE" | jq .
else
  print_error "Cannot click todo - no todo ID available"
fi

# 5.6 Wait for auto-return
print_info "Waiting 6 seconds for auto-return..."
sleep 6

# 5.7 Verify Auto-Return
print_info "5.7 Verify Auto-Return"
TODO_LIST_AFTER_RESPONSE=$(curl -s -X GET \
  "${REST_URL}/api/todos" \
  -H "Authorization: Bearer ${TOKEN}")
echo "$TODO_LIST_AFTER_RESPONSE" | jq .

# 5.8 Delete Todo
if [[ ! -z "$TODO_ID" ]]; then
  print_info "5.8 Delete Todo"
  TODO_DELETE_RESPONSE=$(curl -s -X DELETE \
    "${REST_URL}/api/todos/${TODO_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  echo "$TODO_DELETE_RESPONSE" | jq .
else
  print_error "Cannot delete todo - no todo ID available"
fi

print_header "REST API Tests Completed"