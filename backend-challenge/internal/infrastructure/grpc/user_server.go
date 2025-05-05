package grpc

import (
	"context"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/service"
	"backend-challenge/internal/infrastructure/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Temporary type definitions for compilation
type UnimplementedUserServiceServer struct{}
func (UnimplementedUserServiceServer) CreateUser(context.Context, *CreateUserRequest) (*UserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (UnimplementedUserServiceServer) GetUser(context.Context, *GetUserRequest) (*UserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedUserServiceServer) ListUsers(context.Context, *ListUsersRequest) (*ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}
func (UnimplementedUserServiceServer) UpdateUser(context.Context, *UpdateUserRequest) (*UserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedUserServiceServer) DeleteUser(context.Context, *DeleteUserRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedUserServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

// Temporary type stubs for compilation
type CreateUserRequest struct {
	Name     string
	Email    string
	Password string
}
type GetUserRequest struct {
	Id string
}
type ListUsersRequest struct {
	Page     int32
	PageSize int32
}
type UpdateUserRequest struct {
	Id    string
	Name  string
	Email string
}
type DeleteUserRequest struct {
	Id string
}
type LoginRequest struct {
	Email    string
	Password string
}
type LoginResponse struct {
	Token string
	User  *User
}
type ListUsersResponse struct {
	Users      []*User
	Page       int32
	PageSize   int32
	TotalItems int64
}
type UserResponse struct {
	User *User
}
type User struct {
	Id        string
	Name      string
	Email     string
	CreatedAt *timestamppb.Timestamp
}

// UserServer implements the UserService gRPC service
type UserServer struct {
	userService service.UserService
	authService auth.AuthService
}

// NewUserServer creates a new UserServer
func NewUserServer(userService service.UserService, authService auth.AuthService) *UserServer {
	return &UserServer{
		userService: userService,
		authService: authService,
	}
}

// Register registers the gRPC server
func Register(s *grpc.Server, userService service.UserService, authService auth.AuthService) {
	// Temporarily do nothing 
	_ = s
	_ = userService
	_ = authService
}

// mapUserToProto converts a user model to a protobuf user message
func mapUserToProto(user *model.User) *User {
	return &User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}

// mapDomainErrorToGRPC maps domain errors to gRPC status errors
func mapDomainErrorToGRPC(err error) error {
	switch err {
	case service.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case service.ErrEmailExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case service.ErrInvalidID:
		return status.Error(codes.InvalidArgument, err.Error())
	case service.ErrInvalidPassword:
		return status.Error(codes.Unauthenticated, err.Error())
	case auth.ErrMissingToken, auth.ErrInvalidToken, auth.ErrTokenExpired:
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}