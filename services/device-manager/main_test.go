// +build unit

package main

import (
	"context"
	"sync"
	"testing"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
	"github.com/yourusername/iot-platform/services/device-manager/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestCreateDevice tests device creation functionality.
func TestCreateDevice(t *testing.T) {
	tests := []struct {
		name        string
		request     *pb.CreateDeviceRequest
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name: "valid_device",
			request: &pb.CreateDeviceRequest{
				Name: "Test Sensor",
				Type: "temperature",
				Metadata: map[string]string{
					"location": "room-101",
				},
			},
			wantErr:     false,
			description: "should successfully create a valid device",
		},
		{
			name: "missing_name",
			request: &pb.CreateDeviceRequest{
				Type: "temperature",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when name is missing",
		},
		{
			name: "missing_type",
			request: &pb.CreateDeviceRequest{
				Name: "Test Sensor",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when type is missing",
		},
		{
			name: "with_metadata",
			request: &pb.CreateDeviceRequest{
				Name: "Smart Thermostat",
				Type: "hvac",
				Metadata: map[string]string{
					"floor":    "2",
					"building": "A",
					"zone":     "north",
				},
			},
			wantErr:     false,
			description: "should create device with metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewDeviceServer(storage.NewMemoryStorage())
			ctx := context.Background()

			resp, err := server.CreateDevice(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				} else {
					t.Errorf("%s: expected gRPC status error, got %v", tt.description, err)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			// Validate response
			if resp.Device == nil {
				t.Errorf("%s: device is nil", tt.description)
				return
			}

			device := resp.Device

			// Check ID is generated
			if device.Id == "" {
				t.Errorf("%s: device ID should not be empty", tt.description)
			}

			// Check fields match request
			if device.Name != tt.request.Name {
				t.Errorf("%s: expected name %s, got %s", tt.description, tt.request.Name, device.Name)
			}
			if device.Type != tt.request.Type {
				t.Errorf("%s: expected type %s, got %s", tt.description, tt.request.Type, device.Type)
			}

			// Check default status
			if device.Status != pb.DeviceStatus_ONLINE {
				t.Errorf("%s: expected status ONLINE, got %v", tt.description, device.Status)
			}

			// Check timestamps are set
			if device.CreatedAt == 0 {
				t.Errorf("%s: CreatedAt should be set", tt.description)
			}
			if device.LastSeen == 0 {
				t.Errorf("%s: LastSeen should be set", tt.description)
			}

			// Check metadata if provided
			if tt.request.Metadata != nil {
				if device.Metadata == nil {
					t.Errorf("%s: metadata should not be nil", tt.description)
				}
				for key, val := range tt.request.Metadata {
					if device.Metadata[key] != val {
						t.Errorf("%s: metadata[%s]: expected %s, got %s", tt.description, key, val, device.Metadata[key])
					}
				}
			}
		})
	}
}

// TestGetDevice tests device retrieval functionality.
func TestGetDevice(t *testing.T) {
	server := NewDeviceServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create a test device first
	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Test Device",
		Type: "sensor",
	})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}
	deviceID := createResp.Device.Id

	tests := []struct {
		name        string
		deviceID    string
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name:        "existing_device",
			deviceID:    deviceID,
			wantErr:     false,
			description: "should retrieve existing device",
		},
		{
			name:        "non_existing_device",
			deviceID:    "non-existent-id",
			wantErr:     true,
			wantCode:    codes.NotFound,
			description: "should return NotFound for non-existing device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetDevice(ctx, &pb.GetDeviceRequest{
				Id: tt.deviceID,
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.Device == nil {
				t.Errorf("%s: device is nil", tt.description)
				return
			}

			if resp.Device.Id != tt.deviceID {
				t.Errorf("%s: expected ID %s, got %s", tt.description, tt.deviceID, resp.Device.Id)
			}
		})
	}
}

// TestListDevices tests device listing functionality.
func TestListDevices(t *testing.T) {
	server := NewDeviceServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Test empty list
	t.Run("empty_list", func(t *testing.T) {
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     1,
			PageSize: 10,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(resp.Devices))
		}
		if resp.Total != 0 {
			t.Errorf("expected total 0, got %d", resp.Total)
		}
	})

	// Create test devices
	deviceIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
			Name: "Device " + string(rune('A'+i)),
			Type: "sensor",
		})
		if err != nil {
			t.Fatalf("failed to create test device: %v", err)
		}
		deviceIDs[i] = createResp.Device.Id
	}

	t.Run("list_all_devices", func(t *testing.T) {
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     1,
			PageSize: 10,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Devices) != 3 {
			t.Errorf("expected 3 devices, got %d", len(resp.Devices))
		}
		if resp.Total != 3 {
			t.Errorf("expected total 3, got %d", resp.Total)
		}

		// Check all created devices are in the list
		foundIDs := make(map[string]bool)
		for _, device := range resp.Devices {
			foundIDs[device.Id] = true
		}

		for _, id := range deviceIDs {
			if !foundIDs[id] {
				t.Errorf("device ID %s not found in list", id)
			}
		}
	})
}

// TestUpdateDevice tests device update functionality.
func TestUpdateDevice(t *testing.T) {
	server := NewDeviceServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create a test device
	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Original Device",
		Type: "sensor",
		Metadata: map[string]string{
			"version": "1.0",
		},
	})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}
	deviceID := createResp.Device.Id
	originalCreatedAt := createResp.Device.CreatedAt

	tests := []struct {
		name        string
		request     *pb.UpdateDeviceRequest
		wantErr     bool
		wantCode    codes.Code
		validate    func(t *testing.T, device *pb.Device)
		description string
	}{
		{
			name: "update_name",
			request: &pb.UpdateDeviceRequest{
				Id:   deviceID,
				Name: "Updated Device",
			},
			wantErr: false,
			validate: func(t *testing.T, device *pb.Device) {
				if device.Name != "Updated Device" {
					t.Errorf("expected name 'Updated Device', got %s", device.Name)
				}
				if device.CreatedAt != originalCreatedAt {
					t.Errorf("CreatedAt should be immutable")
				}
			},
			description: "should update device name",
		},
		{
			name: "update_status",
			request: &pb.UpdateDeviceRequest{
				Id:     deviceID,
				Status: pb.DeviceStatus_OFFLINE,
			},
			wantErr: false,
			validate: func(t *testing.T, device *pb.Device) {
				if device.Status != pb.DeviceStatus_OFFLINE {
					t.Errorf("expected status OFFLINE, got %v", device.Status)
				}
			},
			description: "should update device status",
		},
		{
			name: "update_metadata",
			request: &pb.UpdateDeviceRequest{
				Id: deviceID,
				Metadata: map[string]string{
					"version":  "2.0",
					"firmware": "beta",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, device *pb.Device) {
				if device.Metadata["version"] != "2.0" {
					t.Errorf("expected version 2.0, got %s", device.Metadata["version"])
				}
				if device.Metadata["firmware"] != "beta" {
					t.Errorf("expected firmware beta, got %s", device.Metadata["firmware"])
				}
			},
			description: "should update device metadata",
		},
		{
			name: "missing_id",
			request: &pb.UpdateDeviceRequest{
				Name: "Test",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when ID is missing",
		},
		{
			name: "non_existing_device",
			request: &pb.UpdateDeviceRequest{
				Id:   "non-existent-id",
				Name: "Test",
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			description: "should return NotFound for non-existing device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.UpdateDevice(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.Device == nil {
				t.Errorf("%s: device is nil", tt.description)
				return
			}

			if tt.validate != nil {
				tt.validate(t, resp.Device)
			}
		})
	}
}

// TestDeleteDevice tests device deletion functionality.
func TestDeleteDevice(t *testing.T) {
	server := NewDeviceServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Create a test device
	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Device to Delete",
		Type: "sensor",
	})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}
	deviceID := createResp.Device.Id

	tests := []struct {
		name        string
		deviceID    string
		wantErr     bool
		wantCode    codes.Code
		description string
	}{
		{
			name:        "delete_existing_device",
			deviceID:    deviceID,
			wantErr:     false,
			description: "should delete existing device",
		},
		{
			name:        "delete_already_deleted",
			deviceID:    deviceID,
			wantErr:     true,
			wantCode:    codes.NotFound,
			description: "should return NotFound when deleting already deleted device",
		},
		{
			name:        "missing_id",
			deviceID:    "",
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			description: "should return error when ID is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.DeleteDevice(ctx, &pb.DeleteDeviceRequest{
				Id: tt.deviceID,
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("%s: expected code %v, got %v", tt.description, tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if !resp.Success {
				t.Errorf("%s: expected success=true, got false", tt.description)
			}

			// Verify device is actually deleted
			if tt.name == "delete_existing_device" {
				_, getErr := server.GetDevice(ctx, &pb.GetDeviceRequest{Id: tt.deviceID})
				if getErr == nil {
					t.Errorf("%s: device should not exist after deletion", tt.description)
				}
			}
		})
	}
}

// TestConcurrentOperations tests thread safety with concurrent access.
func TestConcurrentOperations(t *testing.T) {
	server := NewDeviceServer(storage.NewMemoryStorage())
	ctx := context.Background()

	// Number of concurrent goroutines
	numGoroutines := 50
	var wg sync.WaitGroup

	// Create devices concurrently
	t.Run("concurrent_create", func(t *testing.T) {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				_, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
					Name: "Concurrent Device",
					Type: "sensor",
				})
				if err != nil {
					t.Errorf("concurrent create failed: %v", err)
				}
			}(i)
		}
		wg.Wait()

		// Verify all devices were created
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     1,
			PageSize: 100,
		})
		if err != nil {
			t.Fatalf("failed to list devices: %v", err)
		}
		if len(resp.Devices) != numGoroutines {
			t.Errorf("expected %d devices, got %d", numGoroutines, len(resp.Devices))
		}
	})

	// Read concurrently while writing
	t.Run("concurrent_read_write", func(t *testing.T) {
		// Create a device to update
		createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
			Name: "Shared Device",
			Type: "sensor",
		})
		if err != nil {
			t.Fatalf("failed to create device: %v", err)
		}
		sharedDeviceID := createResp.Device.Id

		wg.Add(numGoroutines * 2)

		// Start readers
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				_, err := server.GetDevice(ctx, &pb.GetDeviceRequest{
					Id: sharedDeviceID,
				})
				if err != nil {
					t.Errorf("concurrent read failed: %v", err)
				}
			}()
		}

		// Start writers
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				_, err := server.UpdateDevice(ctx, &pb.UpdateDeviceRequest{
					Id:     sharedDeviceID,
					Status: pb.DeviceStatus_ONLINE,
				})
				if err != nil {
					t.Errorf("concurrent write failed: %v", err)
				}
			}(i)
		}

		wg.Wait()

		// Verify device still exists and is in valid state
		resp, err := server.GetDevice(ctx, &pb.GetDeviceRequest{
			Id: sharedDeviceID,
		})
		if err != nil {
			t.Fatalf("failed to get device after concurrent operations: %v", err)
		}
		if resp.Device == nil {
			t.Error("device should still exist")
		}
	})
}
