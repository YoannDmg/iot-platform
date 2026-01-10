// +build integration

package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	pb "github.com/yourusername/iot-platform/shared/proto/user"
)

func init() {
	// Load .env from project root (2 levels up from storage/)
	envPath := filepath.Join("..", "..", "..", ".env")
	_ = godotenv.Load(envPath) // Ignore error if .env doesn't exist
}

// setupPostgresStorage creates a PostgreSQL storage for testing.
// Requires PostgreSQL to be running (via docker-compose).
func setupPostgresStorage(t *testing.T) *PostgresStorage {
	t.Helper()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnvOrDefault("DB_USER", "iot_user"),
		getEnvOrDefault("DB_PASSWORD", "iot_password"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		getEnvOrDefault("DB_NAME", "iot_platform"),
		getEnvOrDefault("DB_SSLMODE", "disable"),
	)

	store, err := NewPostgresStorage(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v\nMake sure PostgreSQL is running: make up && make db-migrate", err)
	}

	// Clean up function
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	})

	return store
}

// cleanDatabase removes all users from the database before each test.
func cleanDatabase(t *testing.T, store Storage) {
	t.Helper()

	ctx := context.Background()
	users, _, err := store.ListUsers(ctx, 1, 1000, "")
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	for _, user := range users {
		if err := store.DeleteUser(ctx, user.Id); err != nil {
			t.Logf("Warning: Failed to delete user %s: %v", user.Id, err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestPostgresStorage_CreateUser(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     "postgres@example.com",
		Name:      "PostgreSQL Test User",
		Role:      "user",
		CreatedAt: 1704067200, // 2024-01-01
		IsActive:  true,
	}

	created, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	if created.Id != user.Id {
		t.Errorf("Id = %v, want %v", created.Id, user.Id)
	}
	if created.Email != user.Email {
		t.Errorf("Email = %v, want %v", created.Email, user.Email)
	}
	if !created.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestPostgresStorage_GetUser(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	userID := uuid.New().String()
	user := &pb.User{
		Id:        userID,
		Email:     "get@example.com",
		Name:      "Get Test User",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get existing user
	retrieved, err := store.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if retrieved.Id != user.Id {
		t.Errorf("Id = %v, want %v", retrieved.Id, user.Id)
	}
	if retrieved.Email != user.Email {
		t.Errorf("Email = %v, want %v", retrieved.Email, user.Email)
	}

	// Get non-existent user
	_, err = store.GetUser(ctx, uuid.New().String())
	if err == nil {
		t.Error("GetUser() should fail for non-existent user")
	}
}

func TestPostgresStorage_GetUserByEmail(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     "email-lookup@example.com",
		Name:      "Email Lookup Test",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get by email
	retrieved, err := store.GetUserByEmail(ctx, "email-lookup@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() failed: %v", err)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Email = %v, want %v", retrieved.Email, user.Email)
	}

	// Get non-existent email
	_, err = store.GetUserByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("GetUserByEmail() should fail for non-existent email")
	}
}

func TestPostgresStorage_GetPasswordHash(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	password := "SecurePassword123!"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     "hash@example.com",
		Name:      "Hash Test User",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get password hash
	retrievedHash, err := store.GetPasswordHash(ctx, "hash@example.com")
	if err != nil {
		t.Fatalf("GetPasswordHash() failed: %v", err)
	}

	// Verify hash works
	err = bcrypt.CompareHashAndPassword([]byte(retrievedHash), []byte(password))
	if err != nil {
		t.Error("Password hash verification failed")
	}

	// Wrong password should fail
	err = bcrypt.CompareHashAndPassword([]byte(retrievedHash), []byte("WrongPassword"))
	if err == nil {
		t.Error("Wrong password should not match")
	}
}

func TestPostgresStorage_UpdateUser(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     "update@example.com",
		Name:      "Original Name",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Update user
	user.Name = "Updated Name PostgreSQL"
	user.Role = "admin"
	user.IsActive = false

	updated, err := store.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("UpdateUser() failed: %v", err)
	}

	if updated.Name != "Updated Name PostgreSQL" {
		t.Errorf("Name = %v, want 'Updated Name PostgreSQL'", updated.Name)
	}
	if updated.Role != "admin" {
		t.Errorf("Role = %v, want 'admin'", updated.Role)
	}
	if updated.IsActive {
		t.Error("IsActive should be false")
	}

	// Email should remain unchanged
	if updated.Email != "update@example.com" {
		t.Errorf("Email changed: %v", updated.Email)
	}
}

func TestPostgresStorage_UpdateLastLogin(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	userID := uuid.New().String()
	user := &pb.User{
		Id:        userID,
		Email:     "login@example.com",
		Name:      "Login Test User",
		Role:      "user",
		CreatedAt: 1704067200,
		LastLogin: 0,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Update last login
	err = store.UpdateLastLogin(ctx, userID)
	if err != nil {
		t.Fatalf("UpdateLastLogin() failed: %v", err)
	}

	// Verify last login was updated
	retrieved, err := store.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if retrieved.LastLogin == 0 {
		t.Error("LastLogin should be updated to non-zero value")
	}
}

func TestPostgresStorage_DeleteUser(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	userID := uuid.New().String()
	user := &pb.User{
		Id:        userID,
		Email:     "delete@example.com",
		Name:      "Delete Test User",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err := store.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Delete user
	err = store.DeleteUser(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteUser() failed: %v", err)
	}

	// Verify user is deleted
	_, err = store.GetUser(ctx, userID)
	if err == nil {
		t.Error("GetUser() should fail after deletion")
	}

	// Verify cannot get by email
	_, err = store.GetUserByEmail(ctx, "delete@example.com")
	if err == nil {
		t.Error("GetUserByEmail() should fail after deletion")
	}
}

func TestPostgresStorage_ListUsers(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := &pb.User{
			Id:        uuid.New().String(),
			Email:     fmt.Sprintf("list%d@example.com", i),
			Name:      fmt.Sprintf("List User %d", i),
			Role:      "user",
			CreatedAt: 1704067200,
			IsActive:  true,
		}
		_, err := store.CreateUser(ctx, user, string(passwordHash))
		if err != nil {
			t.Fatalf("CreateUser(%d) failed: %v", i, err)
		}
	}

	// List all users
	users, total, err := store.ListUsers(ctx, 1, 10, "")
	if err != nil {
		t.Fatalf("ListUsers() failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Total = %d, want 5", total)
	}
	if len(users) != 5 {
		t.Errorf("Users count = %d, want 5", len(users))
	}
}

func TestPostgresStorage_ListUsersByRole(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	
	// Create users with different roles
	roles := []string{"user", "admin", "user", "device", "user"}
	for i, role := range roles {
		user := &pb.User{
			Id:        uuid.New().String(),
			Email:     fmt.Sprintf("role%d@example.com", i),
			Name:      fmt.Sprintf("Role User %d", i),
			Role:      role,
			CreatedAt: 1704067200,
			IsActive:  true,
		}
		_, err := store.CreateUser(ctx, user, string(passwordHash))
		if err != nil {
			t.Fatalf("CreateUser(%d) failed: %v", i, err)
		}
	}

	// List user role
	users, total, err := store.ListUsers(ctx, 1, 10, "user")
	if err != nil {
		t.Fatalf("ListUsers(role=user) failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Total users with role 'user' = %d, want 3", total)
	}

	if len(users) != 3 {
		t.Errorf("Users count = %d, want 3", len(users))
	}

	// Verify all returned users have role 'user'
	for _, user := range users {
		if user.Role != "user" {
			t.Errorf("User role = %v, want 'user'", user.Role)
		}
	}
}

func TestPostgresStorage_AuthenticationFlow(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	// Register user with hashed password
	plainPassword := "MySecurePassword123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &pb.User{
		Id:        uuid.New().String(),
		Email:     "auth@example.com",
		Name:      "Auth Test User",
		Role:      "user",
		CreatedAt: 1704067200,
		IsActive:  true,
	}

	_, err = store.CreateUser(ctx, user, string(hashedPassword))
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Authenticate with correct password
	storedHash, err := store.GetPasswordHash(ctx, "auth@example.com")
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
