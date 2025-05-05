# User Service API

A RESTful API built in Golang that manages a list of users with MongoDB for persistence and JWT for authentication.

## Features

- User management (create, read, update, delete)
- Authentication with JWT
- MongoDB integration
- Hexagonal Architecture
- Input validation
- Graceful shutdown
- gRPC server
- Docker and docker-compose configuration

## Architecture

The application follows the hexagonal (ports & adapters) architecture:

- **Domain layer**: Contains the core business logic and entities
- **Application layer**: Contains the application services and handlers
- **Infrastructure layer**: Contains adapters for external dependencies

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MongoDB
- Docker and docker-compose (optional)

### Running with Docker

The easiest way to start the application is using Docker:

```bash
# Clone the repository
git clone <repository-url>
cd backend-challenge

# Start the application with Docker Compose
docker-compose up -d
```

### Running Locally

```bash
# Clone the repository
git clone <repository-url>
cd backend-challenge

# Copy the example env file
cp .env.example .env

# Update the .env file with your configuration

# Run the application
go run cmd/api/main.go
```

## API Endpoints

### Authentication

- **POST /api/auth/register** - Register a new user
- **POST /api/auth/login** - Login and get JWT token

### User Management

- **GET /api/users** - List all users (requires authentication)
- **GET /api/users/{id}** - Get a user by ID (requires authentication)
- **PUT /api/users/{id}** - Update a user (requires authentication)
- **DELETE /api/users/{id}** - Delete a user (requires authentication)

## gRPC Services

The application also provides a gRPC API on port 50051 with the following services:

- **CreateUser** - Create a new user
- **GetUser** - Get a user by ID
- **ListUsers** - List all users with pagination
- **UpdateUser** - Update a user
- **DeleteUser** - Delete a user
- **Login** - Login and get JWT token

## JWT Authentication

The API uses JWT (JSON Web Tokens) for authentication:

1. Register or login to get a JWT token
2. Include the token in the `Authorization` header as `Bearer <token>` for protected endpoints

## Testing

Run the tests with:

```bash
go test ./...
```
