// Package storage provides an abstraction layer for device persistence.
// Supports multiple backends (PostgreSQL, in-memory) via a common interface.
package storage

import (
	"context"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
)

// Storage defines the interface for device persistence operations.
// Implementations: PostgresStorage (production), MemoryStorage (tests/dev).
type Storage interface {
	// CreateDevice stores a new device and returns the stored device.
	CreateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error)

	// GetDevice retrieves a device by ID.
	// Returns nil, ErrNotFound if device doesn't exist.
	GetDevice(ctx context.Context, id string) (*pb.Device, error)

	// ListDevices returns a paginated list of devices.
	// Returns devices, total count, and error.
	ListDevices(ctx context.Context, page, pageSize int32) ([]*pb.Device, int32, error)

	// UpdateDevice updates an existing device.
	// Only non-zero/non-nil fields are updated.
	// Returns updated device or ErrNotFound.
	UpdateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error)

	// DeleteDevice removes a device by ID.
	// Returns ErrNotFound if device doesn't exist.
	DeleteDevice(ctx context.Context, id string) error

	// Close releases any resources held by the storage.
	Close() error
}
