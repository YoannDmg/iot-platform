package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// TelemetryEvent represents a telemetry data point published to Redis
type TelemetryEvent struct {
	DeviceID   string  `json:"device_id"`
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  string  `json:"timestamp"`
}

// RedisPublisher handles publishing telemetry data to Redis Pub/Sub
type RedisPublisher struct {
	client *redis.Client
}

// Config holds Redis connection configuration
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisPublisher creates a new Redis publisher
func NewRedisPublisher(ctx context.Context, cfg Config) (*RedisPublisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Connected to Redis at %s:%d", cfg.Host, cfg.Port)

	return &RedisPublisher{client: client}, nil
}

// PublishTelemetry publishes a telemetry event to Redis
func (p *RedisPublisher) PublishTelemetry(ctx context.Context, deviceID, metricName string, value float64, unit string, timestamp int64) error {
	event := TelemetryEvent{
		DeviceID:   deviceID,
		MetricName: metricName,
		Value:      value,
		Unit:       unit,
		Timestamp:  time.Unix(timestamp, 0).UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry event: %w", err)
	}

	// Publish to device-specific channel: iot:telemetry:{device_id}
	channel := fmt.Sprintf("iot:telemetry:%s", deviceID)

	if err := p.client.Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (p *RedisPublisher) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
