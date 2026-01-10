// Package storage provides an abstraction layer for user persistence.
// Supports multiple backends (PostgreSQL, in-memory) via a common interface.
package storage

import (
	"context"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
)

// Storage defines the interface for user persistence operations.
// Implementations: PostgresStorage (production), MemoryStorage (tests/dev).
type Storage interface {
	// CreateUser stores a new user and returns the stored user.
	// Password should already be hashed before calling this method.
	CreateUser(ctx context.Context, user *pb.User, passwordHash string) (*pb.User, error)

	// GetUser retrieves a user by ID.
	// Returns nil, ErrNotFound if user doesn't exist.
	GetUser(ctx context.Context, id string) (*pb.User, error)

	// GetUserByEmail retrieves a user by email.
	// Returns nil, ErrNotFound if user doesn't exist.
	GetUserByEmail(ctx context.Context, email string) (*pb.User, error)

	// GetPasswordHash retrieves the password hash for a user by email.
	// Used for authentication.
	GetPasswordHash(ctx context.Context, email string) (string, error)

	// ListUsers returns a paginated list of users.
	// Returns users, total count, and error.
	ListUsers(ctx context.Context, page, pageSize int32, role string) ([]*pb.User, int32, error)

	// UpdateUser updates an existing user.
	// Only non-zero/non-nil fields are updated.
	// Returns updated user or ErrNotFound.
	UpdateUser(ctx context.Context, user *pb.User) (*pb.User, error)

	// UpdateLastLogin updates the last_login timestamp for a user.
	UpdateLastLogin(ctx context.Context, userID string) error

	// DeleteUser removes a user by ID.
	// Returns ErrNotFound if user doesn't exist.
	DeleteUser(ctx context.Context, id string) error

	// Close releases any resources held by the storage.
	Close() error
}
