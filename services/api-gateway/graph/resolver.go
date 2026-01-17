package graph

import (
	"github.com/yourusername/iot-platform/services/api-gateway/auth"
	devicepb "github.com/yourusername/iot-platform/shared/proto/device"
	telemetrypb "github.com/yourusername/iot-platform/shared/proto/telemetry"
	userpb "github.com/yourusername/iot-platform/shared/proto/user"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// Resolver holds dependencies for GraphQL resolvers.
type Resolver struct {
	DeviceClient    devicepb.DeviceServiceClient
	UserClient      userpb.UserServiceClient
	TelemetryClient telemetrypb.TelemetryServiceClient
	JWTManager      *auth.JWTManager
}
