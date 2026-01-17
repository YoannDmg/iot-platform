// +build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

// TestE2E_TelemetryPipeline tests the complete telemetry flow:
// MQTT publish -> Telemetry Collector -> TimescaleDB -> GraphQL query
func TestE2E_TelemetryPipeline(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	// Setup: Create user and device
	var jwtToken string
	var deviceID string

	// Unique email per test run
	uniqueID := time.Now().UnixNano()
	testEmail := fmt.Sprintf("telemetry-test-%d@example.com", uniqueID)

	t.Run("setup_user_and_device", func(t *testing.T) {
		// Register user
		registerMutation := map[string]interface{}{
			"query": `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						token
						user { id }
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail,
					"password": "Password123!",
					"name":     "Telemetry Test User",
				},
			},
		}
		graphqlRequest(t, client, gatewayURL, registerMutation, "")

		// Login
		loginMutation := map[string]interface{}{
			"query": `
				mutation Login($input: LoginInput!) {
					login(input: $input) {
						token
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail,
					"password": "Password123!",
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, loginMutation, "")
		data := resp["data"].(map[string]interface{})
		login := data["login"].(map[string]interface{})
		jwtToken = login["token"].(string)

		// Create device
		createDeviceMutation := map[string]interface{}{
			"query": `
				mutation CreateDevice($input: CreateDeviceInput!) {
					createDevice(input: $input) {
						id
						name
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"name": "E2E Telemetry Sensor",
					"type": "temperature_sensor",
					"metadata": []map[string]interface{}{
						{"key": "location", "value": "test-lab"},
					},
				},
			},
		}

		resp = graphqlRequest(t, client, gatewayURL, createDeviceMutation, jwtToken)
		data = resp["data"].(map[string]interface{})
		device := data["createDevice"].(map[string]interface{})
		deviceID = device["id"].(string)

		t.Logf("Setup complete: device=%s", deviceID)
	})

	// Connect to MQTT and publish telemetry
	t.Run("publish_telemetry_via_mqtt", func(t *testing.T) {
		mqttBroker := "tcp://" + env.MQTTBrokerAddr

		opts := mqtt.NewClientOptions()
		opts.AddBroker(mqttBroker)
		opts.SetClientID("e2e-test-publisher")
		opts.SetConnectTimeout(10 * time.Second)

		mqttClient := mqtt.NewClient(opts)
		token := mqttClient.Connect()
		if !token.WaitTimeout(10 * time.Second) {
			t.Fatalf("MQTT connection timeout")
		}
		if token.Error() != nil {
			t.Fatalf("Failed to connect to MQTT: %v", token.Error())
		}
		defer mqttClient.Disconnect(1000)

		t.Log("Connected to MQTT broker")

		// Publish multiple telemetry points
		for i := 0; i < 5; i++ {
			msg := TelemetryMessage{
				DeviceID:  deviceID,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Metrics: []Metric{
					{Name: "temperature", Value: 22.5 + float64(i)*0.5, Unit: "C"},
					{Name: "humidity", Value: 45.0 + float64(i)*2, Unit: "%"},
				},
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("Failed to marshal telemetry: %v", err)
			}

			topic := fmt.Sprintf("devices/%s/telemetry", deviceID)
			pubToken := mqttClient.Publish(topic, 1, false, payload)
			if !pubToken.WaitTimeout(5 * time.Second) {
				t.Fatalf("MQTT publish timeout")
			}
			if pubToken.Error() != nil {
				t.Fatalf("Failed to publish: %v", pubToken.Error())
			}

			t.Logf("Published telemetry point %d: temperature=%.1f", i+1, msg.Metrics[0].Value)
			time.Sleep(100 * time.Millisecond) // Small delay between messages
		}

		t.Log("All telemetry points published")
	})

	// Wait for telemetry to be processed
	t.Run("wait_for_processing", func(t *testing.T) {
		// Give the telemetry collector time to process
		time.Sleep(2 * time.Second)
		t.Log("Waited for telemetry processing")
	})

	// Query telemetry via GraphQL
	t.Run("query_device_metrics", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query GetDeviceMetrics($deviceId: ID!) {
					deviceMetrics(deviceId: $deviceId)
				}
			`,
			"variables": map[string]interface{}{
				"deviceId": deviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		// Check for errors
		if errors, hasErrors := resp["errors"].([]interface{}); hasErrors && len(errors) > 0 {
			t.Fatalf("GraphQL error: %+v", errors)
		}

		data := resp["data"].(map[string]interface{})
		metrics := data["deviceMetrics"].([]interface{})

		if len(metrics) == 0 {
			t.Error("Expected at least one metric, got none")
		}

		t.Logf("Available metrics: %v", metrics)

		// Should have temperature and humidity
		hasTemp := false
		hasHumidity := false
		for _, m := range metrics {
			if m == "temperature" {
				hasTemp = true
			}
			if m == "humidity" {
				hasHumidity = true
			}
		}

		if !hasTemp {
			t.Error("Expected 'temperature' metric")
		}
		if !hasHumidity {
			t.Error("Expected 'humidity' metric")
		}
	})

	t.Run("query_latest_metric", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query GetLatestMetric($deviceId: ID!, $metricName: String!) {
					deviceLatestMetric(deviceId: $deviceId, metricName: $metricName) {
						time
						value
						unit
					}
				}
			`,
			"variables": map[string]interface{}{
				"deviceId":   deviceID,
				"metricName": "temperature",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		if errors, hasErrors := resp["errors"].([]interface{}); hasErrors && len(errors) > 0 {
			t.Fatalf("GraphQL error: %+v", errors)
		}

		data := resp["data"].(map[string]interface{})
		latest := data["deviceLatestMetric"].(map[string]interface{})

		// Last published value should be 22.5 + 4*0.5 = 24.5
		value := latest["value"].(float64)
		if value < 22.0 || value > 25.0 {
			t.Errorf("Expected temperature between 22-25, got %.2f", value)
		}

		t.Logf("Latest temperature: %.2f", value)
	})

	t.Run("query_telemetry_series", func(t *testing.T) {
		// Query last hour of data
		now := time.Now().Unix()
		from := now - 3600 // 1 hour ago

		query := map[string]interface{}{
			"query": `
				query GetTelemetry($deviceId: ID!, $metricName: String!, $from: Int!, $to: Int!, $limit: Int) {
					deviceTelemetry(deviceId: $deviceId, metricName: $metricName, from: $from, to: $to, limit: $limit) {
						metricName
						points {
							time
							value
							unit
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"deviceId":   deviceID,
				"metricName": "temperature",
				"from":       from,
				"to":         now,
				"limit":      100,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		if errors, hasErrors := resp["errors"].([]interface{}); hasErrors && len(errors) > 0 {
			t.Fatalf("GraphQL error: %+v", errors)
		}

		data := resp["data"].(map[string]interface{})
		series := data["deviceTelemetry"].(map[string]interface{})
		points := series["points"].([]interface{})

		if len(points) < 5 {
			t.Errorf("Expected at least 5 telemetry points, got %d", len(points))
		}

		t.Logf("Retrieved %d telemetry points", len(points))

		// Verify first few points
		for i, p := range points {
			if i >= 3 {
				break
			}
			point := p.(map[string]interface{})
			t.Logf("  Point %d: time=%v, value=%.2f", i, point["time"], point["value"].(float64))
		}
	})

	t.Run("query_aggregated_telemetry", func(t *testing.T) {
		now := time.Now().Unix()
		from := now - 3600

		query := map[string]interface{}{
			"query": `
				query GetAggregatedTelemetry($deviceId: ID!, $metricName: String!, $from: Int!, $to: Int!, $interval: String!) {
					deviceTelemetryAggregated(deviceId: $deviceId, metricName: $metricName, from: $from, to: $to, interval: $interval) {
						bucket
						avg
						min
						max
						count
					}
				}
			`,
			"variables": map[string]interface{}{
				"deviceId":   deviceID,
				"metricName": "temperature",
				"from":       from,
				"to":         now,
				"interval":   "1 minute",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		if errors, hasErrors := resp["errors"].([]interface{}); hasErrors && len(errors) > 0 {
			t.Fatalf("GraphQL error: %+v", errors)
		}

		data := resp["data"].(map[string]interface{})
		aggregations := data["deviceTelemetryAggregated"].([]interface{})

		if len(aggregations) == 0 {
			t.Error("Expected at least one aggregation bucket")
		}

		t.Logf("Retrieved %d aggregation buckets", len(aggregations))

		// Check first bucket
		if len(aggregations) > 0 {
			bucket := aggregations[0].(map[string]interface{})
			t.Logf("First bucket: avg=%.2f, min=%.2f, max=%.2f, count=%v",
				bucket["avg"].(float64),
				bucket["min"].(float64),
				bucket["max"].(float64),
				bucket["count"])
		}
	})

	// Cleanup: delete device
	t.Run("cleanup_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation DeleteDevice($id: ID!) {
					deleteDevice(id: $id) {
						success
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": deviceID,
			},
		}

		graphqlRequest(t, client, gatewayURL, mutation, jwtToken)
		t.Log("Device cleaned up")
	})
}

// TestE2E_TelemetryWithoutAuth tests that telemetry queries require authentication
func TestE2E_TelemetryWithoutAuth(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	t.Run("query_telemetry_without_token", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					deviceMetrics(deviceId: "some-device-id")
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, "") // No token

		// Should have authentication error
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected authentication error, got: %+v", resp)
		}

		t.Log("Unauthenticated telemetry query blocked")
	})
}

// TestE2E_TelemetryForNonExistentDevice tests querying telemetry for a device that doesn't exist
func TestE2E_TelemetryForNonExistentDevice(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	// Setup: Get auth token
	var jwtToken string

	t.Run("setup", func(t *testing.T) {
		registerMutation := map[string]interface{}{
			"query": `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						token
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    "telemetry-nodevice@example.com",
					"password": "Password123!",
					"name":     "Test User",
				},
			},
		}
		resp := graphqlRequest(t, client, gatewayURL, registerMutation, "")
		data := resp["data"].(map[string]interface{})
		register := data["register"].(map[string]interface{})
		jwtToken = register["token"].(string)
	})

	t.Run("query_metrics_for_nonexistent_device", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					deviceMetrics(deviceId: "nonexistent-device-id-12345")
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		// Should return empty array or error
		data, hasData := resp["data"].(map[string]interface{})
		if hasData {
			metrics := data["deviceMetrics"].([]interface{})
			if len(metrics) != 0 {
				t.Errorf("Expected empty metrics for nonexistent device, got %d", len(metrics))
			}
			t.Log("Correctly returned empty metrics for nonexistent device")
		} else {
			// GraphQL error is also acceptable
			t.Log("Returned error for nonexistent device (acceptable)")
		}
	})
}
