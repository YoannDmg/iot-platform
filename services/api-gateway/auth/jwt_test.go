package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		duration time.Duration
		userID   string
		email    string
		userName string
		role     string
		wantErr  bool
	}{
		{
			name:     "valid token generation",
			secret:   "test-secret-key",
			duration: 1 * time.Hour,
			userID:   "user-123",
			email:    "test@example.com",
			userName: "Test User",
			role:     "user",
			wantErr:  false,
		},
		{
			name:     "admin role token",
			secret:   "test-secret-key",
			duration: 24 * time.Hour,
			userID:   "admin-456",
			email:    "admin@example.com",
			userName: "Admin User",
			role:     "admin",
			wantErr:  false,
		},
		{
			name:     "device role token",
			secret:   "test-secret-key",
			duration: 30 * 24 * time.Hour, // 30 days for devices
			userID:   "device-789",
			email:    "device@example.com",
			userName: "IoT Device",
			role:     "device",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager(tt.secret, tt.duration)

			token, err := manager.GenerateToken(tt.userID, tt.email, tt.userName, tt.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			// Verify token can be parsed
			if !tt.wantErr {
				parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(tt.secret), nil
				})

				if err != nil {
					t.Errorf("Failed to parse generated token: %v", err)
					return
				}

				if !parsedToken.Valid {
					t.Error("Generated token is not valid")
					return
				}

				claims, ok := parsedToken.Claims.(*Claims)
				if !ok {
					t.Error("Failed to extract claims from token")
					return
				}

				// Verify claims
				if claims.UserID != tt.userID {
					t.Errorf("UserID = %v, want %v", claims.UserID, tt.userID)
				}
				if claims.Email != tt.email {
					t.Errorf("Email = %v, want %v", claims.Email, tt.email)
				}
				if claims.Name != tt.userName {
					t.Errorf("Name = %v, want %v", claims.Name, tt.userName)
				}
				if claims.Role != tt.role {
					t.Errorf("Role = %v, want %v", claims.Role, tt.role)
				}

				// Verify expiration is set correctly
				expectedExpiry := time.Now().Add(tt.duration)
				actualExpiry := claims.ExpiresAt.Time
				diff := actualExpiry.Sub(expectedExpiry).Abs()
				if diff > 5*time.Second {
					t.Errorf("Expiration time diff = %v, expected ~%v", actualExpiry, expectedExpiry)
				}
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 1*time.Hour)

	tests := []struct {
		name       string
		setupToken func() string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := manager.GenerateToken("user-123", "test@example.com", "Test User", "user")
				return token
			},
			wantErr: false,
		},
		{
			name: "expired token",
			setupToken: func() string {
				// Create a manager with expired duration
				expiredManager := NewJWTManager(secret, -1*time.Hour)
				token, _ := expiredManager.GenerateToken("user-123", "test@example.com", "Test User", "user")
				return token
			},
			wantErr:    true,
			wantErrMsg: "token has expired",
		},
		{
			name: "invalid signature",
			setupToken: func() string {
				wrongManager := NewJWTManager("wrong-secret", 1*time.Hour)
				token, _ := wrongManager.GenerateToken("user-123", "test@example.com", "Test User", "user")
				return token
			},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "this.is.not.a.valid.jwt"
			},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
		{
			name: "empty token",
			setupToken: func() string {
				return ""
			},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := manager.ValidateToken(token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.wantErrMsg != "" {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ValidateToken() error message = %v, want to contain %v", err.Error(), tt.wantErrMsg)
				}
			}

			if !tt.wantErr && claims == nil {
				t.Error("ValidateToken() returned nil claims for valid token")
			}
		})
	}
}

func TestJWTManager_ValidateToken_Claims(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 1*time.Hour)

	userID := "user-123"
	email := "test@example.com"
	name := "Test User"
	role := "admin"

	token, err := manager.GenerateToken(userID, email, name, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}
	if claims.Name != name {
		t.Errorf("Name = %v, want %v", claims.Name, name)
	}
	if claims.Role != role {
		t.Errorf("Role = %v, want %v", claims.Role, role)
	}
}

func TestJWTManager_TokenLifecycle(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 2*time.Second)

	// Generate token
	token, err := manager.GenerateToken("user-123", "test@example.com", "Test User", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Should be valid immediately
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Errorf("Token should be valid immediately after creation: %v", err)
	}
	if claims == nil {
		t.Error("Claims should not be nil for valid token")
	}

	// Wait for token to expire
	time.Sleep(3 * time.Second)

	// Should be invalid after expiration
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Token should be invalid after expiration")
	}
	if err != nil && err.Error() != "token has expired" {
		t.Errorf("Expected 'token has expired' error, got: %v", err)
	}
}

func TestJWTManager_DifferentSecrets(t *testing.T) {
	manager1 := NewJWTManager("secret1", 1*time.Hour)
	manager2 := NewJWTManager("secret2", 1*time.Hour)

	// Generate token with manager1
	token, err := manager1.GenerateToken("user-123", "test@example.com", "Test User", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate with manager1 (should succeed)
	_, err = manager1.ValidateToken(token)
	if err != nil {
		t.Errorf("Token should be valid with same secret: %v", err)
	}

	// Validate with manager2 (should fail)
	_, err = manager2.ValidateToken(token)
	if err == nil {
		t.Error("Token should be invalid with different secret")
	}
}
