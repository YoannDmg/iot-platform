package graph

import (
	"context"
	"errors"
	"testing"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockDeviceServiceClient is a mock implementation of pb.DeviceServiceClient for testing.
type MockDeviceServiceClient struct {
	pb.DeviceServiceClient

	// Mock function implementations
	CreateDeviceFunc func(ctx context.Context, req *pb.CreateDeviceRequest, opts ...grpc.CallOption) (*pb.CreateDeviceResponse, error)
	GetDeviceFunc    func(ctx context.Context, req *pb.GetDeviceRequest, opts ...grpc.CallOption) (*pb.GetDeviceResponse, error)
	ListDevicesFunc  func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error)
	UpdateDeviceFunc func(ctx context.Context, req *pb.UpdateDeviceRequest, opts ...grpc.CallOption) (*pb.UpdateDeviceResponse, error)
	DeleteDeviceFunc func(ctx context.Context, req *pb.DeleteDeviceRequest, opts ...grpc.CallOption) (*pb.DeleteDeviceResponse, error)
}

func (m *MockDeviceServiceClient) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest, opts ...grpc.CallOption) (*pb.CreateDeviceResponse, error) {
	if m.CreateDeviceFunc != nil {
		return m.CreateDeviceFunc(ctx, req, opts...)
	}
	return nil, errors.New("CreateDeviceFunc not implemented")
}

func (m *MockDeviceServiceClient) GetDevice(ctx context.Context, req *pb.GetDeviceRequest, opts ...grpc.CallOption) (*pb.GetDeviceResponse, error) {
	if m.GetDeviceFunc != nil {
		return m.GetDeviceFunc(ctx, req, opts...)
	}
	return nil, errors.New("GetDeviceFunc not implemented")
}

func (m *MockDeviceServiceClient) ListDevices(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
	if m.ListDevicesFunc != nil {
		return m.ListDevicesFunc(ctx, req, opts...)
	}
	return nil, errors.New("ListDevicesFunc not implemented")
}

func (m *MockDeviceServiceClient) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest, opts ...grpc.CallOption) (*pb.UpdateDeviceResponse, error) {
	if m.UpdateDeviceFunc != nil {
		return m.UpdateDeviceFunc(ctx, req, opts...)
	}
	return nil, errors.New("UpdateDeviceFunc not implemented")
}

func (m *MockDeviceServiceClient) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest, opts ...grpc.CallOption) (*pb.DeleteDeviceResponse, error) {
	if m.DeleteDeviceFunc != nil {
		return m.DeleteDeviceFunc(ctx, req, opts...)
	}
	return nil, errors.New("DeleteDeviceFunc not implemented")
}

// Helper function to create a test resolver with mock client
func newTestResolver(mock *MockDeviceServiceClient) *Resolver {
	return &Resolver{
		DeviceClient: mock,
	}
}

// TestCreateDeviceImpl tests the CreateDevice mutation resolver.
func TestCreateDeviceImpl(t *testing.T) {
	tests := []struct {
		name      string
		input     model.CreateDeviceInput
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, device *model.Device)
	}{
		{
			name: "successful_creation",
			input: model.CreateDeviceInput{
				Name: "Test Sensor",
				Type: "temperature",
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.CreateDeviceFunc = func(ctx context.Context, req *pb.CreateDeviceRequest, opts ...grpc.CallOption) (*pb.CreateDeviceResponse, error) {
					return &pb.CreateDeviceResponse{
						Device: &pb.Device{
							Id:        "test-id-123",
							Name:      req.Name,
							Type:      req.Type,
							Status:    pb.DeviceStatus_ONLINE,
							CreatedAt: 1234567890,
							LastSeen:  1234567890,
							Metadata:  req.Metadata,
						},
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, device *model.Device) {
				if device.ID != "test-id-123" {
					t.Errorf("expected ID 'test-id-123', got %s", device.ID)
				}
				if device.Name != "Test Sensor" {
					t.Errorf("expected name 'Test Sensor', got %s", device.Name)
				}
				if device.Status != model.DeviceStatusOnline {
					t.Errorf("expected status ONLINE, got %v", device.Status)
				}
			},
		},
		{
			name: "with_metadata",
			input: model.CreateDeviceInput{
				Name: "Smart Device",
				Type: "hvac",
				Metadata: []*model.MetadataEntryInput{
					{Key: "location", Value: "room-101"},
					{Key: "floor", Value: "2"},
				},
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.CreateDeviceFunc = func(ctx context.Context, req *pb.CreateDeviceRequest, opts ...grpc.CallOption) (*pb.CreateDeviceResponse, error) {
					return &pb.CreateDeviceResponse{
						Device: &pb.Device{
							Id:        "test-id-456",
							Name:      req.Name,
							Type:      req.Type,
							Status:    pb.DeviceStatus_ONLINE,
							CreatedAt: 1234567890,
							LastSeen:  1234567890,
							Metadata:  req.Metadata,
						},
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, device *model.Device) {
				if len(device.Metadata) != 2 {
					t.Errorf("expected 2 metadata entries, got %d", len(device.Metadata))
				}
			},
		},
		{
			name: "grpc_error",
			input: model.CreateDeviceInput{
				Name: "Test",
				Type: "sensor",
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.CreateDeviceFunc = func(ctx context.Context, req *pb.CreateDeviceRequest, opts ...grpc.CallOption) (*pb.CreateDeviceResponse, error) {
					return nil, status.Error(codes.InvalidArgument, "name required")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			mutationResolver := &mutationResolver{resolver}

			device, err := mutationResolver.CreateDeviceImpl(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if device == nil {
				t.Error("device is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, device)
			}
		})
	}
}

// TestGetDeviceImpl tests the GetDevice query resolver.
func TestGetDeviceImpl(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, device *model.Device)
	}{
		{
			name:     "existing_device",
			deviceID: "test-id-123",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.GetDeviceFunc = func(ctx context.Context, req *pb.GetDeviceRequest, opts ...grpc.CallOption) (*pb.GetDeviceResponse, error) {
					return &pb.GetDeviceResponse{
						Device: &pb.Device{
							Id:        req.Id,
							Name:      "Test Device",
							Type:      "sensor",
							Status:    pb.DeviceStatus_ONLINE,
							CreatedAt: 1234567890,
							LastSeen:  1234567890,
						},
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, device *model.Device) {
				if device.ID != "test-id-123" {
					t.Errorf("expected ID 'test-id-123', got %s", device.ID)
				}
				if device.Name != "Test Device" {
					t.Errorf("expected name 'Test Device', got %s", device.Name)
				}
			},
		},
		{
			name:     "device_not_found",
			deviceID: "non-existent",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.GetDeviceFunc = func(ctx context.Context, req *pb.GetDeviceRequest, opts ...grpc.CallOption) (*pb.GetDeviceResponse, error) {
					return nil, status.Error(codes.NotFound, "device not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			queryResolver := &queryResolver{resolver}

			device, err := queryResolver.DeviceImpl(context.Background(), tt.deviceID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if device == nil {
				t.Error("device is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, device)
			}
		})
	}
}

// TestListDevicesImpl tests the ListDevices query resolver.
func TestListDevicesImpl(t *testing.T) {
	tests := []struct {
		name      string
		page      *int
		pageSize  *int
		typeArg   *string
		status    *model.DeviceStatus
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, conn *model.DeviceConnection)
	}{
		{
			name: "default_pagination",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.ListDevicesFunc = func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
					return &pb.ListDevicesResponse{
						Devices: []*pb.Device{
							{
								Id:        "id-1",
								Name:      "Device 1",
								Type:      "sensor",
								Status:    pb.DeviceStatus_ONLINE,
								CreatedAt: 1234567890,
								LastSeen:  1234567890,
							},
							{
								Id:        "id-2",
								Name:      "Device 2",
								Type:      "actuator",
								Status:    pb.DeviceStatus_OFFLINE,
								CreatedAt: 1234567890,
								LastSeen:  1234567890,
							},
						},
						Total:    2,
						Page:     1,
						PageSize: 10,
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, conn *model.DeviceConnection) {
				if len(conn.Devices) != 2 {
					t.Errorf("expected 2 devices, got %d", len(conn.Devices))
				}
				if conn.Total != 2 {
					t.Errorf("expected total 2, got %d", conn.Total)
				}
			},
		},
		{
			name: "with_type_filter",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.ListDevicesFunc = func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
					return &pb.ListDevicesResponse{
						Devices: []*pb.Device{
							{
								Id:        "id-1",
								Name:      "Device 1",
								Type:      "sensor",
								Status:    pb.DeviceStatus_ONLINE,
								CreatedAt: 1234567890,
								LastSeen:  1234567890,
							},
							{
								Id:        "id-2",
								Name:      "Device 2",
								Type:      "actuator",
								Status:    pb.DeviceStatus_ONLINE,
								CreatedAt: 1234567890,
								LastSeen:  1234567890,
							},
						},
						Total:    2,
						Page:     1,
						PageSize: 10,
					}, nil
				}
			},
			typeArg: stringPtr("sensor"),
			wantErr: false,
			validate: func(t *testing.T, conn *model.DeviceConnection) {
				if len(conn.Devices) != 1 {
					t.Errorf("expected 1 device after filtering, got %d", len(conn.Devices))
				}
				if conn.Devices[0].Type != "sensor" {
					t.Errorf("expected type 'sensor', got %s", conn.Devices[0].Type)
				}
			},
		},
		{
			name: "empty_list",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.ListDevicesFunc = func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
					return &pb.ListDevicesResponse{
						Devices:  []*pb.Device{},
						Total:    0,
						Page:     1,
						PageSize: 10,
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, conn *model.DeviceConnection) {
				if len(conn.Devices) != 0 {
					t.Errorf("expected 0 devices, got %d", len(conn.Devices))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			queryResolver := &queryResolver{resolver}

			conn, err := queryResolver.DevicesImpl(context.Background(), tt.page, tt.pageSize, tt.typeArg, tt.status)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if conn == nil {
				t.Error("connection is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, conn)
			}
		})
	}
}

// TestUpdateDeviceImpl tests the UpdateDevice mutation resolver.
func TestUpdateDeviceImpl(t *testing.T) {
	tests := []struct {
		name      string
		input     model.UpdateDeviceInput
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, device *model.Device)
	}{
		{
			name: "update_name",
			input: model.UpdateDeviceInput{
				ID:   "test-id-123",
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.UpdateDeviceFunc = func(ctx context.Context, req *pb.UpdateDeviceRequest, opts ...grpc.CallOption) (*pb.UpdateDeviceResponse, error) {
					return &pb.UpdateDeviceResponse{
						Device: &pb.Device{
							Id:        req.Id,
							Name:      req.Name,
							Type:      "sensor",
							Status:    pb.DeviceStatus_ONLINE,
							CreatedAt: 1234567890,
							LastSeen:  1234567890,
						},
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, device *model.Device) {
				if device.Name != "Updated Name" {
					t.Errorf("expected name 'Updated Name', got %s", device.Name)
				}
			},
		},
		{
			name: "update_status",
			input: model.UpdateDeviceInput{
				ID:     "test-id-123",
				Status: statusPtr(model.DeviceStatusOffline),
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.UpdateDeviceFunc = func(ctx context.Context, req *pb.UpdateDeviceRequest, opts ...grpc.CallOption) (*pb.UpdateDeviceResponse, error) {
					return &pb.UpdateDeviceResponse{
						Device: &pb.Device{
							Id:        req.Id,
							Name:      "Test Device",
							Type:      "sensor",
							Status:    req.Status,
							CreatedAt: 1234567890,
							LastSeen:  1234567890,
						},
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, device *model.Device) {
				if device.Status != model.DeviceStatusOffline {
					t.Errorf("expected status OFFLINE, got %v", device.Status)
				}
			},
		},
		{
			name: "device_not_found",
			input: model.UpdateDeviceInput{
				ID:   "non-existent",
				Name: stringPtr("Test"),
			},
			mockSetup: func(m *MockDeviceServiceClient) {
				m.UpdateDeviceFunc = func(ctx context.Context, req *pb.UpdateDeviceRequest, opts ...grpc.CallOption) (*pb.UpdateDeviceResponse, error) {
					return nil, status.Error(codes.NotFound, "device not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			mutationResolver := &mutationResolver{resolver}

			device, err := mutationResolver.UpdateDeviceImpl(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if device == nil {
				t.Error("device is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, device)
			}
		})
	}
}

// TestDeleteDeviceImpl tests the DeleteDevice mutation resolver.
func TestDeleteDeviceImpl(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, result *model.DeleteResult)
	}{
		{
			name:     "successful_deletion",
			deviceID: "test-id-123",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.DeleteDeviceFunc = func(ctx context.Context, req *pb.DeleteDeviceRequest, opts ...grpc.CallOption) (*pb.DeleteDeviceResponse, error) {
					return &pb.DeleteDeviceResponse{
						Success: true,
						Message: "Device deleted successfully",
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *model.DeleteResult) {
				if !result.Success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name:     "device_not_found",
			deviceID: "non-existent",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.DeleteDeviceFunc = func(ctx context.Context, req *pb.DeleteDeviceRequest, opts ...grpc.CallOption) (*pb.DeleteDeviceResponse, error) {
					return nil, status.Error(codes.NotFound, "device not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			mutationResolver := &mutationResolver{resolver}

			result, err := mutationResolver.DeleteDeviceImpl(context.Background(), tt.deviceID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestStatsImpl tests the Stats query resolver.
func TestStatsImpl(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockDeviceServiceClient)
		wantErr   bool
		validate  func(t *testing.T, stats *model.Stats)
	}{
		{
			name: "compute_stats",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.ListDevicesFunc = func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
					return &pb.ListDevicesResponse{
						Devices: []*pb.Device{
							{Id: "1", Status: pb.DeviceStatus_ONLINE},
							{Id: "2", Status: pb.DeviceStatus_ONLINE},
							{Id: "3", Status: pb.DeviceStatus_OFFLINE},
							{Id: "4", Status: pb.DeviceStatus_ERROR},
						},
						Total: 4,
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, stats *model.Stats) {
				if stats.TotalDevices != 4 {
					t.Errorf("expected total 4, got %d", stats.TotalDevices)
				}
				if stats.OnlineDevices != 2 {
					t.Errorf("expected online 2, got %d", stats.OnlineDevices)
				}
				if stats.OfflineDevices != 1 {
					t.Errorf("expected offline 1, got %d", stats.OfflineDevices)
				}
				if stats.ErrorDevices != 1 {
					t.Errorf("expected error 1, got %d", stats.ErrorDevices)
				}
			},
		},
		{
			name: "empty_stats",
			mockSetup: func(m *MockDeviceServiceClient) {
				m.ListDevicesFunc = func(ctx context.Context, req *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
					return &pb.ListDevicesResponse{
						Devices: []*pb.Device{},
						Total:   0,
					}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, stats *model.Stats) {
				if stats.TotalDevices != 0 {
					t.Errorf("expected total 0, got %d", stats.TotalDevices)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDeviceServiceClient{}
			tt.mockSetup(mock)

			resolver := newTestResolver(mock)
			queryResolver := &queryResolver{resolver}

			stats, err := queryResolver.StatsImpl(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if stats == nil {
				t.Error("stats is nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, stats)
			}
		})
	}
}

// Test helper conversion functions

func TestProtoToGraphQLDevice(t *testing.T) {
	proto := &pb.Device{
		Id:        "test-123",
		Name:      "Test Device",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: 1234567890,
		LastSeen:  1234567890,
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	graphql := protoToGraphQLDevice(proto)

	if graphql == nil {
		t.Fatal("result is nil")
	}

	if graphql.ID != proto.Id {
		t.Errorf("expected ID %s, got %s", proto.Id, graphql.ID)
	}
	if graphql.Name != proto.Name {
		t.Errorf("expected name %s, got %s", proto.Name, graphql.Name)
	}
	if graphql.Status != model.DeviceStatusOnline {
		t.Errorf("expected status ONLINE, got %v", graphql.Status)
	}
	if len(graphql.Metadata) != 2 {
		t.Errorf("expected 2 metadata entries, got %d", len(graphql.Metadata))
	}
}

func TestProtoToGraphQLDevice_Nil(t *testing.T) {
	result := protoToGraphQLDevice(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestProtoToGraphQLStatus(t *testing.T) {
	tests := []struct {
		proto    pb.DeviceStatus
		expected model.DeviceStatus
	}{
		{pb.DeviceStatus_ONLINE, model.DeviceStatusOnline},
		{pb.DeviceStatus_OFFLINE, model.DeviceStatusOffline},
		{pb.DeviceStatus_ERROR, model.DeviceStatusError},
		{pb.DeviceStatus_MAINTENANCE, model.DeviceStatusMaintenance},
		{pb.DeviceStatus_UNKNOWN, model.DeviceStatusUnknown},
	}

	for _, tt := range tests {
		result := protoToGraphQLStatus(tt.proto)
		if result != tt.expected {
			t.Errorf("for proto status %v, expected %v, got %v", tt.proto, tt.expected, result)
		}
	}
}

func TestGraphQLToProtoStatus(t *testing.T) {
	tests := []struct {
		graphql  *model.DeviceStatus
		expected pb.DeviceStatus
	}{
		{statusPtr(model.DeviceStatusOnline), pb.DeviceStatus_ONLINE},
		{statusPtr(model.DeviceStatusOffline), pb.DeviceStatus_OFFLINE},
		{statusPtr(model.DeviceStatusError), pb.DeviceStatus_ERROR},
		{statusPtr(model.DeviceStatusMaintenance), pb.DeviceStatus_MAINTENANCE},
		{statusPtr(model.DeviceStatusUnknown), pb.DeviceStatus_UNKNOWN},
		{nil, pb.DeviceStatus_UNKNOWN},
	}

	for _, tt := range tests {
		result := graphQLToProtoStatus(tt.graphql)
		if result != tt.expected {
			t.Errorf("for graphql status %v, expected %v, got %v", tt.graphql, tt.expected, result)
		}
	}
}

// Helper functions for tests

func stringPtr(s string) *string {
	return &s
}

func statusPtr(s model.DeviceStatus) *model.DeviceStatus {
	return &s
}
