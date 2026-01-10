// Package grpc provides gRPC client connections to backend services.
package grpc

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	devicepb "github.com/yourusername/iot-platform/shared/proto/device"
	userpb "github.com/yourusername/iot-platform/shared/proto/user"
)

// DeviceClient wraps the gRPC client for Device Manager service.
type DeviceClient struct {
	conn   *grpc.ClientConn
	client devicepb.DeviceServiceClient
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
		client: devicepb.NewDeviceServiceClient(conn),
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
func (c *DeviceClient) GetClient() devicepb.DeviceServiceClient {
	return c.client
}

// UserClient wraps the gRPC client for User Service.
type UserClient struct {
	conn   *grpc.ClientConn
	client userpb.UserServiceClient
}

// NewUserClient creates a new gRPC client connection to User Service.
func NewUserClient(address string) (*UserClient, error) {
	// TODO Production: Add TLS credentials
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	log.Printf("✅ Connected to User Service at %s", address)

	return &UserClient{
		conn:   conn,
		client: userpb.NewUserServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetClient returns the underlying gRPC client.
func (c *UserClient) GetClient() userpb.UserServiceClient {
	return c.client
}
