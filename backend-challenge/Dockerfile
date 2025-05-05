FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies including protoc
RUN apk add --no-cache git protobuf protobuf-dev

# Install protoc-gen-go and protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# Copy the source code
COPY . .

# Install required packages
RUN go mod tidy

# Generate Go code from protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/user.proto

# List generated files for debugging
RUN find api/proto -type f -name "*.pb.go" || echo "No generated files found"

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o userservice ./cmd/api

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/userservice /app/userservice

# Expose ports
EXPOSE 8080 50051

# Set the binary as the entrypoint
ENTRYPOINT ["/app/userservice"]