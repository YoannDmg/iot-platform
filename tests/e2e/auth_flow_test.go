// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestE2E_UserRegistrationAndLogin tests the complete user registration and authentication flow.
func TestE2E_UserRegistrationAndLogin(t *testing.T) {
	env := SetupE2EEnvironment(t)

	client := &http.Client{}
	gatewayURL := "http://" + env.APIGatewayAddr + "/query"

	// Test data
	testEmail := "e2e-test@example.com"
	testPassword := "SecurePassword123!"
	testName := "E2E Test User"

	t.Run("register_new_user", func(t *testing.T) {
		// GraphQL mutation for registration
		mutation := map[string]interface{}{
			"query": `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						token
						user {
							id
							email
							name
							role
							isActive
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail,
					"password": testPassword,
					"name":     testName,
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, "")

		// Verify response structure
		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid response structure: %+v", resp)
		}

		register, ok := data["register"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid register response: %+v", data)
		}

		// Verify user data
		user, ok := register["user"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid user data: %+v", register)
		}

		if user["email"] != testEmail {
			t.Errorf("Email = %v, want %v", user["email"], testEmail)
		}
		if user["name"] != testName {
			t.Errorf("Name = %v, want %v", user["name"], testName)
		}
		if user["role"] != "user" {
			t.Errorf("Role = %v, want 'user'", user["role"])
		}
		if user["isActive"] != true {
			t.Errorf("IsActive = %v, want true", user["isActive"])
		}

		// Verify user ID is generated
		userID, ok := user["id"].(string)
		if !ok || userID == "" {
			t.Errorf("User ID should be generated, got: %v", user["id"])
		}

		t.Logf("✓ User registered: %s (%s)", testEmail, userID)
	})

	var jwtToken string

	t.Run("login_with_valid_credentials", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation Login($input: LoginInput!) {
					login(input: $input) {
						token
						user {
							id
							email
							name
							role
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail,
					"password": testPassword,
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, "")

		data := resp["data"].(map[string]interface{})
		login := data["login"].(map[string]interface{})

		// Extract JWT token
		token, ok := login["token"].(string)
		if !ok || token == "" {
			t.Fatalf("JWT token not returned: %+v", login)
		}

		jwtToken = token

		// Verify user data
		user := login["user"].(map[string]interface{})
		if user["email"] != testEmail {
			t.Errorf("Email = %v, want %v", user["email"], testEmail)
		}

		t.Logf("✓ Login successful, token received: %s...", token[:20])
	})

	t.Run("login_with_invalid_password", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation Login($input: LoginInput!) {
					login(input: $input) {
						token
						user {
							id
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail,
					"password": "WrongPassword123!",
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, "")

		// Should have errors
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected error for invalid password, got: %+v", resp)
		}

		t.Logf("✓ Invalid password rejected as expected")
	})

	t.Run("access_protected_resource_with_token", func(t *testing.T) {
		// Query that requires authentication
		query := map[string]interface{}{
			"query": `
				query {
					me {
						id
						email
						name
						role
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, jwtToken)

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid response: %+v", resp)
		}

		me, ok := data["me"].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid me data: %+v", data)
		}

		if me["email"] != testEmail {
			t.Errorf("Email = %v, want %v", me["email"], testEmail)
		}

		t.Logf("✓ Authenticated request successful")
	})

	t.Run("access_protected_resource_without_token", func(t *testing.T) {
		query := map[string]interface{}{
			"query": `
				query {
					me {
						id
					}
				}
			`,
		}

		resp := graphqlRequest(t, client, gatewayURL, query, "")

		// Should have errors
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected authentication error, got: %+v", resp)
		}

		t.Logf("✓ Unauthenticated request blocked as expected")
	})

	t.Run("duplicate_registration_fails", func(t *testing.T) {
		mutation := map[string]interface{}{
			"query": `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						token
						user {
							id
						}
					}
				}
			`,
			"variables": map[string]interface{}{
				"input": map[string]interface{}{
					"email":    testEmail, // Same email
					"password": "AnotherPassword123!",
					"name":     "Another User",
				},
			},
		}

		resp := graphqlRequest(t, client, gatewayURL, mutation, "")

		// Should have errors for duplicate email
		errors, ok := resp["errors"].([]interface{})
		if !ok || len(errors) == 0 {
			t.Fatalf("Expected error for duplicate email, got: %+v", resp)
		}

		t.Logf("✓ Duplicate registration blocked as expected")
	})
}

// TestE2E_TokenExpiration tests JWT token expiration handling.
func TestE2E_TokenExpiration(t *testing.T) {
	// This would require restarting services with very short token duration
	// Skipping for now as it would make tests slower
	t.Skip("Token expiration test requires custom configuration")
}

// graphqlRequest sends a GraphQL request and returns the parsed response.
func graphqlRequest(t *testing.T, client *http.Client, url string, query map[string]interface{}, token string) map[string]interface{} {
	t.Helper()

	// Marshal query
	body, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("Failed to marshal query: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add auth token if provided
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("Failed to parse response: %v\nBody: %s", err, respBody)
	}

	return result
}
