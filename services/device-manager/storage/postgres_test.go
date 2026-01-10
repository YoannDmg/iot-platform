// +build integration

package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	pb "github.com/yourusername/iot-platform/shared/proto/device"
)

func init() {
	// Load .env from project root (2 levels up from storage/)
	envPath := filepath.Join("..", "..", "..", ".env")
	_ = godotenv.Load(envPath) // Ignore error if .env doesn't exist
}

// setupPostgresStorage creates a PostgreSQL storage for testing.
// Requires PostgreSQL to be running (via docker-compose).
func setupPostgresStorage(t *testing.T) *PostgresStorage {
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

	store, err := NewPostgresStorage(context.Background(), dsn)
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
func cleanDatabase(t *testing.T, store Storage) {
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

func TestPostgresStorage_CreateDevice(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	device := &pb.Device{
		Id:        uuid.New().String(),
		Name:      "PostgreSQL Test Device",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
		Metadata: map[string]string{
			"location": "datacenter-1",
			"rack":     "A-01",
		},
	}

	created, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	if created.Id != device.Id {
		t.Errorf("Id = %v, want %v", created.Id, device.Id)
	}
	if created.Name != device.Name {
		t.Errorf("Name = %v, want %v", created.Name, device.Name)
	}
	if len(created.Metadata) != 2 {
		t.Errorf("Metadata count = %v, want 2", len(created.Metadata))
	}
}

func TestPostgresStorage_GetDevice(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	deviceID := uuid.New().String()
	device := &pb.Device{
		Id:        deviceID,
		Name:      "Get Test Device",
		Type:      "actuator",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}

	_, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Get existing device
	retrieved, err := store.GetDevice(ctx, deviceID)
	if err != nil {
		t.Fatalf("GetDevice() failed: %v", err)
	}

	if retrieved.Id != device.Id {
		t.Errorf("Id = %v, want %v", retrieved.Id, device.Id)
	}
	if retrieved.Name != device.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, device.Name)
	}

	// Get non-existent device
	_, err = store.GetDevice(ctx, uuid.New().String())
	if err == nil {
		t.Error("GetDevice() should fail for non-existent device")
	}
}

func TestPostgresStorage_UpdateDevice(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	device := &pb.Device{
		Id:        uuid.New().String(),
		Name:      "Original Name",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
		Metadata: map[string]string{
			"version": "1.0",
		},
	}

	_, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Update device
	device.Name = "Updated Name PostgreSQL"
	device.Status = pb.DeviceStatus_MAINTENANCE
	device.Metadata["version"] = "2.0"
	device.Metadata["env"] = "prod"

	updated, err := store.UpdateDevice(ctx, device)
	if err != nil {
		t.Fatalf("UpdateDevice() failed: %v", err)
	}

	if updated.Name != "Updated Name PostgreSQL" {
		t.Errorf("Name = %v, want 'Updated Name PostgreSQL'", updated.Name)
	}
	if updated.Status != pb.DeviceStatus_MAINTENANCE {
		t.Errorf("Status = %v, want MAINTENANCE", updated.Status)
	}
	if updated.Metadata["version"] != "2.0" {
		t.Errorf("Metadata[version] = %v, want '2.0'", updated.Metadata["version"])
	}
}

func TestPostgresStorage_DeleteDevice(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	deviceID := uuid.New().String()
	device := &pb.Device{
		Id:        deviceID,
		Name:      "Delete Test Device",
		Type:      "gateway",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}

	_, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Delete device
	err = store.DeleteDevice(ctx, deviceID)
	if err != nil {
		t.Fatalf("DeleteDevice() failed: %v", err)
	}

	// Verify device is deleted
	_, err = store.GetDevice(ctx, deviceID)
	if err == nil {
		t.Error("GetDevice() should fail after deletion")
	}
}

func TestPostgresStorage_ListDevices(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	// Create multiple devices
	for i := 0; i < 7; i++ {
		device := &pb.Device{
			Id:        uuid.New().String(),
			Name:      fmt.Sprintf("List Device %d", i),
			Type:      "sensor",
			Status:    pb.DeviceStatus_ONLINE,
			CreatedAt: time.Now().Unix(),
			LastSeen:  time.Now().Unix(),
		}
		_, err := store.CreateDevice(ctx, device)
		if err != nil {
			t.Fatalf("CreateDevice(%d) failed: %v", i, err)
		}
	}

	// List all devices
	devices, total, err := store.ListDevices(ctx, 1, 10)
	if err != nil {
		t.Fatalf("ListDevices() failed: %v", err)
	}

	if total != 7 {
		t.Errorf("Total = %d, want 7", total)
	}
	if len(devices) != 7 {
		t.Errorf("Devices count = %d, want 7", len(devices))
	}

	// Test pagination
	devices, total, err = store.ListDevices(ctx, 1, 3)
	if err != nil {
		t.Fatalf("ListDevices(page=1, pageSize=3) failed: %v", err)
	}

	if total != 7 {
		t.Errorf("Total = %d, want 7", total)
	}
	if len(devices) != 3 {
		t.Errorf("Devices count = %d, want 3", len(devices))
	}

	// Second page
	devices, total, err = store.ListDevices(ctx, 2, 3)
	if err != nil {
		t.Fatalf("ListDevices(page=2, pageSize=3) failed: %v", err)
	}

	if total != 7 {
		t.Errorf("Total = %d, want 7", total)
	}
	if len(devices) != 3 {
		t.Errorf("Devices count = %d, want 3", len(devices))
	}
}

func TestPostgresStorage_ListDevicesByType(t *testing.T) {
	t.Skip("ListDevicesByType not implemented yet - will be added in future PR")
}

func TestPostgresStorage_MetadataJSONB(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	deviceID := uuid.New().String()
	// Create device with complex metadata
	device := &pb.Device{
		Id:        deviceID,
		Name:      "Metadata Test",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
		Metadata: map[string]string{
			"location":  "datacenter-1",
			"rack":      "A-01",
			"floor":     "2",
			"building":  "North",
			"zone":      "production",
			"serial":    "SN-12345",
			"firmware":  "v2.4.1",
			"installed": "2024-01-15",
		},
	}

	_, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Retrieve and verify metadata
	retrieved, err := store.GetDevice(ctx, deviceID)
	if err != nil {
		t.Fatalf("GetDevice() failed: %v", err)
	}

	if len(retrieved.Metadata) != 8 {
		t.Errorf("Metadata count = %d, want 8", len(retrieved.Metadata))
	}

	for key, expectedValue := range device.Metadata {
		if actualValue, ok := retrieved.Metadata[key]; !ok {
			t.Errorf("Missing metadata key: %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("Metadata[%s] = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

func TestPostgresStorage_TimestampPersistence(t *testing.T) {
	store := setupPostgresStorage(t)
	cleanDatabase(t, store)
	ctx := context.Background()

	deviceID := uuid.New().String()
	createdAt := time.Now().Unix()
	lastSeen := time.Now().Unix()

	device := &pb.Device{
		Id:        deviceID,
		Name:      "Timestamp Test",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: createdAt,
		LastSeen:  lastSeen,
	}

	_, err := store.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Retrieve and verify timestamps
	retrieved, err := store.GetDevice(ctx, deviceID)
	if err != nil {
		t.Fatalf("GetDevice() failed: %v", err)
	}

	// Allow 1 second tolerance for timestamp comparison
	if retrieved.CreatedAt < createdAt-1 || retrieved.CreatedAt > createdAt+1 {
		t.Errorf("CreatedAt = %d, want %d (±1s)", retrieved.CreatedAt, createdAt)
	}
	if retrieved.LastSeen < lastSeen-1 || retrieved.LastSeen > lastSeen+1 {
		t.Errorf("LastSeen = %d, want %d (±1s)", retrieved.LastSeen, lastSeen)
	}
}
