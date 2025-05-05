# Backend Golang Coding Test

## Objective
Build a simple RESTful API in Golang that manages a list of users and todo items. Use MongoDB for persistence, JWT for authentication, and follow clean code practices. The API also includes a special Todo feature with an auto-return functionality.

---

## Requirements

### 1. User Model
Each user should have:
- `ID` (auto-generated)
- `Name` (string)
- `Email` (string, unique)
- `Password` (hashed)
- `CreatedAt` (timestamp)

---

### 2. Authentication

#### Functions
- Register a new user.
- Authenticate user and return a JWT.

#### JWT
- Use JWT for protecting endpoints.
- Use middleware to validate tokens.
- Use HMAC (HS256) with a secret key.

---

### 3. User Functions

- Create a new user.
- Fetch user by ID.
- List all users.
- Update a user's name or email.
- Delete a user.

---

### 4. MongoDB Integration
- Use the official Go MongoDB driver.
- Store and retrieve users from MongoDB.

---

### 5. Middleware
- Logging middleware that logs HTTP method, path, and execution time.

---

### 6. Concurrency Tasks
- Run a background goroutine every 10 seconds that logs the number of users in the DB.
- Implement an auto-return feature for todo items that automatically returns them to the main list after 5 seconds.

---

### 7. Todo Functionality
- Todo items are categorized by type (Fruit, Vegetable) and status (MAIN, COLUMN).
- Todo items can be "clicked" to move them from the main list to their appropriate category column.
- Clicked items automatically return to the main list after 5 seconds.
- Background processing checks for items that should be returned.

---

### 8. Testing
Write unit tests

Use Go’s `testing` package. Mock MongoDB where possible.

---

## Bonus (Optional)

- Add Docker + `docker-compose` for API + MongoDB.
- Use Go interfaces to abstract MongoDB operations for testability.
- Add input validation (e.g., required fields, valid email).
- Implement graceful shutdown using `context.Context`.
- **gRPC Version**
  - Create a `.proto` file for `CreateUser` and `GetUser`.
  - Implement a gRPC server.
  - (Optional) Secure gRPC with token metadata.
- **Hexagonal Architecture**
  - Structure the project using hexagonal (ports & adapters) architecture:
    - Separate domain, application, and infrastructure layers.
    - Use interfaces for data access and external dependencies.
    - Keep business logic decoupled from frameworks and DB drivers.

---

## Submission Guidelines

- Submit a GitHub repo or zip file.
- Include a `README.md` with:
  - Project setup and run instructions
  - JWT token usage guide
  - Sample API requests/responses
  - Any assumptions or decisions made

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token

### User Management
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get a specific user
- `PUT /api/users/:id` - Update a user
- `DELETE /api/users/:id` - Delete a user

### Todo Management
- `GET /api/todos` - List all todos grouped by status and type
- `POST /api/todos` - Create a new todo
- `GET /api/todos/:id` - Get a specific todo
- `PUT /api/todos/:id` - Update a todo
- `DELETE /api/todos/:id` - Delete a todo
- `POST /api/todos/:id/click` - Click a todo to move it to its type column

---

## Evaluation Criteria

- Code quality, structure, and readability
- REST API correctness and completeness
- JWT implementation and security
- MongoDB usage and abstraction
- Bonus: gRPC, Docker, validation, shutdown
- Testing coverage and mocking
- Use of idiomatic Go