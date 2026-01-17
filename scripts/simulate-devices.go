// Script to simulate IoT devices sending telemetry data via MQTT
//
// Usage:
//   go run scripts/simulate-devices.go [flags]
//
// Flags:
//   -broker    MQTT broker URL (default: tcp://localhost:1883)
//   -api       GraphQL API URL (default: http://localhost:8080/query)
//   -devices   Number of devices to simulate (default: 5)
//   -interval  Interval between messages in seconds (default: 5)
//   -duration  Duration to run in seconds, 0 for infinite (default: 0)
//
// Example:
//   go run scripts/simulate-devices.go -devices 10 -interval 2 -duration 60

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// GraphQL request/response types
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type DevicesQueryResponse struct {
	Devices struct {
		Devices []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"devices"`
	} `json:"devices"`
}

type CreateDeviceResponse struct {
	CreateDevice struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"createDevice"`
}

// TelemetryMessage is the JSON payload sent to MQTT
type TelemetryMessage struct {
	DeviceID  string   `json:"device_id"`
	Timestamp string   `json:"timestamp"`
	Metrics   []Metric `json:"metrics"`
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
	Name     string
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

// Predefined simulated device names (cycling through device types)
func getSimulatedDeviceName(index int) (string, DeviceType) {
	typeIndex := index % len(deviceTypes)
	deviceType := deviceTypes[typeIndex]
	// sim-temperature_sensor-001, sim-air_quality_sensor-001, etc.
	typeCount := (index / len(deviceTypes)) + 1
	name := fmt.Sprintf("sim-%s-%03d", deviceType.Name, typeCount)
	return name, deviceType
}

// GraphQL client helper
func graphqlRequest(apiURL string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, fmt.Errorf("graphql error: %s", gqlResp.Errors[0].Message)
	}

	return &gqlResp, nil
}

// Fetch all existing devices to check for our simulated ones
func fetchExistingDevices(apiURL string) (map[string]string, error) {
	query := `
		query {
			devices(pageSize: 1000) {
				devices {
					id
					name
				}
			}
		}
	`

	resp, err := graphqlRequest(apiURL, query, nil)
	if err != nil {
		return nil, err
	}

	var data DevicesQueryResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse devices: %w", err)
	}

	// Map name -> id
	result := make(map[string]string)
	for _, d := range data.Devices.Devices {
		result[d.Name] = d.ID
	}

	return result, nil
}

// Create a device via GraphQL
func createDevice(apiURL string, name string, deviceType string) (string, error) {
	query := `
		mutation CreateDevice($input: CreateDeviceInput!) {
			createDevice(input: $input) {
				id
				name
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name": name,
			"type": deviceType,
			"metadata": []map[string]string{
				{"key": "simulated", "value": "true"},
				{"key": "created_by", "value": "simulate-devices.go"},
			},
		},
	}

	resp, err := graphqlRequest(apiURL, query, variables)
	if err != nil {
		return "", err
	}

	var data CreateDeviceResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("failed to parse create response: %w", err)
	}

	return data.CreateDevice.ID, nil
}

// ensureDeviceExists checks if device exists, creates if not, returns ID
func ensureDeviceExists(apiURL string, name string, deviceType string, existingDevices map[string]string) (string, bool, error) {
	if id, exists := existingDevices[name]; exists {
		return id, false, nil // already exists
	}

	id, err := createDevice(apiURL, name, deviceType)
	if err != nil {
		return "", false, err
	}

	return id, true, nil // newly created
}

func main() {
	// Parse flags
	brokerURL := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	apiURL := flag.String("api", "http://localhost:8080/query", "GraphQL API URL")
	numDevices := flag.Int("devices", 5, "Number of devices to simulate")
	interval := flag.Int("interval", 5, "Interval between messages in seconds")
	duration := flag.Int("duration", 0, "Duration to run in seconds (0 for infinite)")
	flag.Parse()

	log.Printf("🚀 IoT Device Simulator")
	log.Printf("   Broker: %s", *brokerURL)
	log.Printf("   API: %s", *apiURL)
	log.Printf("   Devices: %d", *numDevices)
	log.Printf("   Interval: %ds", *interval)
	if *duration > 0 {
		log.Printf("   Duration: %ds", *duration)
	} else {
		log.Printf("   Duration: infinite (Ctrl+C to stop)")
	}
	log.Println()

	// Fetch existing devices from API
	log.Printf("🔍 Checking existing devices...")
	existingDevices, err := fetchExistingDevices(*apiURL)
	if err != nil {
		log.Fatalf("❌ Failed to fetch existing devices: %v", err)
	}
	log.Printf("   Found %d existing devices", len(existingDevices))

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
	log.Println()

	// Create device simulators (ensure they exist in DB first)
	log.Printf("📱 Setting up simulated devices...")
	simulators := make([]*DeviceSimulator, *numDevices)
	createdCount := 0
	reusedCount := 0

	for i := 0; i < *numDevices; i++ {
		name, deviceType := getSimulatedDeviceName(i)

		deviceID, created, err := ensureDeviceExists(*apiURL, name, deviceType.Name, existingDevices)
		if err != nil {
			log.Fatalf("❌ Failed to ensure device %s exists: %v", name, err)
		}

		if created {
			createdCount++
			log.Printf("   ✨ Created: %s (id: %s)", name, deviceID)
		} else {
			reusedCount++
			log.Printf("   ♻️  Reusing: %s (id: %s)", name, deviceID)
		}

		simulators[i] = &DeviceSimulator{
			ID:       deviceID,
			Name:     name,
			Type:     deviceType.Name,
			client:   client,
			stopChan: make(chan struct{}),
		}
	}

	log.Println()
	log.Printf("📊 Summary: %d created, %d reused", createdCount, reusedCount)
	log.Println()

	// Start simulators
	for _, sim := range simulators {
		go sim.Run(time.Duration(*interval) * time.Second)
	}
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
		log.Printf("❌ [%s] Failed to marshal JSON: %v", d.Name, err)
		return
	}

	topic := fmt.Sprintf("devices/%s/telemetry", d.ID)
	token := d.client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ [%s] Failed to publish: %v", d.Name, token.Error())
		return
	}

	// Log first metric for visibility
	log.Printf("📤 [%s] %s=%.2f%s", d.Name, metrics[0].Name, metrics[0].Value, metrics[0].Unit)
}
