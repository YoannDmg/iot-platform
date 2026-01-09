// Package main implements the Device Manager service.
// Microservice for IoT device lifecycle management (CRUD + monitoring).
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto"
	"github.com/yourusername/iot-platform/services/device-manager/storage"
)

// DeviceServer implements pb.DeviceServiceServer interface.
// Uses pluggable Storage backend (PostgreSQL or in-memory).
//
// TODO Production:
//   - Add interceptors (logging, auth, metrics)
//   - Implement graceful shutdown
type DeviceServer struct {
	pb.UnimplementedDeviceServiceServer
	storage storage.Storage
}

// NewDeviceServer creates a new server instance with the given storage backend.
func NewDeviceServer(store storage.Storage) *DeviceServer {
	return &DeviceServer{
		storage: store,
	}
}

// CreateDevice creates a new device with generated UUID and timestamps.
func (s *DeviceServer) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CreateDeviceResponse, error) {
	log.Printf("📥 CreateDevice: name=%s, type=%s", req.Name, req.Type)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}
	if req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "type required")
	}

	now := time.Now().Unix()
	device := &pb.Device{
		Id:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: now,
		LastSeen:  now,
		Metadata:  req.Metadata,
	}

	createdDevice, err := s.storage.CreateDevice(ctx, device)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Device created: id=%s", createdDevice.Id)
	return &pb.CreateDeviceResponse{Device: createdDevice}, nil
}

// GetDevice retrieves a device by ID.
func (s *DeviceServer) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	log.Printf("📥 GetDevice: id=%s", req.Id)

	device, err := s.storage.GetDevice(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Device found: id=%s, name=%s", device.Id, device.Name)
	return &pb.GetDeviceResponse{Device: device}, nil
}

// ListDevices returns paginated device list.
func (s *DeviceServer) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	log.Printf("📥 ListDevices: page=%d, pageSize=%d", req.Page, req.PageSize)

	devices, total, err := s.storage.ListDevices(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ %d devices found", len(devices))
	return &pb.ListDevicesResponse{
		Devices:  devices,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateDevice updates an existing device.
// Mutable fields: Name, Status, Metadata. ID and CreatedAt are immutable.
func (s *DeviceServer) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	log.Printf("📥 UpdateDevice: id=%s", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID required")
	}

	// Prepare update with fields to change
	device := &pb.Device{
		Id:       req.Id,
		Name:     req.Name,
		Status:   req.Status,
		Metadata: req.Metadata,
		LastSeen: time.Now().Unix(),
	}

	updatedDevice, err := s.storage.UpdateDevice(ctx, device)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Device updated: id=%s", updatedDevice.Id)
	return &pb.UpdateDeviceResponse{Device: updatedDevice}, nil
}

// DeleteDevice removes a device by ID.
func (s *DeviceServer) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	log.Printf("📥 DeleteDevice: id=%s", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID required")
	}

	if err := s.storage.DeleteDevice(ctx, req.Id); err != nil {
		return nil, err
	}

	log.Printf("✅ Device deleted: id=%s", req.Id)
	return &pb.DeleteDeviceResponse{
		Success: true,
		Message: fmt.Sprintf("Device %s deleted", req.Id),
	}, nil
}

// main initializes and starts the Device Manager gRPC server.
//
// Configuration via environment variables:
//   - STORAGE_TYPE: "postgres" or "memory" (default: memory)
//   - DB_HOST: PostgreSQL host (default: localhost)
//   - DB_PORT: PostgreSQL port (default: 5432)
//   - DB_NAME: Database name (default: iot_platform)
//   - DB_USER: Database user (default: iot_user)
//   - DB_PASSWORD: Database password (default: iot_password)
//   - DB_SSLMODE: SSL mode (default: disable)
//
// TODO Production:
//   - TLS/mTLS support
//   - Health check endpoint
//   - Graceful shutdown
func main() {
	ctx := context.Background()
	port := 8081

	// Configure storage backend
	storageType := getEnv("STORAGE_TYPE", "memory")
	var store storage.Storage
	var err error

	switch storageType {
	case "postgres":
		// Build PostgreSQL DSN
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("DB_USER", "iot_user"),
			getEnv("DB_PASSWORD", "iot_password"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "iot_platform"),
			getEnv("DB_SSLMODE", "disable"),
		)
		store, err = storage.NewPostgresStorage(ctx, dsn)
		if err != nil {
			log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
		}
		defer store.Close()
		log.Printf("✅ Using PostgreSQL storage")
	default:
		store = storage.NewMemoryStorage()
		log.Printf("✅ Using in-memory storage")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("❌ Failed to create listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	deviceServer := NewDeviceServer(store)
	pb.RegisterDeviceServiceServer(grpcServer, deviceServer)

	log.Println("=====================================")
	log.Printf("Device Manager Service")
	log.Println("=====================================")
	log.Printf("Protocol: gRPC (HTTP/2)")
	log.Printf("Port: %d", port)
	log.Printf("Storage: %s", storageType)
	log.Printf("Address: http://localhost:%d", port)
	log.Println("-------------------------------------")
	log.Printf("✅ Server started")
	log.Printf("⏳ Waiting for gRPC connections...")
	log.Println("=====================================")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
