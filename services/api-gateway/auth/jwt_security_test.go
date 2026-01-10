// +build unit

package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWT_TokenTampering tests that modified tokens are rejected.
func TestJWT_TokenTampering(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	// Generate valid token
	token, err := manager.GenerateToken("user-123", "test@example.com", "Test User", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		tamper      func(string) string
		description string
	}{
		{
			name: "modify_payload",
			tamper: func(token string) string {
				// Split token into parts
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					return token
				}
				// Modify middle part (payload) by replacing one character
				payload := parts[1]
				if len(payload) > 0 {
					payload = "X" + payload[1:]
				}
				return parts[0] + "." + payload + "." + parts[2]
			},
			description: "modified payload should be rejected",
		},
		{
			name: "modify_signature",
			tamper: func(token string) string {
				// Split token and modify signature
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					return token
				}
				signature := parts[2]
				if len(signature) > 0 {
					signature = "X" + signature[1:]
				}
				return parts[0] + "." + parts[1] + "." + signature
			},
			description: "modified signature should be rejected",
		},
		{
			name: "remove_signature",
			tamper: func(token string) string {
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					return token
				}
				return parts[0] + "." + parts[1] + "."
			},
			description: "token without signature should be rejected",
		},
		{
			name: "extra_parts",
			tamper: func(token string) string {
				return token + ".extra"
			},
			description: "token with extra parts should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tamperedToken := tt.tamper(token)

			_, err := manager.ValidateToken(tamperedToken)
			if err == nil {
				t.Errorf("%s: expected error for tampered token", tt.description)
			}
		})
	}
}

// TestJWT_ClaimInjection tests that claim injection attacks are prevented.
func TestJWT_ClaimInjection(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	// Attempt to create token with injected claims
	maliciousClaims := &Claims{
		UserID: "user-123",
		Email:  "user@example.com",
		Name:   "Regular User",
		Role:   "user\",\"role\":\"admin", // Injection attempt
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, maliciousClaims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Verify role hasn't been escalated
	if strings.Contains(claims.Role, "admin") && !strings.Contains(claims.Role, "user") {
		t.Error("Role injection succeeded - security vulnerability!")
	}

	// The injected string should be preserved as-is (not parsed)
	expectedRole := "user\",\"role\":\"admin"
	if claims.Role != expectedRole {
		t.Errorf("Role = %v, expected %v", claims.Role, expectedRole)
	}
}

// TestJWT_AlgorithmConfusion tests protection against algorithm confusion attacks.
func TestJWT_AlgorithmConfusion(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	// Try to create token with "none" algorithm
	claims := &Claims{
		UserID: "user-123",
		Email:  "attacker@example.com",
		Name:   "Attacker",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with "none" algorithm (unsigned)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to create unsigned token: %v", err)
	}

	// Validation should fail
	_, err = manager.ValidateToken(tokenString)
	if err == nil {
		t.Error("Unsigned token (alg:none) should be rejected - security vulnerability!")
	}
}

// TestJWT_ReplayAttack tests that old tokens can't be reused indefinitely.
func TestJWT_ReplayAttack(t *testing.T) {
	// Create manager with short lifetime
	manager := NewJWTManager("test-secret", 1*time.Second)

	// Generate token
	token, err := manager.GenerateToken("user-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = manager.ValidateToken(token)
	if err != nil {
		t.Errorf("Fresh token should be valid: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Replayed token should be rejected
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expired token should be rejected to prevent replay attacks")
	}

	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected expiration error, got: %v", err)
	}
}

// TestJWT_WeakSecret tests that weak secrets are detectable.
func TestJWT_WeakSecret(t *testing.T) {
	weakSecrets := []string{
		"",
		"a",
		"12",
		"abc",
		"password",
		"secret",
		"test",
	}

	for _, secret := range weakSecrets {
		t.Run("secret_"+secret, func(t *testing.T) {
			manager := NewJWTManager(secret, 1*time.Hour)

			// Should still work but we can warn in production
			token, err := manager.GenerateToken("user-123", "test@example.com", "User", "user")
			if err != nil {
				t.Errorf("Token generation failed: %v", err)
			}

			// Even weak secrets should validate if correct
			_, err = manager.ValidateToken(token)
			if err != nil {
				t.Errorf("ValidateToken failed with weak secret: %v", err)
			}

			// Verify minimum length
			if len(secret) < 32 {
				t.Logf("WARNING: Secret '%s' is too weak (length=%d, recommended>=32)", secret, len(secret))
			}
		})
	}
}

// TestJWT_ExpirationBoundary tests edge cases around expiration time.
func TestJWT_ExpirationBoundary(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		sleep    time.Duration
		wantErr  bool
	}{
		{
			name:     "token_just_before_expiry",
			duration: 2 * time.Second,
			sleep:    1 * time.Second,
			wantErr:  false,
		},
		{
			name:     "token_just_after_expiry",
			duration: 1 * time.Second,
			sleep:    2 * time.Second,
			wantErr:  true,
		},
		{
			name:     "token_at_expiry_boundary",
			duration: 500 * time.Millisecond,
			sleep:    500 * time.Millisecond,
			wantErr:  true, // Should be expired or very close
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager("test-secret", tt.duration)

			token, err := manager.GenerateToken("user-123", "test@example.com", "User", "user")
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			time.Sleep(tt.sleep)

			_, err = manager.ValidateToken(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestJWT_MultipleTokens tests that different tokens don't interfere.
func TestJWT_MultipleTokens(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	// Generate multiple tokens for different users
	token1, err := manager.GenerateToken("user-1", "user1@example.com", "User One", "user")
	if err != nil {
		t.Fatalf("Failed to generate token1: %v", err)
	}

	token2, err := manager.GenerateToken("user-2", "user2@example.com", "User Two", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token2: %v", err)
	}

	token3, err := manager.GenerateToken("user-3", "user3@example.com", "User Three", "device")
	if err != nil {
		t.Fatalf("Failed to generate token3: %v", err)
	}

	// All tokens should be distinct
	if token1 == token2 || token1 == token3 || token2 == token3 {
		t.Error("Generated tokens should be unique")
	}

	// Each token should validate to its correct user
	claims1, err := manager.ValidateToken(token1)
	if err != nil {
		t.Fatalf("Failed to validate token1: %v", err)
	}
	if claims1.UserID != "user-1" || claims1.Role != "user" {
		t.Error("Token1 has wrong claims")
	}

	claims2, err := manager.ValidateToken(token2)
	if err != nil {
		t.Fatalf("Failed to validate token2: %v", err)
	}
	if claims2.UserID != "user-2" || claims2.Role != "admin" {
		t.Error("Token2 has wrong claims")
	}

	claims3, err := manager.ValidateToken(token3)
	if err != nil {
		t.Fatalf("Failed to validate token3: %v", err)
	}
	if claims3.UserID != "user-3" || claims3.Role != "device" {
		t.Error("Token3 has wrong claims")
	}
}

// TestJWT_SpecialCharacters tests handling of special characters in claims.
func TestJWT_SpecialCharacters(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	specialCases := []struct {
		name     string
		email    string
		userName string
		role     string
	}{
		{
			name:     "unicode_characters",
			email:    "user@example.com",
			userName: "用户 Über Ñoño",
			role:     "user",
		},
		{
			name:     "special_symbols",
			email:    "user+test@example.com",
			userName: "O'Brien-Smith",
			role:     "user",
		},
		{
			name:     "quotes_and_escapes",
			email:    "user@example.com",
			userName: "User \"Test\" \\Name\\",
			role:     "user",
		},
		{
			name:     "whitespace",
			email:    "user@example.com",
			userName: "User   With   Spaces",
			role:     "user",
		},
	}

	for _, tt := range specialCases {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken("user-123", tt.email, tt.userName, tt.role)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			// Verify claims are preserved exactly
			if claims.Email != tt.email {
				t.Errorf("Email = %q, want %q", claims.Email, tt.email)
			}
			if claims.Name != tt.userName {
				t.Errorf("Name = %q, want %q", claims.Name, tt.userName)
			}
			if claims.Role != tt.role {
				t.Errorf("Role = %q, want %q", claims.Role, tt.role)
			}
		})
	}
}

// TestJWT_EmptyAndNullClaims tests handling of empty/null claim values.
func TestJWT_EmptyAndNullClaims(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	tests := []struct {
		name            string
		userID          string
		email           string
		userName        string
		role            string
		wantValidateErr bool
	}{
		{
			name:            "all_populated",
			userID:          "user-123",
			email:           "user@example.com",
			userName:        "Test User",
			role:            "user",
			wantValidateErr: false,
		},
		{
			name:            "empty_name",
			userID:          "user-123",
			email:           "user@example.com",
			userName:        "",
			role:            "user",
			wantValidateErr: false, // Empty name should be allowed
		},
		{
			name:            "empty_userID",
			userID:          "",
			email:           "user@example.com",
			userName:        "Test User",
			role:            "user",
			wantValidateErr: true, // Empty userID should fail validation
		},
		{
			name:            "empty_email",
			userID:          "user-123",
			email:           "",
			userName:        "Test User",
			role:            "user",
			wantValidateErr: true, // Empty email should fail validation
		},
		{
			name:            "empty_role",
			userID:          "user-123",
			email:           "user@example.com",
			userName:        "Test User",
			role:            "",
			wantValidateErr: false, // Empty role should be allowed (default handling)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generation should always succeed (no validation)
			token, err := manager.GenerateToken(tt.userID, tt.email, tt.userName, tt.role)
			if err != nil {
				t.Fatalf("GenerateToken() unexpected error = %v", err)
			}

			// Validation may fail for required fields
			claims, err := manager.ValidateToken(token)

			if tt.wantValidateErr {
				if err == nil {
					t.Error("ValidateToken() should fail for missing required fields")
				}
				return
			}

			if err != nil {
				t.Fatalf("ValidateToken() unexpected error = %v", err)
			}

			// Verify values are preserved exactly
			if claims.UserID != tt.userID {
				t.Errorf("UserID = %q, want %q", claims.UserID, tt.userID)
			}
			if claims.Email != tt.email {
				t.Errorf("Email = %q, want %q", claims.Email, tt.email)
			}
			if claims.Name != tt.userName {
				t.Errorf("Name = %q, want %q", claims.Name, tt.userName)
			}
			if claims.Role != tt.role {
				t.Errorf("Role = %q, want %q", claims.Role, tt.role)
			}
		})
	}
}

// TestJWT_ConcurrentAccess tests thread safety of JWT operations.
func TestJWT_ConcurrentAccess(t *testing.T) {
	manager := NewJWTManager("test-secret", 1*time.Hour)

	const numGoroutines = 100
	tokens := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Generate tokens concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			token, err := manager.GenerateToken(
				"user-"+string(rune('0'+id)),
				"user"+string(rune('0'+id))+"@example.com",
				"User "+string(rune('0'+id)),
				"user",
			)
			if err != nil {
				errors <- err
				return
			}
			tokens <- token
		}(i)
	}

	// Collect tokens
	generatedTokens := make([]string, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		select {
		case token := <-tokens:
			generatedTokens = append(generatedTokens, token)
		case err := <-errors:
			t.Errorf("Concurrent generation error: %v", err)
		}
	}

	// Validate all tokens concurrently
	for _, token := range generatedTokens {
		go func(tok string) {
			_, err := manager.ValidateToken(tok)
			if err != nil {
				errors <- err
			} else {
				errors <- nil
			}
		}(token)
	}

	// Check validation results
	for i := 0; i < len(generatedTokens); i++ {
		err := <-errors
		if err != nil {
			t.Errorf("Concurrent validation error: %v", err)
		}
	}
}

// TestJWT_LongLivedTokens tests tokens with very long expiration.
func TestJWT_LongLivedTokens(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "30_days",
			duration: 30 * 24 * time.Hour,
		},
		{
			name:     "1_year",
			duration: 365 * 24 * time.Hour,
		},
		{
			name:     "10_years",
			duration: 10 * 365 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager("test-secret", tt.duration)

			token, err := manager.GenerateToken("device-123", "device@example.com", "IoT Device", "device")
			if err != nil {
				t.Fatalf("Failed to generate long-lived token: %v", err)
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate long-lived token: %v", err)
			}

			// Verify expiration is far in the future
			expiresIn := time.Until(claims.ExpiresAt.Time)
			expectedMin := tt.duration - 1*time.Minute // Allow some tolerance
			if expiresIn < expectedMin {
				t.Errorf("Token expires too soon: %v, expected ~%v", expiresIn, tt.duration)
			}
		})
	}
}
