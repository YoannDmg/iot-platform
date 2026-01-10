// Package storage provides in-memory implementation for user persistence.
package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MemoryStorage implements Storage interface using an in-memory map.
// Thread-safe via mutex. For development/testing only.
type MemoryStorage struct {
	users         map[string]*pb.User
	passwordHash  map[string]string // email -> password hash
	emailToID     map[string]string // email -> user ID
	mu            sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:        make(map[string]*pb.User),
		passwordHash: make(map[string]string),
		emailToID:    make(map[string]string),
	}
}

// CreateUser stores a new user in memory.
func (m *MemoryStorage) CreateUser(ctx context.Context, user *pb.User, passwordHash string) (*pb.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if email already exists
	if _, exists := m.emailToID[user.Email]; exists {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}

	// Store user
	m.users[user.Id] = user
	m.passwordHash[user.Email] = passwordHash
	m.emailToID[user.Email] = user.Id

	return user, nil
}

// GetUser retrieves a user by ID.
func (m *MemoryStorage) GetUser(ctx context.Context, id string) (*pb.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (m *MemoryStorage) GetUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID, exists := m.emailToID[email]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return m.users[userID], nil
}

// GetPasswordHash retrieves the password hash for a user by email.
func (m *MemoryStorage) GetPasswordHash(ctx context.Context, email string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, exists := m.passwordHash[email]
	if !exists {
		return "", status.Error(codes.NotFound, "user not found")
	}

	return hash, nil
}

// ListUsers returns a paginated list of users.
func (m *MemoryStorage) ListUsers(ctx context.Context, page, pageSize int32, role string) ([]*pb.User, int32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	allUsers := make([]*pb.User, 0, len(m.users))
	for _, user := range m.users {
		// Filter by role if specified
		if role != "" && user.Role != role {
			continue
		}
		allUsers = append(allUsers, user)
	}

	total := int32(len(allUsers))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*pb.User{}, total, nil
	}
	if end > total {
		end = total
	}

	return allUsers[start:end], total, nil
}

// UpdateUser updates an existing user.
func (m *MemoryStorage) UpdateUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.users[user.Id]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Update fields
	if user.Name != "" {
		existing.Name = user.Name
	}
	if user.Role != "" {
		existing.Role = user.Role
	}
	existing.IsActive = user.IsActive

	return existing, nil
}

// UpdateLastLogin updates the last_login timestamp for a user.
func (m *MemoryStorage) UpdateLastLogin(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userID]
	if !exists {
		return status.Error(codes.NotFound, "user not found")
	}

	user.LastLogin = time.Now().Unix()
	return nil
}

// DeleteUser removes a user by ID.
func (m *MemoryStorage) DeleteUser(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return status.Error(codes.NotFound, "user not found")
	}

	// Remove from all maps
	delete(m.users, id)
	delete(m.passwordHash, user.Email)
	delete(m.emailToID, user.Email)

	return nil
}

// Close releases resources (no-op for memory storage).
func (m *MemoryStorage) Close() error {
	return nil
}

// Debug helper to print storage state
func (m *MemoryStorage) Debug() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fmt.Printf("MemoryStorage: %d users\n", len(m.users))
}
