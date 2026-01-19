// Package bootstrap handles initial system setup tasks.
package bootstrap

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
	"github.com/yourusername/iot-platform/services/user-service/storage"
)

// Default admin credentials (override via environment variables)
const (
	defaultAdminEmail    = "admin@iot.local"
	defaultAdminPassword = "admin"
	defaultAdminName     = "Admin"
)

// EnsureAdminExists checks if an admin user exists and creates one if not.
// Credentials are read from environment variables:
//   - ADMIN_EMAIL (default: admin@iot.local)
//   - ADMIN_PASSWORD (default: admin)
//   - ADMIN_NAME (default: Admin)
func EnsureAdminExists(ctx context.Context, store storage.Storage) error {
	email := getEnv("ADMIN_EMAIL", defaultAdminEmail)
	password := getEnv("ADMIN_PASSWORD", defaultAdminPassword)
	name := getEnv("ADMIN_NAME", defaultAdminName)

	// Check if admin already exists
	existingUser, err := store.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		log.Printf("✅ Admin user already exists: %s", email)
		return nil
	}

	// Check if any admin exists
	users, _, err := store.ListUsers(ctx, 1, 100, "admin")
	if err == nil && len(users) > 0 {
		log.Printf("✅ Admin user(s) already exist, skipping bootstrap")
		return nil
	}

	// Create admin user
	log.Printf("🔧 Creating initial admin user...")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Failed to hash admin password: %v", err)
		return err
	}

	admin := &pb.User{
		Id:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Role:      "admin",
		CreatedAt: time.Now().Unix(),
		IsActive:  true,
	}

	_, err = store.CreateUser(ctx, admin, string(passwordHash))
	if err != nil {
		log.Printf("❌ Failed to create admin user: %v", err)
		return err
	}

	log.Println("=====================================")
	log.Println("🔐 INITIAL ADMIN CREATED")
	log.Println("=====================================")
	log.Printf("   Email:    %s", email)
	log.Printf("   Password: %s", password)
	log.Println("=====================================")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
