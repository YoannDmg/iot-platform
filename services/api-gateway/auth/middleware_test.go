package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware_NoAuthHeader(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", 1*time.Hour)
	middleware := Middleware(jwtManager)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that user is not in context
		_, ok := GetUserFromContext(r.Context())
		if ok {
			t.Error("User should not be in context when no auth header provided")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", 1*time.Hour)
	middleware := Middleware(jwtManager)

	// Generate a valid token
	userID := "user-123"
	email := "test@example.com"
	name := "Test User"
	role := "user"
	token, err := jwtManager.GenerateToken(userID, email, name, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that user is in context
		claims, ok := GetUserFromContext(r.Context())
		if !ok {
			t.Error("User should be in context with valid token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Verify claims
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

		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", 1*time.Hour)
	middleware := Middleware(jwtManager)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid token")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %d", rr.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", -1*time.Hour) // Expired token
	middleware := Middleware(NewJWTManager("test-secret", 1*time.Hour))

	// Generate an expired token
	token, err := jwtManager.GenerateToken("user-123", "test@example.com", "Test User", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with expired token")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidBearerFormat(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", 1*time.Hour)
	middleware := Middleware(jwtManager)

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing Bearer prefix",
			header: "some-token",
		},
		{
			name:   "lowercase bearer",
			header: "bearer some-token",
		},
		{
			name:   "Basic auth",
			header: "Basic dXNlcjpwYXNz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called with invalid bearer format")
				w.WriteHeader(http.StatusOK)
			})

			wrapped := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Expected status Unauthorized, got %d", rr.Code)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setupCtx  func() context.Context
		wantOK    bool
		wantUser  *Claims
	}{
		{
			name: "user in context",
			setupCtx: func() context.Context {
				claims := &Claims{
					UserID: "user-123",
					Email:  "test@example.com",
					Name:   "Test User",
					Role:   "user",
				}
				return context.WithValue(context.Background(), UserContextKey, claims)
			},
			wantOK: true,
			wantUser: &Claims{
				UserID: "user-123",
				Email:  "test@example.com",
				Name:   "Test User",
				Role:   "user",
			},
		},
		{
			name: "no user in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantOK:   false,
			wantUser: nil,
		},
		{
			name: "wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserContextKey, "not-a-claims-object")
			},
			wantOK:   false,
			wantUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			user, ok := GetUserFromContext(ctx)

			if ok != tt.wantOK {
				t.Errorf("GetUserFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			if tt.wantOK && user == nil {
				t.Error("GetUserFromContext() returned nil user when expected")
			}

			if tt.wantOK && user != nil {
				if user.UserID != tt.wantUser.UserID {
					t.Errorf("UserID = %v, want %v", user.UserID, tt.wantUser.UserID)
				}
				if user.Email != tt.wantUser.Email {
					t.Errorf("Email = %v, want %v", user.Email, tt.wantUser.Email)
				}
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		wantErr  bool
	}{
		{
			name: "authenticated user",
			setupCtx: func() context.Context {
				claims := &Claims{
					UserID: "user-123",
					Email:  "test@example.com",
					Name:   "Test User",
					Role:   "user",
				}
				return context.WithValue(context.Background(), UserContextKey, claims)
			},
			wantErr: false,
		},
		{
			name: "no user in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			user, err := RequireAuth(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequireAuth() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && user == nil {
				t.Error("RequireAuth() returned nil user when authenticated")
			}

			if tt.wantErr && err != ErrUnauthorized {
				t.Errorf("RequireAuth() error = %v, want ErrUnauthorized", err)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name         string
		setupCtx     func() context.Context
		requiredRole string
		wantErr      bool
		wantErrType  error
	}{
		{
			name: "user has required role",
			setupCtx: func() context.Context {
				claims := &Claims{
					UserID: "user-123",
					Email:  "test@example.com",
					Name:   "Test User",
					Role:   "user",
				}
				return context.WithValue(context.Background(), UserContextKey, claims)
			},
			requiredRole: "user",
			wantErr:      false,
		},
		{
			name: "admin can access any role",
			setupCtx: func() context.Context {
				claims := &Claims{
					UserID: "admin-123",
					Email:  "admin@example.com",
					Name:   "Admin User",
					Role:   "admin",
				}
				return context.WithValue(context.Background(), UserContextKey, claims)
			},
			requiredRole: "user",
			wantErr:      false,
		},
		{
			name: "user cannot access admin role",
			setupCtx: func() context.Context {
				claims := &Claims{
					UserID: "user-123",
					Email:  "test@example.com",
					Name:   "Test User",
					Role:   "user",
				}
				return context.WithValue(context.Background(), UserContextKey, claims)
			},
			requiredRole: "admin",
			wantErr:      true,
			wantErrType:  ErrForbidden,
		},
		{
			name: "no user in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			requiredRole: "user",
			wantErr:      true,
			wantErrType:  ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			user, err := RequireRole(ctx, tt.requiredRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequireRole() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && user == nil {
				t.Error("RequireRole() returned nil user when authorized")
			}

			if tt.wantErr && err != tt.wantErrType {
				t.Errorf("RequireRole() error = %v, want %v", err, tt.wantErrType)
			}
		})
	}
}
