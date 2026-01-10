// +build e2e

package e2e

import (
	"net/http"
	"testing"
)

// TestE2E_RBACPermissions tests role-based access control between users and admins.
func TestE2E_RBACPermissions(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	var userToken string
	var adminToken string
	var userDeviceID string

	// Setup: Create regular user
	t.Run("setup_regular_user", func(t *testing.T) {
		registerMutation := map[string]interface{}{
			"query": `
				mutation Register($email: String!, $password: String!, $name: String!) {
					register(email: $email, password: $password, name: $name) {
						user {
							id
							role
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "regular-user@example.com",
				"password": "UserPassword123!",
				"name":     "Regular User",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, registerMutation, "")
		data := resp["data"].(map[string]interface{})
		register := data["register"].(map[string]interface{})
		user := register["user"].(map[string]interface{})

		if user["role"] != "user" {
			t.Fatalf("Expected role 'user', got: %v", user["role"])
		}

		// Login
		loginMutation := map[string]interface{}{
			"query": `
				mutation Login($email: String!, $password: String!) {
					login(email: $email, password: $password) {
						token
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "regular-user@example.com",
				"password": "UserPassword123!",
			},
		}

		loginResp := graphqlRequest(t, client, gatewayURL, loginMutation, "")
		loginData := loginResp["data"].(map[string]interface{})
		login := loginData["login"].(map[string]interface{})
		userToken = login["token"].(string)

		t.Logf("✓ Regular user created and logged in")
	})

	// Setup: Create admin user
	t.Run("setup_admin_user", func(t *testing.T) {
		registerMutation := map[string]interface{}{
			"query": `
				mutation Register($email: String!, $password: String!, $name: String!, $role: String) {
					register(email: $email, password: $password, name: $name, role: $role) {
						user {
							id
							role
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "admin-user@example.com",
				"password": "AdminPassword123!",
				"name":     "Admin User",
				"role":     "admin",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, registerMutation, "")
		data := resp["data"].(map[string]interface{})
		register := data["register"].(map[string]interface{})
		user := register["user"].(map[string]interface{})

		if user["role"] != "admin" {
			t.Fatalf("Expected role 'admin', got: %v", user["role"])
		}

		// Login
		loginMutation := map[string]interface{}{
			"query": `
				mutation Login($email: String!, $password: String!) {
					login(email: $email, password: $password) {
						token
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "admin-user@example.com",
				"password": "AdminPassword123!",
			},
		}

		loginResp := graphqlRequest(t, client, gatewayURL, loginMutation, "")
		loginData := loginResp["data"].(map[string]interface{})
		login := loginData["login"].(map[string]interface{})
		adminToken = login["token"].(string)

		t.Logf("✓ Admin user created and logged in")
	})

	// User creates a device
	t.Run("user_creates_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation CreateDevice($name: String!, $type: String!) {
					createDevice(name: $name, type: $type) {
						id
						name
					}
				}
			`,
			"variables": map[string]interface{}{
				"name": "User's Device",
				"type": "sensor",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, userToken)

		data := resp["data"].(map[string]interface{})
		device := data["createDevice"].(map[string]interface{})
		userDeviceID = device["id"].(string)

		t.Logf("✓ Regular user created device: %s", userDeviceID)
	})

	// User can see their own devices
	t.Run("user_lists_own_devices", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					devices(page: 1, pageSize: 10) {
						devices {
							id
							name
						}
						total
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, userToken)

		data := resp["data"].(map[string]interface{})
		devicesResp := data["devices"].(map[string]interface{})
		devices := devicesResp["devices"].([]interface{})

		found := false
		for _, d := range devices {
			device := d.(map[string]interface{})
			if device["id"] == userDeviceID {
				found = true
				break
			}
		}

		if !found {
			t.Error("User should see their own device")
		}

		t.Logf("✓ User can list devices (found %d)", len(devices))
	})

	// User can update their own device
	t.Run("user_updates_own_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation UpdateDevice($id: ID!, $name: String!) {
					updateDevice(id: $id, name: $name) {
						id
						name
					}
				}
			`,
			"variables": map[string]interface{}{
				"id":   userDeviceID,
				"name": "User's Updated Device",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, userToken)

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected successful update, got: %+v", resp)
		}

		device := data["updateDevice"].(map[string]interface{})
		if device["name"] != "User's Updated Device" {
			t.Error("Device name not updated")
		}

		t.Logf("✓ User can update their own device")
	})

	// Admin can see all devices
	t.Run("admin_lists_all_devices", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					devices(page: 1, pageSize: 10) {
						devices {
							id
							name
						}
						total
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, adminToken)

		data := resp["data"].(map[string]interface{})
		devicesResp := data["devices"].(map[string]interface{})
		devices := devicesResp["devices"].([]interface{})

		found := false
		for _, d := range devices {
			device := d.(map[string]interface{})
			if device["id"] == userDeviceID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Admin should see all devices including user's device")
		}

		t.Logf("✓ Admin can see all devices (found %d)", len(devices))
	})

	// Admin can update any device
	t.Run("admin_updates_users_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation UpdateDevice($id: ID!, $name: String!) {
					updateDevice(id: $id, name: $name) {
						id
						name
					}
				}
			`,
			"variables": map[string]interface{}{
				"id":   userDeviceID,
				"name": "Admin Updated This Device",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, adminToken)

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected admin to update user's device, got: %+v", resp)
		}

		device := data["updateDevice"].(map[string]interface{})
		if device["name"] != "Admin Updated This Device" {
			t.Error("Device name not updated by admin")
		}

		t.Logf("✓ Admin can update user's device")
	})

	// Admin can delete any device
	t.Run("admin_deletes_users_device", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation DeleteDevice($id: ID!) {
					deleteDevice(id: $id) {
						success
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": userDeviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, adminToken)

		data := resp["data"].(map[string]interface{})
		deleteResp := data["deleteDevice"].(map[string]interface{})

		if deleteResp["success"] != true {
			t.Errorf("Admin should be able to delete user's device")
		}

		t.Logf("✓ Admin can delete user's device")
	})

	// Verify device is deleted
	t.Run("verify_device_deleted_by_admin", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query GetDevice($id: ID!) {
					device(id: $id) {
						id
					}
				}
			`,
			"variables": map[string]interface{}{
				"id": userDeviceID,
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, query, userToken)

		// Should have errors (device not found)
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected device to be deleted, got: %+v", resp)
		}

		t.Logf("✓ Device successfully deleted by admin")
	})
}

// TestE2E_AdminUserManagement tests that admins can manage users.
func TestE2E_AdminUserManagement(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	var adminToken string

	// Create admin
	t.Run("setup_admin", func(t *testing.T) {
		registerMutation := map[string]interface{}{
			"query": `
				mutation Register($email: String!, $password: String!, $name: String!, $role: String) {
					register(email: $email, password: $password, name: $name, role: $role) {
						user { id }
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "user-admin@example.com",
				"password": "AdminPass123!",
				"name":     "User Admin",
				"role":     "admin",
			},
		}
		graphqlRequest(t, client, gatewayURL, registerMutation, "")

		loginMutation := map[string]interface{}{
			"query": `
				mutation Login($email: String!, $password: String!) {
					login(email: $email, password: $password) {
						token
					}
				}
			`,
			"variables": map[string]interface{}{
				"email":    "user-admin@example.com",
				"password": "AdminPass123!",
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, loginMutation, "")
		data := resp["data"].(map[string]interface{})
		login := data["login"].(map[string]interface{})
		adminToken = login["token"].(string)
	})

	t.Run("admin_lists_users", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					users(page: 1, pageSize: 10) {
						users {
							id
							email
							role
						}
						total
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, adminToken)

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected users list, got: %+v", resp)
		}

		usersResp := data["users"].(map[string]interface{})
		users := usersResp["users"].([]interface{})

		if len(users) == 0 {
			t.Error("Expected at least 1 user")
		}

		t.Logf("✓ Admin listed %d users", len(users))
	})
}
