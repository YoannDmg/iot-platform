// Package grpc provides gRPC client connections to backend services.
package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/yourusername/iot-platform/shared/proto"
)

// DeviceClient wraps the gRPC client for Device Manager service.
type DeviceClient struct {
	conn   *grpc.ClientConn
	client pb.DeviceServiceClient
}

// NewDeviceClient creates a new gRPC client connection to Device Manager.
func NewDeviceClient(address string) (*DeviceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO Production: Add TLS credentials
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device manager: %w", err)
	}

	log.Printf("✅ Connected to Device Manager at %s", address)

	return &DeviceClient{
		conn:   conn,
		client: pb.NewDeviceServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *DeviceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetClient returns the underlying gRPC client.
func (c *DeviceClient) GetClient() pb.DeviceServiceClient {
	return c.client
}
