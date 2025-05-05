package grpc

import (
	"context"
	"time"

	"github.com/7-solutions/backend-challenge/internal/domain/model"
	"github.com/7-solutions/backend-challenge/internal/domain/service"
	"github.com/7-solutions/backend-challenge/internal/infrastructure/auth"
	pb "github.com/7-solutions/backend-challenge/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServer implements the UserService gRPC service
type UserServer struct {
	userService service.UserService
	authService auth.AuthService
	pb.UnimplementedUserServiceServer
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
	pb.RegisterUserServiceServer(s, NewUserServer(userService, authService))
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	// Create input
	input := &model.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Register user
	user, err := s.userService.Register(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Convert user to protobuf message
	pbUser := mapUserToProto(user)

	return &pb.UserResponse{
		User: pbUser,
	}, nil
}

// GetUser gets a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	// Authorize request
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userService.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Convert user to protobuf message
	pbUser := mapUserToProto(user)

	return &pb.UserResponse{
		User: pbUser,
	}, nil
}

// ListUsers lists users with pagination
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Authorize request
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}

	// Set default values
	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get users
	users, err := s.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Get total count
	totalItems, err := s.userService.CountUsers(ctx)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Convert users to protobuf messages
	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, mapUserToProto(user))
	}

	return &pb.ListUsersResponse{
		Users:      pbUsers,
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalItems: totalItems,
	}, nil
}

// UpdateUser updates a user
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	// Authorize request
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}

	// Check if user is updating their own profile
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if userID != req.Id {
		return nil, status.Error(codes.PermissionDenied, "Cannot update other users")
	}

	// Create input
	input := &model.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	}

	// Update user
	user, err := s.userService.UpdateUser(ctx, req.Id, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Convert user to protobuf message
	pbUser := mapUserToProto(user)

	return &pb.UserResponse{
		User: pbUser,
	}, nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	// Authorize request
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}

	// Check if user is deleting their own profile
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if userID != req.Id {
		return nil, status.Error(codes.PermissionDenied, "Cannot delete other users")
	}

	// Delete user
	err = s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &emptypb.Empty{}, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Create input
	input := &model.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	// Login user
	user, err := s.userService.Login(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// Generate token
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate token")
	}

	// Convert user to protobuf message
	pbUser := mapUserToProto(user)

	return &pb.LoginResponse{
		Token: token,
		User:  pbUser,
	}, nil
}

// mapUserToProto converts a user model to a protobuf user message
func mapUserToProto(user *model.User) *pb.User {
	return &pb.User{
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

// authorize authorizes a request using JWT token from metadata
func (s *UserServer) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "Missing metadata")
	}

	// Get authorization token
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return status.Error(codes.Unauthenticated, "Missing authorization metadata")
	}

	// Extract token (remove "Bearer " prefix if present)
	tokenString := authHeader[0]
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Validate token
	_, err := s.authService.ValidateToken(tokenString)
	if err != nil {
		return status.Error(codes.Unauthenticated, "Invalid token")
	}

	return nil
}

// getUserIDFromContext gets the user ID from the JWT token in context
func (s *UserServer) getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Missing metadata")
	}

	// Get authorization token
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "Missing authorization metadata")
	}

	// Extract token (remove "Bearer " prefix if present)
	tokenString := authHeader[0]
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Validate token and get claims
	claims, err := s.authService.ValidateToken(tokenString)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "Invalid token")
	}

	return claims.UserID, nil
}