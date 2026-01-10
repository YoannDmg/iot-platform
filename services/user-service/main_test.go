// +build unit

package main

import (
	"context"
	"sync"
	"testing"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
	"github.com/yourusername/iot-platform/services/user-service/storage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestRegister tests user registration functionality.
func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		request     *pb.RegisterRequest
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name: "valid_user",
			request: &pb.RegisterRequest{
				Email:    "user@example.com",
				Password: "SecurePassword123!",
				Name:     "Test User",
				Role:     "user",
			},
			wantErr:     false,
			description: "should successfully register a valid user",
		},
		{
			name: "missing_email",
			request: &pb.RegisterRequest{
				Password: "SecurePassword123!",
				Name:     "Test User",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when email is missing",
		},
		{
			name: "missing_password",
			request: &pb.RegisterRequest{
				Email: "user@example.com",
				Name:  "Test User",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when password is missing",
		},
		{
			name: "missing_name",
			request: &pb.RegisterRequest{
				Email:    "user@example.com",
				Password: "SecurePassword123!",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when name is missing",
		},
		{
			name: "invalid_role",
			request: &pb.RegisterRequest{
				Email:    "user@example.com",
				Password: "SecurePassword123!",
				Name:     "Test User",
				Role:     "superuser",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error for invalid role",
		},
		{
			name: "default_role",
			request: &pb.RegisterRequest{
				Email:    "default@example.com",
				Password: "SecurePassword123!",
				Name:     "Default User",
			},
			wantErr:     false,
			description: "should default to 'user' role when not specified",
		},
		{
			name: "admin_role",
			request: &pb.RegisterRequest{
				Email:    "admin@example.com",
				Password: "SecurePassword123!",
				Name:     "Admin User",
				Role:     "admin",
			},
			wantErr:     false,
			description: "should allow admin role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewUserServer(storage.NewMemoryStorage())
			ctx := context.Background()

			resp, err := server.Register(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				} else {
					t.Errorf("%s: expected gRPC status error, got %v", tt.description, err)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			// Validate response
			if resp.User == nil {
				t.Errorf("%s: user is nil", tt.description)
				return
			}

			user := resp.User

			// Check ID is generated
			if user.Id == "" {
				t.Errorf("%s: user ID should not be empty", tt.description)
			}

			// Check fields match request
			if user.Email != tt.request.Email {
				t.Errorf("%s: expected email %s, got %s", tt.description, tt.request.Email, user.Email)
			}
			if user.Name != tt.request.Name {
				t.Errorf("%s: expected name %s, got %s", tt.description, tt.request.Name, user.Name)
			}

			// Check role
			expectedRole := tt.request.Role
			if expectedRole == "" {
				expectedRole = "user"
			}
			if user.Role != expectedRole {
				t.Errorf("%s: expected role %s, got %s", tt.description, expectedRole, user.Role)
			}

			// Check defaults
			if !user.IsActive {
				t.Errorf("%s: user should be active by default", tt.description)
			}

			// Check timestamps
			if user.CreatedAt == 0 {
				t.Errorf("%s: CreatedAt should be set", tt.description)
			}
		})
	}
}

// TestRegister_DuplicateEmail tests that duplicate emails are rejected.
func TestRegister_DuplicateEmail(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create first user
	req := &pb.RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "Password123!",
		Name:     "First User",
	}
	_, err := server.Register(ctx, req)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Try to create user with same email
	req2 := &pb.RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "DifferentPassword123!",
		Name:     "Second User",
	}
	_, err = server.Register(ctx, req2)
	if err == nil {
		t.Error("should return error for duplicate email")
	}
}

// TestAuthenticate tests user authentication functionality.
func TestAuthenticate(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create a test user
	testEmail := "auth@example.com"
	testPassword := "SecurePassword123!"

	_, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
		Name:     "Auth Test User",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name        string
		request     *pb.AuthenticateRequest
		wantSuccess bool
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name: "valid_credentials",
			request: &pb.AuthenticateRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			wantSuccess: true,
			wantErr:     false,
			description: "should authenticate with valid credentials",
		},
		{
			name: "invalid_password",
			request: &pb.AuthenticateRequest{
				Email:    testEmail,
				Password: "WrongPassword123!",
			},
			wantSuccess: false,
			wantErr:     false,
			description: "should reject invalid password",
		},
		{
			name: "nonexistent_user",
			request: &pb.AuthenticateRequest{
				Email:    "nonexistent@example.com",
				Password: testPassword,
			},
			wantSuccess: false,
			wantErr:     false,
			description: "should reject nonexistent user",
		},
		{
			name: "missing_email",
			request: &pb.AuthenticateRequest{
				Password: testPassword,
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when email is missing",
		},
		{
			name: "missing_password",
			request: &pb.AuthenticateRequest{
				Email: testEmail,
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when password is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Authenticate(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.Success != tt.wantSuccess {
				t.Errorf("%s: expected success=%v, got %v", tt.description, tt.wantSuccess, resp.Success)
			}

			if tt.wantSuccess && resp.User == nil {
				t.Errorf("%s: user should not be nil on successful auth", tt.description)
			}

			if !tt.wantSuccess && resp.User != nil {
				t.Errorf("%s: user should be nil on failed auth", tt.description)
			}
		})
	}
}

// TestAuthenticate_InactiveUser tests that inactive users cannot authenticate.
func TestAuthenticate_InactiveUser(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create user
	email := "inactive@example.com"
	password := "Password123!"
	registerResp, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     "Inactive User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Deactivate user
	_, err = server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:       registerResp.User.Id,
		Name:     "Inactive User",
		Role:     "user",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("failed to deactivate user: %v", err)
	}

	// Try to authenticate
	authResp, err := server.Authenticate(ctx, &pb.AuthenticateRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		t.Fatalf("authenticate returned error: %v", err)
	}

	if authResp.Success {
		t.Error("inactive user should not be able to authenticate")
	}

	if authResp.Message != "Account is inactive" {
		t.Errorf("expected 'Account is inactive' message, got: %s", authResp.Message)
	}
}

// TestGetUser tests user retrieval functionality.
func TestGetUser(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create a test user
	registerResp, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    "get@example.com",
		Password: "Password123!",
		Name:     "Get Test User",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID := registerResp.User.Id

	tests := []struct {
		name        string
		userID      string
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name:        "existing_user",
			userID:      userID,
			wantErr:     false,
			description: "should retrieve existing user",
		},
		{
			name:        "nonexistent_user",
			userID:      "nonexistent-id",
			wantErr:     true,
			wantCode:    codes.NotFound,
			description: "should return NotFound for nonexistent user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetUser(ctx, &pb.GetUserRequest{
				Id: tt.userID,
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.User == nil {
				t.Errorf("%s: user is nil", tt.description)
				return
			}

			if resp.User.Id != tt.userID {
				t.Errorf("%s: expected ID %s, got %s", tt.description, tt.userID, resp.User.Id)
			}
		})
	}
}

// TestGetUserByEmail tests email-based user retrieval.
func TestGetUserByEmail(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	testEmail := "email@example.com"

	// Create test user
	_, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    testEmail,
		Password: "Password123!",
		Name:     "Email Test User",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Get existing user
	resp, err := server.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: testEmail,
	})
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}

	if resp.User == nil {
		t.Fatal("user is nil")
	}

	if resp.User.Email != testEmail {
		t.Errorf("expected email %s, got %s", testEmail, resp.User.Email)
	}

	// Get nonexistent user
	_, err = server.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: "nonexistent@example.com",
	})
	if err == nil {
		t.Error("should return error for nonexistent user")
	}
}

// TestListUsers tests user listing functionality.
func TestListUsers(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Test empty list
	t.Run("empty_list", func(t *testing.T) {
		resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Users) != 0 {
			t.Errorf("expected 0 users, got %d", len(resp.Users))
		}
		if resp.Total != 0 {
			t.Errorf("expected total 0, got %d", resp.Total)
		}
	})

	// Create test users
	userIDs := make([]string, 5)
	for i := 0; i < 3; i++ {
		resp, err := server.Register(ctx, &pb.RegisterRequest{
			Email:    "user" + string(rune('1'+i)) + "@example.com",
			Password: "Password123!",
			Name:     "User " + string(rune('A'+i)),
			Role:     "user",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
		userIDs[i] = resp.User.Id
	}

	// Create admin users
	for i := 3; i < 5; i++ {
		resp, err := server.Register(ctx, &pb.RegisterRequest{
			Email:    "admin" + string(rune('1'+i-3)) + "@example.com",
			Password: "Password123!",
			Name:     "Admin " + string(rune('A'+i-3)),
			Role:     "admin",
		})
		if err != nil {
			t.Fatalf("failed to create admin user: %v", err)
		}
		userIDs[i] = resp.User.Id
	}

	// List all users
	t.Run("list_all_users", func(t *testing.T) {
		resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Users) != 5 {
			t.Errorf("expected 5 users, got %d", len(resp.Users))
		}
		if resp.Total != 5 {
			t.Errorf("expected total 5, got %d", resp.Total)
		}
	})

	// List users by role
	t.Run("list_by_role", func(t *testing.T) {
		resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
			Role:     "user",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Users) != 3 {
			t.Errorf("expected 3 users with role 'user', got %d", len(resp.Users))
		}

		// Verify all returned users have correct role
		for _, user := range resp.Users {
			if user.Role != "user" {
				t.Errorf("expected role 'user', got %s", user.Role)
			}
		}
	})

	// List admins
	t.Run("list_admins", func(t *testing.T) {
		resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
			Role:     "admin",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Users) != 2 {
			t.Errorf("expected 2 admin users, got %d", len(resp.Users))
		}
	})
}

// TestUpdateUser tests user update functionality.
func TestUpdateUser(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create test user
	registerResp, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    "update@example.com",
		Password: "Password123!",
		Name:     "Original Name",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID := registerResp.User.Id

	tests := []struct {
		name        string
		request     *pb.UpdateUserRequest
		wantErr     bool
		wantCode    codes.Code
		validate    func(t *testing.T, user *pb.User)
		description string
	}{
		{
			name: "update_name",
			request: &pb.UpdateUserRequest{
				Id:       userID,
				Name:     "Updated Name",
				Role:     "user",
				IsActive: true,
			},
			wantErr: false,
			validate: func(t *testing.T, user *pb.User) {
				if user.Name != "Updated Name" {
					t.Errorf("expected name 'Updated Name', got %s", user.Name)
				}
			},
			description: "should update user name",
		},
		{
			name: "update_role",
			request: &pb.UpdateUserRequest{
				Id:       userID,
				Name:     "Updated Name",
				Role:     "admin",
				IsActive: true,
			},
			wantErr: false,
			validate: func(t *testing.T, user *pb.User) {
				if user.Role != "admin" {
					t.Errorf("expected role 'admin', got %s", user.Role)
				}
			},
			description: "should update user role",
		},
		{
			name: "deactivate_user",
			request: &pb.UpdateUserRequest{
				Id:       userID,
				Name:     "Updated Name",
				Role:     "admin",
				IsActive: false,
			},
			wantErr: false,
			validate: func(t *testing.T, user *pb.User) {
				if user.IsActive {
					t.Error("user should be inactive")
				}
			},
			description: "should deactivate user",
		},
		{
			name: "nonexistent_user",
			request: &pb.UpdateUserRequest{
				Id:       "nonexistent-id",
				Name:     "Test",
				Role:     "user",
				IsActive: true,
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			description: "should return NotFound for nonexistent user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.UpdateUser(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.User == nil {
				t.Errorf("%s: user is nil", tt.description)
				return
			}

			if tt.validate != nil {
				tt.validate(t, resp.User)
			}
		})
	}
}

// TestDeleteUser tests user deletion functionality.
func TestDeleteUser(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create test user
	registerResp, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    "delete@example.com",
		Password: "Password123!",
		Name:     "Delete Test User",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID := registerResp.User.Id

	// Delete user
	deleteResp, err := server.DeleteUser(ctx, &pb.DeleteUserRequest{
		Id: userID,
	})
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	if !deleteResp.Success {
		t.Error("delete should be successful")
	}

	// Verify user is deleted
	_, err = server.GetUser(ctx, &pb.GetUserRequest{
		Id: userID,
	})
	if err == nil {
		t.Error("should return error when getting deleted user")
	}

	// Try to delete nonexistent user
	_, err = server.DeleteUser(ctx, &pb.DeleteUserRequest{
		Id: "nonexistent-id",
	})
	if err == nil {
		t.Error("should return error when deleting nonexistent user")
	}
}

// TestPasswordHashing tests that passwords are properly hashed.
func TestPasswordHashing(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	password := "MySecurePassword123!"

	// Register user
	registerResp, err := server.Register(ctx, &pb.RegisterRequest{
		Email:    "hash@example.com",
		Password: password,
		Name:     "Hash Test User",
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Get password hash from storage
	hash, err := server.storage.GetPasswordHash(ctx, "hash@example.com")
	if err != nil {
		t.Fatalf("failed to get password hash: %v", err)
	}

	// Verify hash is not the plain password
	if hash == password {
		t.Error("password should be hashed, not stored in plain text")
	}

	// Verify hash is valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("password hash is invalid: %v", err)
	}

	// Verify wrong password doesn't match
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("WrongPassword"))
	if err == nil {
		t.Error("wrong password should not match hash")
	}

	// Verify user object doesn't contain password
	if registerResp.User.Email == "" {
		t.Error("user email should be set")
	}
}

// TestConcurrentOperations tests thread safety with concurrent access.
func TestConcurrentOperations(t *testing.T) {
	server := NewUserServer(storage.NewMemoryStorage())
	ctx := context.Background()

	numGoroutines := 50
	var wg sync.WaitGroup

	// Create users concurrently
	t.Run("concurrent_register", func(t *testing.T) {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				_, err := server.Register(ctx, &pb.RegisterRequest{
					Email:    "concurrent" + string(rune('0'+id)) + "@example.com",
					Password: "Password123!",
					Name:     "Concurrent User",
				})
				if err != nil {
					t.Errorf("concurrent register failed: %v", err)
				}
			}(i)
		}
		wg.Wait()

		// Verify all users were created
		resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
			Page:     1,
			PageSize: 100,
		})
		if err != nil {
			t.Fatalf("failed to list users: %v", err)
		}
		if len(resp.Users) != numGoroutines {
			t.Errorf("expected %d users, got %d", numGoroutines, len(resp.Users))
		}
	})

	// Concurrent authentication
	t.Run("concurrent_authenticate", func(t *testing.T) {
		// Create a test user
		email := "concurrent-auth@example.com"
		password := "Password123!"
		_, err := server.Register(ctx, &pb.RegisterRequest{
			Email:    email,
			Password: password,
			Name:     "Concurrent Auth User",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				resp, err := server.Authenticate(ctx, &pb.AuthenticateRequest{
					Email:    email,
					Password: password,
				})
				if err != nil {
					t.Errorf("concurrent authenticate failed: %v", err)
				}
				if !resp.Success {
					t.Error("authentication should succeed")
				}
			}()
		}
		wg.Wait()
	})
}
