// Package main implements the Data Collector service.
// Microservice for ingesting IoT device telemetry data via MQTT.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"

	"github.com/yourusername/iot-platform/services/data-collector/mqtt"
	"github.com/yourusername/iot-platform/services/data-collector/publisher"
	"github.com/yourusername/iot-platform/services/data-collector/storage"
	pb "github.com/yourusername/iot-platform/shared/proto/telemetry"
)

// TelemetryServer implements pb.TelemetryServiceServer interface.
type TelemetryServer struct {
	pb.UnimplementedTelemetryServiceServer
	storage storage.Storage
}

// NewTelemetryServer creates a new server instance with the given storage backend.
func NewTelemetryServer(store storage.Storage) *TelemetryServer {
	return &TelemetryServer{
		storage: store,
	}
}

// GetTelemetry retrieves telemetry data for a device within a time range.
func (s *TelemetryServer) GetTelemetry(ctx context.Context, req *pb.GetTelemetryRequest) (*pb.GetTelemetryResponse, error) {
	log.Printf("📥 GetTelemetry: device=%s, metric=%s", req.DeviceId, req.MetricName)

	points, err := s.storage.GetTelemetry(ctx, req.DeviceId, req.MetricName, req.FromTime, req.ToTime, int(req.Limit))
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Found %d telemetry points", len(points))
	return &pb.GetTelemetryResponse{Points: points}, nil
}

// GetTelemetryAggregated retrieves aggregated telemetry data.
func (s *TelemetryServer) GetTelemetryAggregated(ctx context.Context, req *pb.GetTelemetryAggregatedRequest) (*pb.GetTelemetryAggregatedResponse, error) {
	log.Printf("📥 GetTelemetryAggregated: device=%s, metric=%s, interval=%s", req.DeviceId, req.MetricName, req.Interval)

	aggregations, err := s.storage.GetTelemetryAggregated(ctx, req.DeviceId, req.MetricName, req.FromTime, req.ToTime, req.Interval)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Found %d aggregation buckets", len(aggregations))
	return &pb.GetTelemetryAggregatedResponse{Aggregations: aggregations}, nil
}

// GetLatestMetric retrieves the latest value for a specific metric.
func (s *TelemetryServer) GetLatestMetric(ctx context.Context, req *pb.GetLatestMetricRequest) (*pb.GetLatestMetricResponse, error) {
	log.Printf("📥 GetLatestMetric: device=%s, metric=%s", req.DeviceId, req.MetricName)

	point, err := s.storage.GetLatestMetric(ctx, req.DeviceId, req.MetricName)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Latest value: %v", point.Value)
	return &pb.GetLatestMetricResponse{Point: point}, nil
}

// GetDeviceMetrics retrieves all available metrics for a device.
func (s *TelemetryServer) GetDeviceMetrics(ctx context.Context, req *pb.GetDeviceMetricsRequest) (*pb.GetDeviceMetricsResponse, error) {
	log.Printf("📥 GetDeviceMetrics: device=%s", req.DeviceId)

	metrics, err := s.storage.GetDeviceMetrics(ctx, req.DeviceId)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Found %d metrics", len(metrics))
	return &pb.GetDeviceMetricsResponse{Metrics: metrics}, nil
}

// main initializes and starts the Telemetry Collector service.
//
// Configuration via environment variables:
//   - TELEMETRY_GRPC_PORT: gRPC server port (default: 8083)
//   - MQTT_BROKER: MQTT broker URL (default: tcp://localhost:1883)
//   - MQTT_CLIENT_ID: MQTT client ID (default: data-collector)
//   - MQTT_TOPIC: MQTT topic pattern (default: devices/+/telemetry)
//   - DB_HOST: PostgreSQL host (default: localhost)
//   - DB_PORT: PostgreSQL port (default: 5432)
//   - DB_NAME: Database name (default: iot_platform)
//   - DB_USER: Database user (default: iot_user)
//   - DB_PASSWORD: Database password (default: iot_password)
//   - DB_SSLMODE: SSL mode (default: disable)
//   - REDIS_HOST: Redis host (default: localhost)
//   - REDIS_PORT: Redis port (default: 6379)
//   - REDIS_PASSWORD: Redis password (default: "")
//   - REDIS_DB: Redis database (default: 0)
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcPort := getEnvInt("TELEMETRY_GRPC_PORT", 8083)

	// Build PostgreSQL DSN
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DB_USER", "iot_user"),
		getEnv("DB_PASSWORD", "iot_password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "iot_platform"),
		getEnv("DB_SSLMODE", "disable"),
	)

	// Initialize storage
	store, err := storage.NewTimescaleStorage(ctx, dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to TimescaleDB: %v", err)
	}
	defer store.Close()
	log.Printf("✅ Connected to TimescaleDB")

	// Initialize Redis publisher
	redisPublisher, err := publisher.NewRedisPublisher(ctx, publisher.Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnvInt("REDIS_PORT", 6379),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvInt("REDIS_DB", 0),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redisPublisher.Close()

	// Initialize MQTT client
	mqttBroker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	mqttClientID := getEnv("MQTT_CLIENT_ID", "data-collector")
	mqttTopic := getEnv("MQTT_TOPIC", "devices/+/telemetry")

	mqttClient, err := mqtt.NewClient(mqtt.Config{
		BrokerURL: mqttBroker,
		ClientID:  mqttClientID,
		Topic:     mqttTopic,
		OnMessage: func(deviceID, metricName string, value float64, unit string, timestamp int64, metadata map[string]string) {
			if err := store.InsertTelemetry(ctx, deviceID, metricName, value, unit, timestamp, metadata); err != nil {
				log.Printf("❌ Failed to insert telemetry: %v", err)
				return
			}
			// Publish to Redis after successful DB insert
			if err := redisPublisher.PublishTelemetry(ctx, deviceID, metricName, value, unit, timestamp); err != nil {
				log.Printf("⚠️ Failed to publish to Redis: %v", err)
			}
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to create MQTT client: %v", err)
	}

	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()
	log.Printf("✅ Connected to MQTT broker: %s", mqttBroker)

	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("❌ Failed to subscribe to MQTT topic: %v", err)
	}
	log.Printf("✅ Subscribed to topic: %s", mqttTopic)

	// Start gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("❌ Failed to create listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	telemetryServer := NewTelemetryServer(store)
	pb.RegisterTelemetryServiceServer(grpcServer, telemetryServer)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("⏳ Shutting down gracefully...")
		grpcServer.GracefulStop()
		mqttClient.Disconnect()
		redisPublisher.Close()
		store.Close()
		cancel()
	}()

	log.Println("=====================================")
	log.Printf("Data Collector Service")
	log.Println("=====================================")
	log.Printf("gRPC Port: %d", grpcPort)
	log.Printf("MQTT Broker: %s", mqttBroker)
	log.Printf("MQTT Topic: %s", mqttTopic)
	log.Printf("Database: TimescaleDB")
	log.Printf("Redis: %s:%d", getEnv("REDIS_HOST", "localhost"), getEnvInt("REDIS_PORT", 6379))
	log.Println("-------------------------------------")
	log.Printf("✅ Server started")
	log.Printf("⏳ Waiting for telemetry data...")
	log.Println("=====================================")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as an integer or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
