// +build unit

package storage

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
	userpb "github.com/yourusername/iot-platform/shared/proto/user"
)

func TestMemoryStorage_CreateUser(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}
	passwordHash := "hashed-password"

	createdUser, err := storage.CreateUser(ctx, user, passwordHash)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	if createdUser.Id != user.Id {
		t.Errorf("Id = %v, want %v", createdUser.Id, user.Id)
	}
	if createdUser.Email != user.Email {
		t.Errorf("Email = %v, want %v", createdUser.Email, user.Email)
	}
}

func TestMemoryStorage_CreateUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user1 := &userpb.User{
		Id:       "user-1",
		Email:    "test@example.com",
		Name:     "User 1",
		Role:     "user",
		IsActive: true,
	}

	_, err := storage.CreateUser(ctx, user1, "hash1")
	if err != nil {
		t.Fatalf("First CreateUser() failed: %v", err)
	}

	user2 := &userpb.User{
		Id:       "user-2",
		Email:    "test@example.com", // Same email
		Name:     "User 2",
		Role:     "user",
		IsActive: true,
	}

	_, err = storage.CreateUser(ctx, user2, "hash2")
	if err == nil {
		t.Error("CreateUser() should fail with duplicate email")
	}
}

func TestMemoryStorage_GetUser(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	_, err := storage.CreateUser(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get existing user
	retrievedUser, err := storage.GetUser(ctx, "user-123")
	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if retrievedUser.Id != user.Id {
		t.Errorf("Id = %v, want %v", retrievedUser.Id, user.Id)
	}
	if retrievedUser.Email != user.Email {
		t.Errorf("Email = %v, want %v", retrievedUser.Email, user.Email)
	}

	// Get non-existent user
	_, err = storage.GetUser(ctx, "non-existent")
	if err == nil {
		t.Error("GetUser() should fail for non-existent user")
	}
}

func TestMemoryStorage_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	_, err := storage.CreateUser(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get existing user by email
	retrievedUser, err := storage.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() failed: %v", err)
	}

	if retrievedUser.Id != user.Id {
		t.Errorf("Id = %v, want %v", retrievedUser.Id, user.Id)
	}

	// Get non-existent user by email
	_, err = storage.GetUserByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("GetUserByEmail() should fail for non-existent email")
	}
}

func TestMemoryStorage_GetPasswordHash(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}
	expectedHash := "hashed-password-123"

	_, err := storage.CreateUser(ctx, user, expectedHash)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get password hash
	hash, err := storage.GetPasswordHash(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetPasswordHash() failed: %v", err)
	}

	if hash != expectedHash {
		t.Errorf("Password hash = %v, want %v", hash, expectedHash)
	}

	// Get password hash for non-existent user
	_, err = storage.GetPasswordHash(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("GetPasswordHash() should fail for non-existent email")
	}
}

func TestMemoryStorage_AuthenticationFlow(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	// Register user with bcrypt hashed password
	plainPassword := "MySecurePassword123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	_, err = storage.CreateUser(ctx, user, string(hashedPassword))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Authenticate with correct password
	storedHash, err := storage.GetPasswordHash(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetPasswordHash() failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	if err != nil {
		t.Error("Authentication should succeed with correct password")
	}

	// Authenticate with wrong password
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("WrongPassword"))
	if err == nil {
		t.Error("Authentication should fail with wrong password")
	}
}

func TestMemoryStorage_UpdateUser(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	_, err := storage.CreateUser(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Update user
	updatedUser := &userpb.User{
		Id:       "user-123",
		Name:     "Updated Name",
		Role:     "admin",
		IsActive: false,
	}

	result, err := storage.UpdateUser(ctx, updatedUser)
	if err != nil {
		t.Fatalf("UpdateUser() failed: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Errorf("Name = %v, want 'Updated Name'", result.Name)
	}
	if result.Role != "admin" {
		t.Errorf("Role = %v, want 'admin'", result.Role)
	}
	if result.IsActive {
		t.Error("IsActive should be false")
	}

	// Verify original email is unchanged
	if result.Email != "test@example.com" {
		t.Errorf("Email should remain unchanged, got %v", result.Email)
	}
}

func TestMemoryStorage_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      "user",
		IsActive:  true,
		LastLogin: 0,
	}

	_, err := storage.CreateUser(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Update last login
	err = storage.UpdateLastLogin(ctx, "user-123")
	if err != nil {
		t.Fatalf("UpdateLastLogin() failed: %v", err)
	}

	// Verify last login was updated
	updatedUser, err := storage.GetUser(ctx, "user-123")
	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if updatedUser.LastLogin == 0 {
		t.Error("LastLogin should be updated to non-zero value")
	}
}

func TestMemoryStorage_DeleteUser(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	user := &userpb.User{
		Id:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	_, err := storage.CreateUser(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Delete user
	err = storage.DeleteUser(ctx, "user-123")
	if err != nil {
		t.Fatalf("DeleteUser() failed: %v", err)
	}

	// Verify user is deleted
	_, err = storage.GetUser(ctx, "user-123")
	if err == nil {
		t.Error("GetUser() should fail after deletion")
	}

	// Verify email mapping is deleted
	_, err = storage.GetUserByEmail(ctx, "test@example.com")
	if err == nil {
		t.Error("GetUserByEmail() should fail after deletion")
	}

	// Verify password hash is deleted
	_, err = storage.GetPasswordHash(ctx, "test@example.com")
	if err == nil {
		t.Error("GetPasswordHash() should fail after deletion")
	}
}

func TestMemoryStorage_ListUsers(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	// Create multiple users
	users := []*userpb.User{
		{Id: "user-1", Email: "user1@example.com", Name: "User 1", Role: "user", IsActive: true},
		{Id: "user-2", Email: "user2@example.com", Name: "User 2", Role: "admin", IsActive: true},
		{Id: "user-3", Email: "user3@example.com", Name: "User 3", Role: "user", IsActive: true},
	}

	for _, user := range users {
		_, err := storage.CreateUser(ctx, user, "hash")
		if err != nil {
			t.Fatalf("CreateUser() failed: %v", err)
		}
	}

	// List all users
	result, total, err := storage.ListUsers(ctx, 1, 10, "")
	if err != nil {
		t.Fatalf("ListUsers() failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Total = %d, want 3", total)
	}
	if len(result) != 3 {
		t.Errorf("Result length = %d, want 3", len(result))
	}

	// List users with role filter
	result, total, err = storage.ListUsers(ctx, 1, 10, "user")
	if err != nil {
		t.Fatalf("ListUsers() with filter failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Total with filter = %d, want 2", total)
	}
}
