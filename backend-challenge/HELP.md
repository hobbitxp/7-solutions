# Backend Challenge - Usage Guide

This guide explains how to set up, run, and test the Backend Challenge project.

## Table of Contents

- [Installation](#installation)
- [Running the Project](#running-the-project)
- [Testing the API](#testing-the-api)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or newer (for non-Docker development)
- MongoDB (for non-Docker running)

### Testing Tools (Optional)

- `jq` - For formatting JSON in tests
  - Install on macOS: `brew install jq`
  - Install on Linux: `apt-get install jq`

- `grpcurl` - For testing gRPC API
  - Install on macOS: `brew install grpcurl`
  - Install on Linux: Download from [GitHub](https://github.com/fullstorydev/grpcurl/releases)

### Cloning the Project

```bash
git clone <repository-url>
cd backend-challenge
```

## Running the Project

### Method 1: Using Docker Compose (Recommended)

Run both API and MongoDB in containers:

```bash
# Build images and run containers
docker compose up -d

# View logs
docker compose logs -f

# Stop everything
docker compose down
```

### Method 2: Running Directly with Go

If you want to run the application directly without Docker:

```bash
# Set up .env file (create from template)
cat > .env << EOL
MONGODB_URI=mongodb://localhost:27017
DB_NAME=user_service
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h
PORT=8080
GRPC_PORT=50051
EOL

# Run MongoDB with Docker (if you don't have MongoDB)
docker run -d -p 27017:27017 --name mongo-local mongo:latest

# Run the application
go run cmd/api/main.go
```

## Testing the API

The project comes with several testing scripts you can use:

### 1. Test Both REST API and gRPC

```bash
chmod +x test_all.sh
./test_all.sh
```

### 2. Test REST API Only

```bash
chmod +x test_rest.sh
./test_rest.sh
```

### 3. Test gRPC API Only

```bash
chmod +x test_grpc.sh
./test_grpc.sh
```

### 4. Generate Protocol Buffer Files (For Development)

```bash
chmod +x gen_proto.sh
./gen_proto.sh
```

### 5. Manual REST API Testing

You can also test the REST API manually using tools like `curl`, Postman, or any HTTP client:

#### User Authentication

1. **Register a new user**:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'
```

2. **Login and get JWT token**:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com", 
    "password": "password123"
  }'
```

3. **Store the token in a variable (for use in other commands)**:
```bash
TOKEN=$(curl -s -X POST \
  "http://localhost:8080/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }' | jq -r '.token')
```

#### User Management

4. **List all users**:
```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN"
```

5. **Get a specific user**:
```bash
curl -X GET http://localhost:8080/api/users/{user_id} \
  -H "Authorization: Bearer $TOKEN"
```

#### Todo Management

6. **Create a new Todo**:
```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Fruit",
    "name": "Apple"
  }'
```

7. **List all Todos**:
```bash
curl -X GET http://localhost:8080/api/todos \
  -H "Authorization: Bearer $TOKEN"
```

8. **Click a Todo** (replace {todo_id} with an actual ID):
```bash
curl -X POST http://localhost:8080/api/todos/{todo_id}/click \
  -H "Authorization: Bearer $TOKEN"
```

9. **Delete a Todo** (replace {todo_id} with an actual ID):
```bash
curl -X DELETE http://localhost:8080/api/todos/{todo_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Project Structure

This project uses Hexagonal (or Ports & Adapters) architecture:

```
backend-challenge/
├── api/                     # API Definitions
│   └── proto/               # Protocol Buffers definitions
├── cmd/                     # Application entry points
│   └── api/                 # REST and gRPC API
├── internal/                # Application internals
│   ├── application/         # Application layer (handlers)
│   ├── domain/              # Domain layer (business logic)
│   │   ├── model/           # Domain models
│   │   ├── repository/      # Repository interfaces
│   │   └── service/         # Business logic services
│   └── infrastructure/      # Infrastructure layer
│       ├── auth/            # Authentication
│       ├── grpc/            # gRPC server
│       ├── middleware/      # HTTP middlewares
│       └── repository/      # Repository implementations
└── pkg/                     # Shared packages
```

## API Endpoints

### REST API (Port 8080)

#### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token

#### User Management
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get a specific user
- `PUT /api/users/:id` - Update a user
- `DELETE /api/users/:id` - Delete a user

#### Todo Management
- `GET /api/todos` - List all todos grouped by status and type
- `POST /api/todos` - Create a new todo
- `GET /api/todos/:id` - Get a specific todo
- `PUT /api/todos/:id` - Update a todo
- `DELETE /api/todos/:id` - Delete a todo
- `POST /api/todos/:id/click` - Click a todo to move it to its type column (auto-returns after 5 seconds)

### gRPC API (Port 50051)

#### UserService
- `CreateUser` - Create a new user
- `GetUser` - Get a user's details
- `UpdateUser` - Update a user
- `DeleteUser` - Delete a user
- `ListUsers` - List all users

## Special Features

### Todo Auto-Return

The project includes an "auto-return" feature for Todos. When you click on a Todo, it moves from the main list to its type column (Fruit or Vegetable) and automatically returns to the main list after 5 seconds.

## Troubleshooting

### MongoDB Connection Issues

If the application can't connect to MongoDB:

```bash
# Check if MongoDB container is running
docker ps -a | grep mongo

# Start MongoDB container if it's stopped
docker start user-service-mongo
```

### Resetting Docker Containers

If you need to start fresh:

```bash
docker compose down -v  # Remove containers and volumes
docker compose up -d    # Create and start containers again
```

### Checking Logs

```bash
# View API server logs
docker logs user-service -f

# View MongoDB logs
docker logs user-service-mongo -f
```