// Package mqtt provides MQTT client functionality for telemetry ingestion.
package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// TelemetryMessage represents the JSON payload from devices.
type TelemetryMessage struct {
	DeviceID  string    `json:"device_id"`
	Timestamp string    `json:"timestamp,omitempty"`
	Metrics   []Metric  `json:"metrics"`
}

// Metric represents a single metric measurement.
type Metric struct {
	Name     string            `json:"name"`
	Value    float64           `json:"value"`
	Unit     string            `json:"unit,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MessageHandler is called for each received telemetry point.
type MessageHandler func(deviceID, metricName string, value float64, unit string, timestamp int64, metadata map[string]string)

// Config holds MQTT client configuration.
type Config struct {
	BrokerURL string
	ClientID  string
	Topic     string
	Username  string
	Password  string
	OnMessage MessageHandler
}

// Client wraps the Paho MQTT client with telemetry-specific functionality.
type Client struct {
	config     Config
	pahoClient pahomqtt.Client
}

// NewClient creates a new MQTT client with the given configuration.
func NewClient(config Config) (*Client, error) {
	if config.BrokerURL == "" {
		return nil, fmt.Errorf("broker URL is required")
	}
	if config.ClientID == "" {
		config.ClientID = "data-collector"
	}
	if config.Topic == "" {
		config.Topic = "devices/+/telemetry"
	}
	if config.OnMessage == nil {
		return nil, fmt.Errorf("message handler is required")
	}

	return &Client{
		config: config,
	}, nil
}

// Connect establishes a connection to the MQTT broker.
func (c *Client) Connect() error {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(c.config.BrokerURL)
	opts.SetClientID(c.config.ClientID)

	if c.config.Username != "" {
		opts.SetUsername(c.config.Username)
		opts.SetPassword(c.config.Password)
	}

	// Connection settings
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)

	// Callbacks
	opts.SetOnConnectHandler(func(client pahomqtt.Client) {
		log.Printf("✅ MQTT connected to %s", c.config.BrokerURL)
		// Re-subscribe on reconnect
		if err := c.subscribe(); err != nil {
			log.Printf("❌ Failed to re-subscribe: %v", err)
		}
	})

	opts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		log.Printf("⚠️  MQTT connection lost: %v", err)
	})

	opts.SetReconnectingHandler(func(client pahomqtt.Client, opts *pahomqtt.ClientOptions) {
		log.Printf("⏳ MQTT reconnecting...")
	})

	c.pahoClient = pahomqtt.NewClient(opts)

	token := c.pahoClient.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect: %w", token.Error())
	}

	return nil
}

// Subscribe subscribes to the telemetry topic.
func (c *Client) Subscribe() error {
	return c.subscribe()
}

func (c *Client) subscribe() error {
	token := c.pahoClient.Subscribe(c.config.Topic, 1, c.handleMessage)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe: %w", token.Error())
	}
	return nil
}

// handleMessage processes incoming MQTT messages.
func (c *Client) handleMessage(client pahomqtt.Client, msg pahomqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	log.Printf("📨 Received message on topic: %s", topic)

	// Extract device ID from topic (devices/{device_id}/telemetry)
	deviceID := extractDeviceID(topic)
	if deviceID == "" {
		log.Printf("⚠️  Could not extract device ID from topic: %s", topic)
		return
	}

	// Parse JSON payload
	var telemetry TelemetryMessage
	if err := json.Unmarshal(payload, &telemetry); err != nil {
		log.Printf("❌ Failed to parse telemetry JSON: %v", err)
		return
	}

	// Use device ID from topic if not in payload
	if telemetry.DeviceID == "" {
		telemetry.DeviceID = deviceID
	}

	// Parse timestamp or use current time
	var timestamp int64
	if telemetry.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, telemetry.Timestamp)
		if err != nil {
			log.Printf("⚠️  Invalid timestamp format, using current time: %v", err)
			timestamp = time.Now().Unix()
		} else {
			timestamp = t.Unix()
		}
	} else {
		timestamp = time.Now().Unix()
	}

	// Process each metric
	for _, metric := range telemetry.Metrics {
		c.config.OnMessage(
			telemetry.DeviceID,
			metric.Name,
			metric.Value,
			metric.Unit,
			timestamp,
			metric.Metadata,
		)
		log.Printf("📊 Metric: device=%s, %s=%v %s", telemetry.DeviceID, metric.Name, metric.Value, metric.Unit)
	}
}

// extractDeviceID extracts the device ID from the MQTT topic.
// Expected format: devices/{device_id}/telemetry
func extractDeviceID(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) >= 3 && parts[0] == "devices" && parts[2] == "telemetry" {
		return parts[1]
	}
	return ""
}

// Disconnect gracefully disconnects from the MQTT broker.
func (c *Client) Disconnect() {
	if c.pahoClient != nil && c.pahoClient.IsConnected() {
		c.pahoClient.Disconnect(1000)
		log.Printf("✅ MQTT disconnected")
	}
}

// IsConnected returns true if the client is connected to the broker.
func (c *Client) IsConnected() bool {
	return c.pahoClient != nil && c.pahoClient.IsConnected()
}
