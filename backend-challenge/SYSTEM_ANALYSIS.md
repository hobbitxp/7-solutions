# System Analysis

This document provides a detailed overview of the system architecture and functionality.

## Core Features

### BN (User Service)
- User registration and authentication
- JWT-based access control
- CRUD operations for user management
- MongoDB persistence
- gRPC interface alongside REST

### FN (Todo Service)
- Todo item management with categorization (Fruit/Vegetable)
- Auto-return functionality (items return to main list after 5 seconds)
- Status tracking (MAIN/COLUMN)
- JWT-protected operations

## Data Models

```go
// BN: User Service
type User struct {
  ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
  Name         string             `bson:"name"          json:"name"`
  Email        string             `bson:"email"         json:"email"`
  PasswordHash string             `bson:"password_hash" json:"-"`
  CreatedAt    time.Time          `bson:"created_at"    json:"created_at"`
  UpdatedAt    time.Time          `bson:"updated_at"    json:"updated_at"`
}

// FN: Todo List
type TodoItem struct {
  ID        string     `json:"id" bson:"_id,omitempty"`   // UUID
  Type      ItemType   `json:"type" bson:"type"`          // "Fruit" | "Vegetable"
  Name      string     `json:"name" bson:"name"`
  Status    ItemStatus `json:"status" bson:"status"`      // "MAIN" | "COLUMN"
  ClickedAt time.Time  `json:"clicked_at,omitempty" bson:"clicked_at,omitempty"`
  ReturnAt  time.Time  `json:"return_at,omitempty" bson:"return_at,omitempty"`
  CreatedAt time.Time  `json:"created_at" bson:"created_at"`
  UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
}
```

## BN: REST Endpoints (User Management & Auth)

| Method | Path                | Description                                    |
|--------|---------------------|------------------------------------------------|
| POST   | /api/auth/register  | Register a new user (accepts name, email, password) |
| POST   | /api/auth/login     | Authenticate user (returns JWT token)          |
| GET    | /api/users          | List all users (requires JWT)                  |
| GET    | /api/users/{id}     | Get user by ID (requires JWT)                  |
| PUT    | /api/users/{id}     | Update user (requires JWT)                     |
| DELETE | /api/users/{id}     | Delete user (requires JWT)                     |

## FN: REST Endpoints (Todo CRUD + Click)

| Method | Path                | Description                                    |
|--------|---------------------|------------------------------------------------|
| GET    | /api/todos          | List all todos (grouped by main and column)    |
| GET    | /api/todos/{id}     | Get todo details by ID                         |
| POST   | /api/todos          | Create new todo ({ type, name })               |
| PUT    | /api/todos/{id}     | Update todo name or type                       |
| DELETE | /api/todos/{id}     | Delete todo                                    |
| POST   | /api/todos/{id}/click | Move todo to column or return to main        |

## Service Interfaces

```go
// BN: Auth Service
type AuthService interface {
  // Register creates a new user
  Register(ctx context.Context, input *model.RegisterUserInput) (*model.User, error)
  
  // Login authenticates a user and returns a JWT token
  Login(ctx context.Context, input *model.LoginUserInput) (string, *model.User, error)
  
  // ValidateToken validates a JWT token and returns the claims
  ValidateToken(tokenString string) (*auth.Claims, error)
  
  // ExtractTokenFromRequest extracts a token from an HTTP request
  ExtractTokenFromRequest(r *http.Request) (string, error)
}

// BN: User Service
type UserService interface {
  // GetByID fetches a user by ID
  GetByID(ctx context.Context, id string) (*model.User, error)
  
  // List returns all users with pagination
  List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
  
  // Create creates a new user
  Create(ctx context.Context, input *model.CreateUserInput) (*model.User, error)
  
  // Update updates a user
  Update(ctx context.Context, id string, input *model.UpdateUserInput) (*model.User, error)
  
  // Delete removes a user
  Delete(ctx context.Context, id string) error
  
  // CountUsers returns the total number of users
  CountUsers(ctx context.Context) (int64, error)
}

// FN: Todo Service
type TodoService interface {
  // Create creates a new todo item
  Create(ctx context.Context, input *model.CreateTodoInput) (*model.TodoItem, error)
  
  // GetByID fetches a todo item by ID
  GetByID(ctx context.Context, id string) (*model.TodoItem, error)
  
  // Update updates a todo item
  Update(ctx context.Context, id string, input *model.UpdateTodoInput) (*model.TodoItem, error)
  
  // Delete removes a todo item
  Delete(ctx context.Context, id string) error
  
  // List returns all todo items grouped by status and type
  List(ctx context.Context) (*model.TodosGrouped, error)
  
  // Click handles the click action on a todo item
  Click(ctx context.Context, id string) (*model.TodoItem, error)
  
  // TimeoutReturn handles the automatic return of a todo item to the main list
  TimeoutReturn(ctx context.Context, id string) error
  
  // ReturnTimedOutItems returns all todo items that should be returned to the main list
  ReturnTimedOutItems(ctx context.Context, currentTime string) (int, error)
}
```

## Repository Interfaces

```go
// BN: User Repository
type UserRepository interface {
  // GetByID fetches a user by ID
  GetByID(ctx context.Context, id string) (*model.User, error)
  
  // GetByEmail fetches a user by email
  GetByEmail(ctx context.Context, email string) (*model.User, error)
  
  // Create creates a new user
  Create(ctx context.Context, user *model.User) error
  
  // Update updates a user
  Update(ctx context.Context, user *model.User) error
  
  // Delete removes a user
  Delete(ctx context.Context, id string) error
  
  // List returns all users
  List(ctx context.Context, skip, limit int) ([]*model.User, error)
  
  // CountUsers returns the total number of users
  CountUsers(ctx context.Context) (int64, error)
}

// FN: Todo Repository
type TodoRepository interface {
  // Create creates a new todo item in the database
  Create(ctx context.Context, todo *model.TodoItem) error
  
  // GetByID fetches a todo item by ID
  GetByID(ctx context.Context, id string) (*model.TodoItem, error)
  
  // Update updates a todo item in the database
  Update(ctx context.Context, todo *model.TodoItem) error
  
  // Delete removes a todo item from the database
  Delete(ctx context.Context, id string) error
  
  // List returns all todo items
  List(ctx context.Context) ([]*model.TodoItem, error)
  
  // FindByStatus returns all todo items with a specific status
  FindByStatus(ctx context.Context, status model.ItemStatus) ([]*model.TodoItem, error)
  
  // FindByTypeAndStatus returns all todo items with a specific type and status
  FindByTypeAndStatus(ctx context.Context, itemType model.ItemType, status model.ItemStatus) ([]*model.TodoItem, error)
  
  // UpdateStatus updates the status of a todo item
  UpdateStatus(ctx context.Context, id string, status model.ItemStatus) error
  
  // FindToReturn finds all todo items that should be returned to the main list
  FindToReturn(ctx context.Context, currentTime string) ([]*model.TodoItem, error)
}
```

## Routing Setup (Gorilla Mux)

```go
// Auth + User
handler.RegisterAuthHandler(r, authService, userService)
handler.RegisterUserHandler(r, userService, authService)

// Todo
handler.RegisterTodoHandler(r, todoService, authService)
```

## Authentication Middleware

```go
// Create auth middleware
func createAuthMiddleware(authService auth.AuthService) mux.MiddlewareFunc {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // Extract token from request
      tokenString, err := authService.ExtractTokenFromRequest(r)
      if err != nil {
        respondWithError(w, err, http.StatusUnauthorized)
        return
      }
      
      // Validate token
      claims, err := authService.ValidateToken(tokenString)
      if err != nil {
        respondWithError(w, err, http.StatusUnauthorized)
        return
      }
      
      // Store user ID in request context
      ctx := r.Context()
      r = r.WithContext(ctx)
      
      // Set user ID in header for use in handlers
      r.Header.Set("X-User-ID", claims.UserID)
      
      next.ServeHTTP(w, r)
    })
  }
}
```

## Auto-Return Timer Logic

```go
// Click handles the click action on a todo item
func (s *todoService) Click(ctx context.Context, id string) (*model.TodoItem, error) {
  // Get todo item
  todo, err := s.repo.GetByID(ctx, id)
  if err != nil {
    return nil, ErrTodoNotFound
  }
  
  // Mark as clicked and update status
  todo.Click()
  
  // Save to repository
  if err := s.repo.Update(ctx, todo); err != nil {
    return nil, err
  }
  
  // Schedule auto-return
  time.AfterFunc(5*time.Second, func() {
    // Create a background context since the HTTP context will be gone
    bgCtx := context.Background()
    // Call TimeoutReturn
    _ = s.TimeoutReturn(bgCtx, id)
  })
  
  return todo, nil
}

// TimeoutReturn handles the automatic return of a todo item to the main list
func (s *todoService) TimeoutReturn(ctx context.Context, id string) error {
  // Get todo item
  todo, err := s.repo.GetByID(ctx, id)
  if err != nil {
    return ErrTodoNotFound
  }
  
  // Only return if still in COLUMN status
  if todo.Status == model.StatusColumn {
    // Return to main list
    todo.Return()
    
    // Save to repository
    if err := s.repo.Update(ctx, todo); err != nil {
      return err
    }
  }
  
  return nil
}
```

## Background Task (Auto-Return Check)

```go
// Background goroutine that checks and returns todo items that have reached their return time
func startBackgroundTodoReturn(ctx context.Context, todoService service.TodoService) {
  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()
  
  for {
    select {
    case <-ticker.C:
      now := time.Now().Format(time.RFC3339)
      count, err := todoService.ReturnTimedOutItems(ctx, now)
      if err != nil {
        log.Printf("Error returning timed out todo items: %v", err)
      } else if count > 0 {
        log.Printf("Returned %d timed out todo items", count)
      }
    case <-ctx.Done():
      log.Println("Stopping background todo return checker")
      return
    }
  }
}
```

## Communication Protocols

### REST API
- Main interface for client applications
- JSON data format
- JWT authentication
- HTTP status codes for error handling

### gRPC API (Optional/Bonus)
- High-performance RPC framework
- Protocol Buffers for compact serialization
- Supports server reflection for discovery
- Used mainly for internal service communication

## Deployment

The system is deployed using Docker containers:

```yaml
services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"   # REST API
      - "50051:50051" # gRPC
    environment:
      - MONGODB_URI=mongodb://user-service-mongo:27017
      - DB_NAME=user_service
      - JWT_SECRET=your-secret-key-change-in-production
      - JWT_EXPIRY=24h
    depends_on:
      - mongo

  mongo:
    image: mongo:latest
    volumes:
      - mongo-data:/data/db
```

## Key Features Implementation

### JWT Authentication
- Token generation upon login
- Token validation middleware
- Proper error handling for expired/invalid tokens

### MongoDB Integration
- Repository pattern for database access
- BSON tagging for proper MongoDB serialization
- Proper error handling for database operations

### Hexagonal Architecture
- Domain, application, and infrastructure layers
- Clear separation of concerns
- Dependency injection for easier testing

### Error Handling
- Domain-specific errors
- HTTP status code mapping
- Consistent error response format

### Auto-Return Functionality
- Scheduled background tasks
- Concurrent processing with goroutines
- Time-based status updates

### Graceful Shutdown
- Context cancellation for cleanup
- Signal handling (SIGINT, SIGTERM)
- Resource cleanup (database connections, etc.)