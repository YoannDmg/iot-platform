package storage

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto"
)

// MemoryStorage implements Storage interface using in-memory map.
// Thread-safe using RWMutex. Primarily for testing and development.
type MemoryStorage struct {
	mu      sync.RWMutex
	devices map[string]*pb.Device
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		devices: make(map[string]*pb.Device),
	}
}

// CreateDevice implements Storage.CreateDevice.
func (s *MemoryStorage) CreateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy to avoid external mutations
	stored := &pb.Device{
		Id:        device.Id,
		Name:      device.Name,
		Type:      device.Type,
		Status:    device.Status,
		CreatedAt: device.CreatedAt,
		LastSeen:  device.LastSeen,
		Metadata:  copyMetadata(device.Metadata),
	}

	s.devices[device.Id] = stored
	return stored, nil
}

// GetDevice implements Storage.GetDevice.
func (s *MemoryStorage) GetDevice(ctx context.Context, id string) (*pb.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "device %s not found", id)
	}

	// Return a copy
	return &pb.Device{
		Id:        device.Id,
		Name:      device.Name,
		Type:      device.Type,
		Status:    device.Status,
		CreatedAt: device.CreatedAt,
		LastSeen:  device.LastSeen,
		Metadata:  copyMetadata(device.Metadata),
	}, nil
}

// ListDevices implements Storage.ListDevices.
func (s *MemoryStorage) ListDevices(ctx context.Context, page, pageSize int32) ([]*pb.Device, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := int32(len(s.devices))

	// Simple pagination - collect all devices
	devices := make([]*pb.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, &pb.Device{
			Id:        device.Id,
			Name:      device.Name,
			Type:      device.Type,
			Status:    device.Status,
			CreatedAt: device.CreatedAt,
			LastSeen:  device.LastSeen,
			Metadata:  copyMetadata(device.Metadata),
		})
	}

	return devices, total, nil
}

// UpdateDevice implements Storage.UpdateDevice.
func (s *MemoryStorage) UpdateDevice(ctx context.Context, device *pb.Device) (*pb.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.devices[device.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "device %s not found", device.Id)
	}

	// Update mutable fields
	if device.Name != "" {
		existing.Name = device.Name
	}
	if device.Status != pb.DeviceStatus_UNKNOWN {
		existing.Status = device.Status
	}
	if device.Metadata != nil {
		existing.Metadata = copyMetadata(device.Metadata)
	}
	existing.LastSeen = device.LastSeen

	// Return copy
	return &pb.Device{
		Id:        existing.Id,
		Name:      existing.Name,
		Type:      existing.Type,
		Status:    existing.Status,
		CreatedAt: existing.CreatedAt,
		LastSeen:  existing.LastSeen,
		Metadata:  copyMetadata(existing.Metadata),
	}, nil
}

// DeleteDevice implements Storage.DeleteDevice.
func (s *MemoryStorage) DeleteDevice(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; !exists {
		return status.Errorf(codes.NotFound, "device %s not found", id)
	}

	delete(s.devices, id)
	return nil
}

// Close implements Storage.Close.
func (s *MemoryStorage) Close() error {
	// Nothing to clean up for in-memory storage
	return nil
}

// Helper function to copy metadata map
func copyMetadata(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
