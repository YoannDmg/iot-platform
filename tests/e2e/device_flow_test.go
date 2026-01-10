// +build e2e

package e2e

import (
	"net/http"
	"testing"
)

// TestE2E_DeviceLifecycle tests the complete device CRUD lifecycle.
func TestE2E_DeviceLifecycle(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	// First, register and login a user to get token
	var jwtToken string
	var deviceID string

	t.Run("setup_user", func(t *testing.T) {
		// Register
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
					"email":    "device-test@example.com",
					"password": "Password123!",
					"name":     "Device Test User",
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
					"email":    "device-test@example.com",
					"password": "Password123!",
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, loginMutation, "")
		data := resp["data"].(map[string]interface{})
		login := data["login"].(map[string]interface{})
		jwtToken = login["token"].(string)

		t.Logf("✓ User setup complete, token obtained")
	})

	t.Run("create_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation CreateDevice($input: CreateDeviceInput!) {
					createDevice(input: $input) {
						id
						name
						type
						status
						metadata {
							key
							value
						}
						createdAt
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"name": "E2E Test Sensor",
					"type": "temperature",
					"metadata": []map[string]interface{}{
						{"key": "location", "value": "test-lab"},
						{"key": "floor", "value": "3"},
					},
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, jwtToken)

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid response: %+v", resp)
		}

		device, ok := data["createDevice"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid device data: %+v", data)
		}

		// Verify device fields
		if device["name"] != "E2E Test Sensor" {
			t.Errorf("Name = %v, want 'E2E Test Sensor'", device["name"])
		}
		if device["type"] != "temperature" {
			t.Errorf("Type = %v, want 'temperature'", device["type"])
		}
		if device["status"] != "ONLINE" {
			t.Errorf("Status = %v, want 'ONLINE'", device["status"])
		}

		// Extract device ID for later tests
		deviceID, ok = device["id"].(string)
		if !ok || deviceID == "" {
			t.Fatalf("Device ID not generated: %v", device["id"])
		}

		// Verify metadata
		metadata, ok := device["metadata"].([]interface{})
		if !ok {
			t.Errorf("Metadata not present or wrong type")
		} else if len(metadata) < 2 {
			t.Errorf("Expected at least 2 metadata entries, got %d", len(metadata))
		} else {
			// Check first metadata entry
			entry := metadata[0].(map[string]interface{})
			if entry["key"] != "location" || entry["value"] != "test-lab" {
				t.Errorf("Metadata[0] = %v, want {key: location, value: test-lab}", entry)
			}
		}

		t.Logf("✓ Device created: %s", deviceID)
	})

	t.Run("get_device", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query GetDevice($id: ID!) {
					device(id: $id) {
						id
						name
						type
						status
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": deviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		data := resp["data"].(map[string]interface{})
		device := data["device"].(map[string]interface{})

		if device["id"] != deviceID {
			t.Errorf("Device ID = %v, want %v", device["id"], deviceID)
		}
		if device["name"] != "E2E Test Sensor" {
			t.Errorf("Device name = %v, want 'E2E Test Sensor'", device["name"])
		}

		t.Logf("✓ Device retrieved successfully")
	})

	t.Run("list_devices", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query ListDevices($page: Int!, $pageSize: Int!) {
					devices(page: $page, pageSize: $pageSize) {
						devices {
							id
							name
							type
						}
						total
						page
						pageSize
					}
				}
			`,
			"variables": map[string]interface{}{
				"page":     1,
				"pageSize": 10,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		data := resp["data"].(map[string]interface{})
		devicesResp := data["devices"].(map[string]interface{})

		devices := devicesResp["devices"].([]interface{})
		if len(devices) == 0 {
			t.Error("Expected at least 1 device in list")
		}

		// Find our device in the list
		found := false
		for _, d := range devices {
			device := d.(map[string]interface{})
			if device["id"] == deviceID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Created device %s not found in list", deviceID)
		}

		total := devicesResp["total"]
		t.Logf("✓ Listed %v device(s), total: %v", len(devices), total)
	})

	t.Run("update_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation UpdateDevice($input: UpdateDeviceInput!) {
					updateDevice(input: $input) {
						id
						name
						status
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"id":     deviceID,
					"name":   "Updated E2E Sensor",
					"status": "MAINTENANCE",
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, jwtToken)

		data := resp["data"].(map[string]interface{})
		device := data["updateDevice"].(map[string]interface{})

		if device["name"] != "Updated E2E Sensor" {
			t.Errorf("Name not updated: %v", device["name"])
		}
		if device["status"] != "MAINTENANCE" {
			t.Errorf("Status not updated: %v", device["status"])
		}

		t.Logf("✓ Device updated successfully")
	})

	t.Run("delete_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation DeleteDevice($id: ID!) {
					deleteDevice(id: $id) {
						success
						message
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": deviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, jwtToken)

		data := resp["data"].(map[string]interface{})
		deleteResp := data["deleteDevice"].(map[string]interface{})

		if deleteResp["success"] != true {
			t.Errorf("Delete failed: %+v", deleteResp)
		}

		t.Logf("✓ Device deleted successfully")
	})

	t.Run("verify_device_deleted", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query GetDevice($id: ID!) {
					device(id: $id) {
						id
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": deviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		// Should have errors (device not found)
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected error for deleted device, got: %+v", resp)
		}

		t.Logf("✓ Deleted device not accessible")
	})
}

// TestE2E_DeviceWithoutAuth tests that device operations require authentication.
func TestE2E_DeviceWithoutAuth(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	t.Run("create_device_without_token", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation CreateDevice($name: String!, $type: String!) {
					createDevice(name: $name, type: $type) {
						id
					}
				}
			`,
			"variables": map[string]interface{}{
				"name": "Unauthorized Device",
				"type": "sensor",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, "") // No token

		// Should have authentication error
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected authentication error, got: %+v", resp)
		}

		t.Logf("✓ Unauthenticated device creation blocked")
	})

	t.Run("list_devices_without_token", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					devices(page: 1, pageSize: 10) {
						devices { id }
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, "") // No token

		// Should have authentication error
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected authentication error, got: %+v", resp)
		}

		t.Logf("✓ Unauthenticated device listing blocked")
	})
}
