// +build integration

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
	"github.com/yourusername/iot-platform/services/device-manager/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// setupPostgresStorage creates a PostgreSQL storage for testing.
// Requires PostgreSQL to be running (via docker-compose).
func setupPostgresStorage(t *testing.T) storage.Storage {
	t.Helper()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnvOrDefault("DB_USER", "iot_user"),
		getEnvOrDefault("DB_PASSWORD", "iot_password"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		getEnvOrDefault("DB_NAME", "iot_platform"),
		getEnvOrDefault("DB_SSLMODE", "disable"),
	)

	store, err := storage.NewPostgresStorage(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v\nMake sure PostgreSQL is running: make up && make db-migrate", err)
	}

	// Clean up function
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	})

	return store
}

// cleanDatabase removes all devices from the database before each test.
func cleanDatabase(t *testing.T, store storage.Storage) {
	t.Helper()

	ctx := context.Background()
	devices, _, err := store.ListDevices(ctx, 1, 1000)
	if err != nil {
		t.Fatalf("Failed to list devices: %v", err)
	}

	for _, device := range devices {
		if err := store.DeleteDevice(ctx, device.Id); err != nil {
			t.Logf("Warning: Failed to delete device %s: %v", device.Id, err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestPostgresCreateDevice tests device creation with PostgreSQL.
func TestPostgresCreateDevice(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)

	server := NewDeviceServer(store)
	ctx := context.Background()

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
			description: "should successfully create a valid device in PostgreSQL",
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
			name: "with_complex_metadata",
			request: &pb.CreateDeviceRequest{
				Name: "Smart Thermostat",
				Type: "hvac",
				Metadata: map[string]string{
					"floor":    "2",
					"building": "A",
					"zone":     "north",
					"model":    "TH-2000",
					"version":  "1.2.3",
				},
			},
			wantErr:     false,
			description: "should create device with complex metadata in JSONB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.CreateDevice(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: expected error, got none", tt.description)
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.wantCode {
						t.Errorf("expected error code %v, got %v", tt.wantCode, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if resp.Device.Id == "" {
				t.Error("expected device ID, got empty string")
			}
			if resp.Device.Name != tt.request.Name {
				t.Errorf("expected name %s, got %s", tt.request.Name, resp.Device.Name)
			}
			if resp.Device.Type != tt.request.Type {
				t.Errorf("expected type %s, got %s", tt.request.Type, resp.Device.Type)
			}
			if resp.Device.Status != pb.DeviceStatus_ONLINE {
				t.Errorf("expected status ONLINE, got %v", resp.Device.Status)
			}

			// Verify metadata
			if len(resp.Device.Metadata) != len(tt.request.Metadata) {
				t.Errorf("expected %d metadata entries, got %d", len(tt.request.Metadata), len(resp.Device.Metadata))
			}

			// Verify device was actually persisted
			getResp, err := server.GetDevice(ctx, &pb.GetDeviceRequest{Id: resp.Device.Id})
			if err != nil {
				t.Errorf("failed to retrieve created device: %v", err)
				return
			}
			if getResp.Device.Name != tt.request.Name {
				t.Errorf("persisted device name mismatch: expected %s, got %s", tt.request.Name, getResp.Device.Name)
			}
		})
	}
}

// TestPostgresCRUDOperations tests full CRUD cycle with PostgreSQL.
func TestPostgresCRUDOperations(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)

	server := NewDeviceServer(store)
	ctx := context.Background()

	// Create
	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Integration Test Device",
		Type: "sensor",
		Metadata: map[string]string{
			"test": "true",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	deviceID := createResp.Device.Id

	// Read
	getResp, err := server.GetDevice(ctx, &pb.GetDeviceRequest{Id: deviceID})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if getResp.Device.Name != "Integration Test Device" {
		t.Errorf("expected name 'Integration Test Device', got %s", getResp.Device.Name)
	}

	// Update
	updateResp, err := server.UpdateDevice(ctx, &pb.UpdateDeviceRequest{
		Id:     deviceID,
		Name:   "Updated Device",
		Status: pb.DeviceStatus_OFFLINE,
		Metadata: map[string]string{
			"test":    "true",
			"updated": "true",
		},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updateResp.Device.Name != "Updated Device" {
		t.Errorf("expected name 'Updated Device', got %s", updateResp.Device.Name)
	}
	if updateResp.Device.Status != pb.DeviceStatus_OFFLINE {
		t.Errorf("expected status OFFLINE, got %v", updateResp.Device.Status)
	}

	// Verify update persistence
	getResp2, err := server.GetDevice(ctx, &pb.GetDeviceRequest{Id: deviceID})
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if getResp2.Device.Name != "Updated Device" {
		t.Errorf("update not persisted: expected 'Updated Device', got %s", getResp2.Device.Name)
	}

	// Delete
	_, err = server.DeleteDevice(ctx, &pb.DeleteDeviceRequest{Id: deviceID})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = server.GetDevice(ctx, &pb.GetDeviceRequest{Id: deviceID})
	if err == nil {
		t.Error("expected error when getting deleted device, got none")
	}
	if st, ok := status.FromError(err); ok {
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound error, got %v", st.Code())
		}
	}
}

// TestPostgresListDevices tests listing with pagination.
func TestPostgresListDevices(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)

	server := NewDeviceServer(store)
	ctx := context.Background()

	// Create multiple devices
	deviceIDs := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		resp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
			Name: fmt.Sprintf("Device %d", i),
			Type: "sensor",
		})
		if err != nil {
			t.Fatalf("Failed to create device %d: %v", i, err)
		}
		deviceIDs = append(deviceIDs, resp.Device.Id)
	}

	// Test pagination
	t.Run("first_page", func(t *testing.T) {
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     1,
			PageSize: 5,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(resp.Devices) != 5 {
			t.Errorf("expected 5 devices, got %d", len(resp.Devices))
		}
		if resp.Total != 10 {
			t.Errorf("expected total 10, got %d", resp.Total)
		}
	})

	t.Run("second_page", func(t *testing.T) {
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     2,
			PageSize: 5,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(resp.Devices) != 5 {
			t.Errorf("expected 5 devices, got %d", len(resp.Devices))
		}
	})

	t.Run("all_devices", func(t *testing.T) {
		resp, err := server.ListDevices(ctx, &pb.ListDevicesRequest{
			Page:     1,
			PageSize: 100,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(resp.Devices) != 10 {
			t.Errorf("expected 10 devices, got %d", len(resp.Devices))
		}
	})
}

// TestPostgresTransactionConsistency tests that operations maintain consistency.
func TestPostgresTransactionConsistency(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)

	server := NewDeviceServer(store)
	ctx := context.Background()

	// Create device
	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Consistency Test Device",
		Type: "sensor",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	deviceID := createResp.Device.Id

	// Perform multiple updates rapidly
	for i := 0; i < 5; i++ {
		_, err := server.UpdateDevice(ctx, &pb.UpdateDeviceRequest{
			Id:   deviceID,
			Name: fmt.Sprintf("Device Update %d", i),
			Metadata: map[string]string{
				"iteration": fmt.Sprintf("%d", i),
			},
		})
		if err != nil {
			t.Fatalf("Update %d failed: %v", i, err)
		}
	}

	// Verify final state
	getResp, err := server.GetDevice(ctx, &pb.GetDeviceRequest{Id: deviceID})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if getResp.Device.Name != "Device Update 4" {
		t.Errorf("expected final name 'Device Update 4', got %s", getResp.Device.Name)
	}
	if getResp.Device.Metadata["iteration"] != "4" {
		t.Errorf("expected iteration '4', got %s", getResp.Device.Metadata["iteration"])
	}
}

// TestPostgresTimestamps tests that timestamps are correctly stored and retrieved.
func TestPostgresTimestamps(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)

	server := NewDeviceServer(store)
	ctx := context.Background()

	beforeCreate := time.Now().Unix()

	createResp, err := server.CreateDevice(ctx, &pb.CreateDeviceRequest{
		Name: "Timestamp Test Device",
		Type: "sensor",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	afterCreate := time.Now().Unix()

	// Check created_at is within reasonable range
	if createResp.Device.CreatedAt < beforeCreate || createResp.Device.CreatedAt > afterCreate {
		t.Errorf("created_at timestamp out of range: expected between %d and %d, got %d",
			beforeCreate, afterCreate, createResp.Device.CreatedAt)
	}

	// Check last_seen is set
	if createResp.Device.LastSeen < beforeCreate || createResp.Device.LastSeen > afterCreate {
		t.Errorf("last_seen timestamp out of range: expected between %d and %d, got %d",
			beforeCreate, afterCreate, createResp.Device.LastSeen)
	}

	// Wait a bit to ensure different timestamp (1+ second)
	time.Sleep(1100 * time.Millisecond)
	beforeUpdate := time.Now().Unix()

	updateResp, err := server.UpdateDevice(ctx, &pb.UpdateDeviceRequest{
		Id:   createResp.Device.Id,
		Name: "Updated Device",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// last_seen should be updated
	if updateResp.Device.LastSeen <= createResp.Device.LastSeen {
		t.Errorf("last_seen not updated: was %d, now %d", createResp.Device.LastSeen, updateResp.Device.LastSeen)
	}

	// created_at should remain unchanged
	if updateResp.Device.CreatedAt != createResp.Device.CreatedAt {
		t.Errorf("created_at changed: was %d, now %d", createResp.Device.CreatedAt, updateResp.Device.CreatedAt)
	}

	if updateResp.Device.LastSeen < beforeUpdate {
		t.Errorf("last_seen not properly updated: expected >= %d, got %d", beforeUpdate, updateResp.Device.LastSeen)
	}
}
