package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto/user"
	"github.com/yourusername/iot-platform/services/user-service/db/sqlc"
)

// PostgresStorage implements Storage interface using PostgreSQL with pgx.
type PostgresStorage struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewPostgresStorage creates a new PostgreSQL storage instance.
// dsn format: "postgres://user:pass@host:port/dbname?sslmode=disable"
func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{
		pool:    pool,
		queries: sqlc.New(pool),
	}, nil
}

// CreateUser implements Storage.CreateUser.
func (s *PostgresStorage) CreateUser(ctx context.Context, user *pb.User, passwordHash string) (*pb.User, error) {
	// Parse UUID string to pgtype.UUID
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(user.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Create timestamps
	createdAt := pgtype.Timestamptz{}
	if err := createdAt.Scan(time.Unix(user.CreatedAt, 0)); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	// Insert user
	dbUser, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgUUID,
		Email:        user.Email,
		PasswordHash: passwordHash,
		Name:         user.Name,
		Role:         user.Role,
		CreatedAt:    createdAt,
		IsActive:     user.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return dbUserToProto(dbUser, passwordHash)
}

// GetUser implements Storage.GetUser.
func (s *PostgresStorage) GetUser(ctx context.Context, id string) (*pb.User, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	dbUser, err := s.queries.GetUser(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user %s not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get password hash separately
	passwordHash, err := s.queries.GetPasswordHash(ctx, dbUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get password hash: %w", err)
	}

	return dbUserToProto(dbUser, passwordHash)
}

// GetUserByEmail implements Storage.GetUserByEmail.
func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return dbUserToProto(dbUser, dbUser.PasswordHash)
}

// GetPasswordHash implements Storage.GetPasswordHash.
func (s *PostgresStorage) GetPasswordHash(ctx context.Context, email string) (string, error) {
	passwordHash, err := s.queries.GetPasswordHash(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", status.Errorf(codes.NotFound, "user with email %s not found", email)
		}
		return "", fmt.Errorf("failed to get password hash: %w", err)
	}

	return passwordHash, nil
}

// UpdateUser implements Storage.UpdateUser.
func (s *PostgresStorage) UpdateUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(user.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	dbUser, err := s.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       pgUUID,
		Name:     user.Name,
		Role:     user.Role,
		IsActive: user.IsActive,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user %s not found", user.Id)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return dbUserToProto(dbUser, dbUser.PasswordHash)
}

// UpdateLastLogin implements Storage.UpdateLastLogin.
func (s *PostgresStorage) UpdateLastLogin(ctx context.Context, id string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	lastLogin := pgtype.Timestamptz{}
	if err := lastLogin.Scan(time.Now()); err != nil {
		return fmt.Errorf("failed to create timestamp: %w", err)
	}

	err := s.queries.UpdateLastLogin(ctx, sqlc.UpdateLastLoginParams{
		ID:        pgUUID,
		LastLogin: lastLogin,
	})
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// DeleteUser implements Storage.DeleteUser.
func (s *PostgresStorage) DeleteUser(ctx context.Context, id string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	err := s.queries.DeleteUser(ctx, pgUUID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers implements Storage.ListUsers.
func (s *PostgresStorage) ListUsers(ctx context.Context, page, pageSize int32, roleFilter string) ([]*pb.User, int32, error) {
	var dbUsers []sqlc.User
	var err error

	limit := pageSize
	offset := (page - 1) * pageSize

	if roleFilter != "" {
		// List by role
		dbUsers, err = s.queries.ListUsersByRole(ctx, sqlc.ListUsersByRoleParams{
			Role:   roleFilter,
			Limit:  limit,
			Offset: offset,
		})
	} else {
		// List all
		dbUsers, err = s.queries.ListUsers(ctx, sqlc.ListUsersParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count
	total, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Convert to proto
	users := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i], err = dbUserToProto(dbUser, dbUser.PasswordHash)
		if err != nil {
			return nil, 0, err
		}
	}

	return users, int32(total), nil
}

// Close closes the database connection pool.
func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

// Helper function to convert sqlc.User to pb.User
func dbUserToProto(dbUser sqlc.User, passwordHash string) (*pb.User, error) {
	user := &pb.User{
		Id:       dbUser.ID.String(),
		Email:    dbUser.Email,
		Name:     dbUser.Name,
		Role:     dbUser.Role,
		IsActive: dbUser.IsActive,
	}

	// Convert timestamps
	if dbUser.CreatedAt.Valid {
		user.CreatedAt = dbUser.CreatedAt.Time.Unix()
	}

	if dbUser.LastLogin.Valid {
		user.LastLogin = dbUser.LastLogin.Time.Unix()
	}

	return user, nil
}
