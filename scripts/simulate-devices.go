// Script to simulate IoT devices sending telemetry data via MQTT
//
// Usage:
//   go run scripts/simulate-devices.go [flags]
//
// Flags:
//   -broker    MQTT broker URL (default: tcp://localhost:1883)
//   -devices   Number of devices to simulate (default: 5)
//   -interval  Interval between messages in seconds (default: 5)
//   -duration  Duration to run in seconds, 0 for infinite (default: 0)
//
// Example:
//   go run scripts/simulate-devices.go -devices 10 -interval 2 -duration 60

package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// generateUUID generates a random UUID v4
func generateUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	// Set version (4) and variant (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// TelemetryMessage is the JSON payload sent to MQTT
type TelemetryMessage struct {
	DeviceID  string    `json:"device_id"`
	Timestamp string    `json:"timestamp"`
	Metrics   []Metric  `json:"metrics"`
}

// Metric represents a single sensor reading
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// DeviceSimulator simulates a single IoT device
type DeviceSimulator struct {
	ID       string
	Type     string
	client   mqtt.Client
	stopChan chan struct{}
}

// DeviceType defines the type of sensors a device has
type DeviceType struct {
	Name    string
	Metrics []MetricSpec
}

// MetricSpec defines how to generate a metric value
type MetricSpec struct {
	Name     string
	Unit     string
	BaseVal  float64
	Variance float64
}

var deviceTypes = []DeviceType{
	{
		Name: "temperature_sensor",
		Metrics: []MetricSpec{
			{Name: "temperature", Unit: "°C", BaseVal: 22.0, Variance: 5.0},
			{Name: "humidity", Unit: "%", BaseVal: 50.0, Variance: 20.0},
		},
	},
	{
		Name: "air_quality_sensor",
		Metrics: []MetricSpec{
			{Name: "co2", Unit: "ppm", BaseVal: 400.0, Variance: 100.0},
			{Name: "pm25", Unit: "µg/m³", BaseVal: 10.0, Variance: 15.0},
			{Name: "voc", Unit: "ppb", BaseVal: 100.0, Variance: 50.0},
		},
	},
	{
		Name: "power_meter",
		Metrics: []MetricSpec{
			{Name: "power", Unit: "W", BaseVal: 1500.0, Variance: 500.0},
			{Name: "voltage", Unit: "V", BaseVal: 230.0, Variance: 10.0},
			{Name: "current", Unit: "A", BaseVal: 6.5, Variance: 2.0},
		},
	},
	{
		Name: "weather_station",
		Metrics: []MetricSpec{
			{Name: "temperature", Unit: "°C", BaseVal: 18.0, Variance: 10.0},
			{Name: "humidity", Unit: "%", BaseVal: 60.0, Variance: 25.0},
			{Name: "pressure", Unit: "hPa", BaseVal: 1013.0, Variance: 20.0},
			{Name: "wind_speed", Unit: "m/s", BaseVal: 5.0, Variance: 10.0},
		},
	},
	{
		Name: "motion_sensor",
		Metrics: []MetricSpec{
			{Name: "motion_detected", Unit: "", BaseVal: 0.0, Variance: 1.0},
			{Name: "light_level", Unit: "lux", BaseVal: 300.0, Variance: 500.0},
		},
	},
}

func main() {
	// Parse flags
	brokerURL := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	numDevices := flag.Int("devices", 5, "Number of devices to simulate")
	interval := flag.Int("interval", 5, "Interval between messages in seconds")
	duration := flag.Int("duration", 0, "Duration to run in seconds (0 for infinite)")
	flag.Parse()

	log.Printf("🚀 IoT Device Simulator")
	log.Printf("   Broker: %s", *brokerURL)
	log.Printf("   Devices: %d", *numDevices)
	log.Printf("   Interval: %ds", *interval)
	if *duration > 0 {
		log.Printf("   Duration: %ds", *duration)
	} else {
		log.Printf("   Duration: infinite (Ctrl+C to stop)")
	}
	log.Println()

	// Connect to MQTT broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*brokerURL)
	opts.SetClientID(fmt.Sprintf("device-simulator-%d", time.Now().Unix()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Failed to connect to MQTT broker: %v", token.Error())
	}
	log.Printf("✅ Connected to MQTT broker")

	// Create device simulators
	simulators := make([]*DeviceSimulator, *numDevices)
	for i := 0; i < *numDevices; i++ {
		deviceType := deviceTypes[mathrand.Intn(len(deviceTypes))]
		simulators[i] = &DeviceSimulator{
			ID:       generateUUID(),
			Type:     deviceType.Name,
			client:   client,
			stopChan: make(chan struct{}),
		}
		log.Printf("📱 Created device: %s (type: %s)", simulators[i].ID, simulators[i].Type)
	}

	// Start simulators
	for _, sim := range simulators {
		go sim.Run(time.Duration(*interval) * time.Second)
	}
	log.Println()
	log.Printf("📡 Sending telemetry data...")
	log.Println()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *duration > 0 {
		select {
		case <-sigChan:
			log.Println("\n⏹️  Stopping simulation...")
		case <-time.After(time.Duration(*duration) * time.Second):
			log.Println("\n⏱️  Duration reached, stopping...")
		}
	} else {
		<-sigChan
		log.Println("\n⏹️  Stopping simulation...")
	}

	// Stop all simulators
	for _, sim := range simulators {
		close(sim.stopChan)
	}

	// Disconnect
	client.Disconnect(1000)
	log.Println("✅ Simulation stopped")
}

// Run starts the device simulation loop
func (d *DeviceSimulator) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Get device type spec
	var deviceType DeviceType
	for _, dt := range deviceTypes {
		if dt.Name == d.Type {
			deviceType = dt
			break
		}
	}

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			d.sendTelemetry(deviceType)
		}
	}
}

func (d *DeviceSimulator) sendTelemetry(deviceType DeviceType) {
	metrics := make([]Metric, len(deviceType.Metrics))
	for i, spec := range deviceType.Metrics {
		value := spec.BaseVal + (mathrand.Float64()*2-1)*spec.Variance
		if spec.Name == "motion_detected" {
			// Binary value for motion
			if mathrand.Float64() > 0.8 {
				value = 1.0
			} else {
				value = 0.0
			}
		}
		metrics[i] = Metric{
			Name:  spec.Name,
			Value: value,
			Unit:  spec.Unit,
		}
	}

	msg := TelemetryMessage{
		DeviceID:  d.ID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metrics:   metrics,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ [%s] Failed to marshal JSON: %v", d.ID, err)
		return
	}

	topic := fmt.Sprintf("devices/%s/telemetry", d.ID)
	token := d.client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ [%s] Failed to publish: %v", d.ID, token.Error())
		return
	}

	// Log first metric for visibility
	log.Printf("📤 [%s] %s=%.2f%s", d.ID, metrics[0].Name, metrics[0].Value, metrics[0].Unit)
}
