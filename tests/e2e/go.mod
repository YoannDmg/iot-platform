module github.com/yourusername/iot-platform/tests/e2e

go 1.24.0

require github.com/eclipse/paho.mqtt.golang v1.5.0

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
)

replace (
	github.com/yourusername/iot-platform/services/api-gateway => ../../services/api-gateway
	github.com/yourusername/iot-platform/services/device-manager => ../../services/device-manager
	github.com/yourusername/iot-platform/services/telemetry-collector => ../../services/telemetry-collector
	github.com/yourusername/iot-platform/services/user-service => ../../services/user-service
	github.com/yourusername/iot-platform/shared => ../../shared
)
