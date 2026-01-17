// Package main implements the User Service.
// Microservice for user management and authentication.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
	"github.com/yourusername/iot-platform/services/user-service/storage"
)

// UserServer implements pb.UserServiceServer interface.
// Uses pluggable Storage backend (PostgreSQL or in-memory).
type UserServer struct {
	pb.UnimplementedUserServiceServer
	storage storage.Storage
}

// NewUserServer creates a new server instance with the given storage backend.
func NewUserServer(store storage.Storage) *UserServer {
	return &UserServer{
		storage: store,
	}
}

// Register creates a new user account with hashed password.
func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("📥 Register: email=%s, name=%s, role=%s", req.Email, req.Name, req.Role)

	// Validate input
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}

	// Default role to "user" if not specified
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Validate role
	if role != "admin" && role != "user" && role != "device" {
		return nil, status.Error(codes.InvalidArgument, "invalid role: must be admin, user, or device")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Failed to hash password: %v", err)
		return nil, status.Error(codes.Internal, "failed to process password")
	}

	// Create user
	now := time.Now().Unix()
	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		Role:      role,
		CreatedAt: now,
		IsActive:  true,
	}

	createdUser, err := s.storage.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		log.Printf("❌ Failed to create user: %v", err)
		return nil, err
	}

	log.Printf("✅ User registered: id=%s, email=%s", createdUser.Id, createdUser.Email)
	return &pb.RegisterResponse{
		User:    createdUser,
		Message: "User registered successfully",
	}, nil
}

// Authenticate verifies user credentials and returns user info.
func (s *UserServer) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	log.Printf("📥 Authenticate: email=%s", req.Email)

	// Validate input
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password required")
	}

	// Get user by email
	user, err := s.storage.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("❌ User not found: %s", req.Email)
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Check if user is active
	if !user.IsActive {
		log.Printf("❌ User account is inactive: %s", req.Email)
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "Account is inactive",
		}, nil
	}

	// Get password hash
	passwordHash, err := s.storage.GetPasswordHash(ctx, req.Email)
	if err != nil {
		log.Printf("❌ Failed to retrieve password hash: %v", err)
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		log.Printf("❌ Invalid password for: %s", req.Email)
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Update last login
	if err := s.storage.UpdateLastLogin(ctx, user.Id); err != nil {
		log.Printf("⚠️  Failed to update last login: %v", err)
	}

	log.Printf("✅ Authentication successful: %s", req.Email)
	return &pb.AuthenticateResponse{
		User:    user,
		Success: true,
		Message: "Authentication successful",
	}, nil
}

// GetUser retrieves a user by ID.
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log.Printf("📥 GetUser: id=%s", req.Id)

	user, err := s.storage.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ User found: id=%s, email=%s", user.Id, user.Email)
	return &pb.GetUserResponse{User: user}, nil
}

// GetUserByEmail retrieves a user by email.
func (s *UserServer) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserByEmailResponse, error) {
	log.Printf("📥 GetUserByEmail: email=%s", req.Email)

	user, err := s.storage.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ User found: id=%s, email=%s", user.Id, user.Email)
	return &pb.GetUserByEmailResponse{User: user}, nil
}

// ListUsers returns paginated user list.
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Printf("📥 ListUsers: page=%d, pageSize=%d, role=%s", req.Page, req.PageSize, req.Role)

	users, total, err := s.storage.ListUsers(ctx, req.Page, req.PageSize, req.Role)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ %d users found", len(users))
	return &pb.ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateUser updates an existing user.
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Printf("📥 UpdateUser: id=%s", req.Id)

	user := &pb.User{
		Id:       req.Id,
		Name:     req.Name,
		Role:     req.Role,
		IsActive: req.IsActive,
	}

	updatedUser, err := s.storage.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ User updated: id=%s", updatedUser.Id)
	return &pb.UpdateUserResponse{User: updatedUser}, nil
}

// DeleteUser removes a user by ID.
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Printf("📥 DeleteUser: id=%s", req.Id)

	err := s.storage.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ User deleted: id=%s", req.Id)
	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// main starts the User Service gRPC server.
//
// Configuration via environment variables:
//   - USER_SERVICE_PORT: gRPC server port (default: 8082)
//   - STORAGE_TYPE: "postgres" or "memory" (default: memory)
//   - DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
func main() {
	ctx := context.Background()
	port := getEnvInt("USER_SERVICE_PORT", 8082)

	// Configure storage backend
	storageType := getEnv("STORAGE_TYPE", "memory")
	var store storage.Storage
	var err error

	switch storageType {
	case "postgres":
		// Build PostgreSQL DSN
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("DB_USER", "iot_user"),
			getEnv("DB_PASSWORD", "iot_password"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "iot_platform"),
			getEnv("DB_SSLMODE", "disable"),
		)
		store, err = storage.NewPostgresStorage(ctx, dsn)
		if err != nil {
			log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				log.Printf("⚠️  Error closing storage: %v", err)
			}
		}()
		log.Printf("✅ Using PostgreSQL storage")
	default:
		store = storage.NewMemoryStorage()
		log.Printf("✅ Using in-memory storage")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("❌ Failed to create listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	userServer := NewUserServer(store)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	log.Println("=====================================")
	log.Printf("User Service")
	log.Println("=====================================")
	log.Printf("Protocol: gRPC (HTTP/2)")
	log.Printf("Port: %d", port)
	log.Printf("Storage: %s", storageType)
	log.Printf("Address: http://localhost:%d", port)
	log.Println("-------------------------------------")
	log.Printf("✅ Server started")
	log.Printf("⏳ Waiting for gRPC connections...")
	log.Println("=====================================")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as int or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
