module github.com/yourusername/iot-platform/services/api-gateway

go 1.24.0

require (
	github.com/99designs/gqlgen v0.17.85
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/vektah/gqlparser/v2 v2.5.31
	github.com/yourusername/iot-platform/shared/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.78.0
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/yourusername/iot-platform/shared/proto => ../../shared/proto
