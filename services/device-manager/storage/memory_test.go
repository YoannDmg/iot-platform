// +build unit

package storage

import (
	"context"
	"testing"
	"time"

	pb "github.com/yourusername/iot-platform/shared/proto/device"
)

func TestMemoryStorage_CreateDevice(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	device := &pb.Device{
		Id:        "device-123",
		Name:      "Test Device",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
		Metadata: map[string]string{
			"location": "room-101",
		},
	}

	created, err := storage.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	if created.Id != device.Id {
		t.Errorf("Id = %v, want %v", created.Id, device.Id)
	}
	if created.Name != device.Name {
		t.Errorf("Name = %v, want %v", created.Name, device.Name)
	}
}

func TestMemoryStorage_GetDevice(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	device := &pb.Device{
		Id:        "device-123",
		Name:      "Test Device",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}

	_, err := storage.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Get existing device
	retrieved, err := storage.GetDevice(ctx, "device-123")
	if err != nil {
		t.Fatalf("GetDevice() failed: %v", err)
	}

	if retrieved.Id != device.Id {
		t.Errorf("Id = %v, want %v", retrieved.Id, device.Id)
	}

	// Get non-existent device
	_, err = storage.GetDevice(ctx, "non-existent")
	if err == nil {
		t.Error("GetDevice() should fail for non-existent device")
	}
}

func TestMemoryStorage_UpdateDevice(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	device := &pb.Device{
		Id:        "device-123",
		Name:      "Original Name",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}

	_, err := storage.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Update device
	device.Name = "Updated Name"
	device.Status = pb.DeviceStatus_OFFLINE

	updated, err := storage.UpdateDevice(ctx, device)
	if err != nil {
		t.Fatalf("UpdateDevice() failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Name = %v, want 'Updated Name'", updated.Name)
	}
	if updated.Status != pb.DeviceStatus_OFFLINE {
		t.Errorf("Status = %v, want OFFLINE", updated.Status)
	}
}

func TestMemoryStorage_DeleteDevice(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	device := &pb.Device{
		Id:        "device-123",
		Name:      "Test Device",
		Type:      "sensor",
		Status:    pb.DeviceStatus_ONLINE,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}

	_, err := storage.CreateDevice(ctx, device)
	if err != nil {
		t.Fatalf("CreateDevice() failed: %v", err)
	}

	// Delete device
	err = storage.DeleteDevice(ctx, "device-123")
	if err != nil {
		t.Fatalf("DeleteDevice() failed: %v", err)
	}

	// Verify device is deleted
	_, err = storage.GetDevice(ctx, "device-123")
	if err == nil {
		t.Error("GetDevice() should fail after deletion")
	}
}

func TestMemoryStorage_ListDevices(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()

	// Create multiple devices
	for i := 0; i < 5; i++ {
		device := &pb.Device{
			Id:        "device-" + string(rune('1'+i)),
			Name:      "Device " + string(rune('A'+i)),
			Type:      "sensor",
			Status:    pb.DeviceStatus_ONLINE,
			CreatedAt: time.Now().Unix(),
			LastSeen:  time.Now().Unix(),
		}
		_, err := storage.CreateDevice(ctx, device)
		if err != nil {
			t.Fatalf("CreateDevice() failed: %v", err)
		}
	}

	// List all devices
	devices, total, err := storage.ListDevices(ctx, 1, 10)
	if err != nil {
		t.Fatalf("ListDevices() failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Total = %d, want 5", total)
	}
	if len(devices) != 5 {
		t.Errorf("Devices count = %d, want 5", len(devices))
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	storage := NewMemoryStorage()
	err := storage.Close()
	if err != nil {
		t.Errorf("Close() should not return error: %v", err)
	}
}
