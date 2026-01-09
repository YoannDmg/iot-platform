// Package main implements the Device Manager service.
// Microservice for IoT device lifecycle management (CRUD + monitoring).
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/iot-platform/shared/proto"
)

// DeviceServer implements pb.DeviceServiceServer interface.
// Thread-safe using sync.RWMutex for concurrent reads.
//
// TODO Production:
//   - Replace in-memory storage with PostgreSQL
//   - Add interceptors (logging, auth, metrics)
//   - Implement graceful shutdown
type DeviceServer struct {
	pb.UnimplementedDeviceServiceServer
	mu      sync.RWMutex
	devices map[string]*pb.Device
}

// NewDeviceServer creates a new server instance.
func NewDeviceServer() *DeviceServer {
	return &DeviceServer{
		devices: make(map[string]*pb.Device),
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

	s.mu.Lock()
	s.devices[device.Id] = device
	s.mu.Unlock()

	log.Printf("✅ Device created: id=%s", device.Id)
	return &pb.CreateDeviceResponse{Device: device}, nil
}

// GetDevice retrieves a device by ID.
func (s *DeviceServer) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	log.Printf("📥 GetDevice: id=%s", req.Id)

	s.mu.RLock()
	device, exists := s.devices[req.Id]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "device %s not found", req.Id)
	}

	log.Printf("✅ Device found: id=%s, name=%s", device.Id, device.Name)
	return &pb.GetDeviceResponse{Device: device}, nil
}

// ListDevices returns paginated device list.
// TODO: Implement actual pagination and sorting.
func (s *DeviceServer) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	log.Printf("📥 ListDevices: page=%d, pageSize=%d", req.Page, req.PageSize)

	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*pb.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}

	log.Printf("✅ %d devices found", len(devices))
	return &pb.ListDevicesResponse{
		Devices:  devices,
		Total:    int32(len(devices)),
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

	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "device %s not found", req.Id)
	}

	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Status != pb.DeviceStatus_UNKNOWN {
		device.Status = req.Status
	}
	if req.Metadata != nil {
		device.Metadata = req.Metadata
	}

	log.Printf("✅ Device updated: id=%s", device.Id)
	return &pb.UpdateDeviceResponse{Device: device}, nil
}

// DeleteDevice removes a device by ID.
func (s *DeviceServer) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	log.Printf("📥 DeleteDevice: id=%s", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[req.Id]; !exists {
		return nil, status.Errorf(codes.NotFound, "device %s not found", req.Id)
	}

	delete(s.devices, req.Id)
	log.Printf("✅ Device deleted: id=%s", req.Id)

	return &pb.DeleteDeviceResponse{
		Success: true,
		Message: fmt.Sprintf("Device %s deleted", req.Id),
	}, nil
}

// main initializes and starts the Device Manager gRPC server.
//
// Configuration:
//   - Port: 8081
//   - Protocol: gRPC (HTTP/2)
//
// TODO Production:
//   - Configurable port via env var
//   - TLS/mTLS support
//   - Health check endpoint
//   - Graceful shutdown
func main() {
	port := 8081

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("❌ Failed to create listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	deviceServer := NewDeviceServer()
	pb.RegisterDeviceServiceServer(grpcServer, deviceServer)

	log.Println("=====================================")
	log.Printf("Device Manager Service")
	log.Println("=====================================")
	log.Printf("Protocol: gRPC (HTTP/2)")
	log.Printf("Port: %d", port)
	log.Printf("Address: http://localhost:%d", port)
	log.Println("-------------------------------------")
	log.Printf("✅ Server started")
	log.Printf("⏳ Waiting for gRPC connections...")
	log.Println("=====================================")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
