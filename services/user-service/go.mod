module github.com/yourusername/iot-platform/services/user-service

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/yourusername/iot-platform/shared/proto v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.44.0
	google.golang.org/grpc v1.78.0
)

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/yourusername/iot-platform/shared/proto => ../../shared/proto
