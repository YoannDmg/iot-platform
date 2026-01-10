module github.com/yourusername/iot-platform/tests/e2e

go 1.21

require (
	github.com/yourusername/iot-platform/services/api-gateway v0.0.0
	github.com/yourusername/iot-platform/services/device-manager v0.0.0
	github.com/yourusername/iot-platform/services/user-service v0.0.0
)

replace (
	github.com/yourusername/iot-platform/services/api-gateway => ../../services/api-gateway
	github.com/yourusername/iot-platform/services/device-manager => ../../services/device-manager
	github.com/yourusername/iot-platform/services/user-service => ../../services/user-service
	github.com/yourusername/iot-platform/shared => ../../shared
)
