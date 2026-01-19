package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
)

// TelemetryEvent represents the JSON format from Redis pub/sub
type TelemetryEvent struct {
	DeviceID   string  `json:"device_id"`
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  string  `json:"timestamp"`
}

// RedisSubscriber listens to Redis pub/sub and dispatches to the broker
type RedisSubscriber struct {
	client *redis.Client
	broker *Broker
	cancel context.CancelFunc
}

// Config holds Redis connection configuration
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisSubscriber creates and starts a new Redis subscriber
func NewRedisSubscriber(ctx context.Context, cfg Config, broker *Broker) (*RedisSubscriber, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  0, // No timeout for pub/sub
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Connected to Redis at %s:%d", cfg.Host, cfg.Port)

	subCtx, cancel := context.WithCancel(ctx)

	subscriber := &RedisSubscriber{
		client: client,
		broker: broker,
		cancel: cancel,
	}

	// Start listening in background
	go subscriber.listen(subCtx)

	return subscriber, nil
}

// listen subscribes to Redis channels and dispatches messages to the broker
func (s *RedisSubscriber) listen(ctx context.Context) {
	// Subscribe to all telemetry channels using pattern
	pubsub := s.client.PSubscribe(ctx, "iot:telemetry:*")
	defer pubsub.Close()

	log.Printf("📡 Subscribed to Redis pattern: iot:telemetry:*")

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			log.Printf("⏹️ Redis subscriber stopped")
			return
		case msg, ok := <-ch:
			if !ok {
				log.Printf("⚠️ Redis pub/sub channel closed")
				return
			}
			s.handleMessage(msg)
		}
	}
}

// handleMessage processes a Redis pub/sub message
func (s *RedisSubscriber) handleMessage(msg *redis.Message) {
	var event TelemetryEvent
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		log.Printf("⚠️ Failed to unmarshal telemetry event: %v", err)
		return
	}

	// Parse timestamp to Unix
	var unixTime int
	timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		unixTime = int(time.Now().Unix())
	} else {
		unixTime = int(timestamp.Unix())
	}

	// Convert to GraphQL model
	unit := event.Unit
	point := &model.TelemetryPoint{
		Time:  unixTime,
		Value: event.Value,
		Unit:  &unit,
	}

	// Dispatch to broker
	s.broker.Publish(event.DeviceID, point)
}

// Close stops the subscriber and closes the Redis connection
func (s *RedisSubscriber) Close() error {
	s.cancel()
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
