package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
	"github.com/yourusername/iot-platform/services/device-manager/db/sqlc"
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

// CreateDevice implements Storage.CreateDevice.
func (s *PostgresStorage) CreateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error) {
	// Convert metadata map to JSONB
	metadataJSON, err := json.Marshal(device.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Parse UUID string to pgtype.UUID
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(device.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device ID: %v", err)
	}

	// Convert status
	dbStatus := protoStatusToDBStatus(device.Status)

	// Create timestamps
	createdAt := pgtype.Timestamptz{}
	if err := createdAt.Scan(time.Unix(device.CreatedAt, 0)); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	lastSeen := pgtype.Timestamptz{}
	if err := lastSeen.Scan(time.Unix(device.LastSeen, 0)); err != nil {
		return nil, fmt.Errorf("failed to parse last_seen: %w", err)
	}

	// Insert device
	dbDevice, err := s.queries.CreateDevice(ctx, sqlc.CreateDeviceParams{
		ID:        pgUUID,
		Name:      device.Name,
		Type:      device.Type,
		Status:    dbStatus,
		CreatedAt: createdAt,
		LastSeen:  lastSeen,
		Metadata:  metadataJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return dbDeviceToProto(dbDevice)
}

// GetDevice implements Storage.GetDevice.
func (s *PostgresStorage) GetDevice(ctx context.Context, id string) (*pb.Device, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device ID: %v", err)
	}

	dbDevice, err := s.queries.GetDevice(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "device %s not found", id)
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return dbDeviceToProto(dbDevice)
}

// ListDevices implements Storage.ListDevices.
func (s *PostgresStorage) ListDevices(ctx context.Context, page, pageSize int32) ([]*pb.Device, int32, error) {
	// Get total count
	total, err := s.queries.CountDevices(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count devices: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get devices
	dbDevices, err := s.queries.ListDevices(ctx, sqlc.ListDevicesParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list devices: %w", err)
	}

	// Convert to proto
	devices := make([]*pb.Device, len(dbDevices))
	for i, dbDevice := range dbDevices {
		device, err := dbDeviceToProto(dbDevice)
		if err != nil {
			return nil, 0, err
		}
		devices[i] = device
	}

	return devices, int32(total), nil
}

// UpdateDevice implements Storage.UpdateDevice.
func (s *PostgresStorage) UpdateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(device.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device ID: %v", err)
	}

	// Prepare metadata
	var metadataJSON []byte
	var err error
	if device.Metadata != nil {
		metadataJSON, err = json.Marshal(device.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Convert status
	dbStatus := protoStatusToDBStatus(device.Status)

	// Update last_seen
	lastSeen := pgtype.Timestamptz{}
	if err := lastSeen.Scan(time.Unix(device.LastSeen, 0)); err != nil {
		return nil, fmt.Errorf("failed to parse last_seen: %w", err)
	}

	// Update device
	dbDevice, err := s.queries.UpdateDevice(ctx, sqlc.UpdateDeviceParams{
		ID:       pgUUID,
		Name:     device.Name,
		Status:   dbStatus,
		Metadata: metadataJSON,
		LastSeen: lastSeen,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "device %s not found", device.Id)
		}
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	return dbDeviceToProto(dbDevice)
}

// DeleteDevice implements Storage.DeleteDevice.
func (s *PostgresStorage) DeleteDevice(ctx context.Context, id string) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid device ID: %v", err)
	}

	// Check if device exists first
	_, err := s.queries.GetDevice(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return status.Errorf(codes.NotFound, "device %s not found", id)
		}
		return fmt.Errorf("failed to check device existence: %w", err)
	}

	// Delete device
	if err := s.queries.DeleteDevice(ctx, pgUUID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

// Close implements Storage.Close.
func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

// Helper functions for conversion

func dbDeviceToProto(dbDevice sqlc.Device) (*pb.Device, error) {
	// Parse metadata
	var metadata map[string]string
	if len(dbDevice.Metadata) > 0 {
		if err := json.Unmarshal(dbDevice.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &pb.Device{
		Id:        dbDevice.ID.String(),
		Name:      dbDevice.Name,
		Type:      dbDevice.Type,
		Status:    dbStatusToProtoStatus(dbDevice.Status),
		CreatedAt: dbDevice.CreatedAt.Time.Unix(),
		LastSeen:  dbDevice.LastSeen.Time.Unix(),
		Metadata:  metadata,
	}, nil
}

func protoStatusToDBStatus(status pb.DeviceStatus) sqlc.DeviceStatus {
	switch status {
	case pb.DeviceStatus_ONLINE:
		return sqlc.DeviceStatusONLINE
	case pb.DeviceStatus_OFFLINE:
		return sqlc.DeviceStatusOFFLINE
	case pb.DeviceStatus_ERROR:
		return sqlc.DeviceStatusERROR
	case pb.DeviceStatus_MAINTENANCE:
		return sqlc.DeviceStatusMAINTENANCE
	default:
		return sqlc.DeviceStatusUNKNOWN
	}
}

func dbStatusToProtoStatus(status sqlc.DeviceStatus) pb.DeviceStatus {
	switch status {
	case sqlc.DeviceStatusONLINE:
		return pb.DeviceStatus_ONLINE
	case sqlc.DeviceStatusOFFLINE:
		return pb.DeviceStatus_OFFLINE
	case sqlc.DeviceStatusERROR:
		return pb.DeviceStatus_ERROR
	case sqlc.DeviceStatusMAINTENANCE:
		return pb.DeviceStatus_MAINTENANCE
	default:
		return pb.DeviceStatus_UNKNOWN
	}
}
