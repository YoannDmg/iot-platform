package graph

import (
	pb "github.com/yourusername/iot-platform/shared/proto"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// Resolver holds dependencies for GraphQL resolvers.
type Resolver struct {
	DeviceClient pb.DeviceServiceClient
}
