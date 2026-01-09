// Package grpc provides gRPC client connections to backend services.
package grpc

import (
	"fmt"
	"log"

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
	// TODO Production: Add TLS credentials
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
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
