// Package auth provides HTTP middleware for JWT authentication
package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key used to store user claims in context
	UserContextKey contextKey = "user"
)

// Middleware creates an HTTP middleware that validates JWT tokens
func Middleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")

			// If no auth header, continue without user context
			// (some queries/mutations may be public like login/register)
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract Bearer token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				// Authorization header present but not in Bearer format
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	user, ok := ctx.Value(UserContextKey).(*Claims)
	return user, ok
}

// WithUser adds user claims to context (used for WebSocket auth)
func WithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

// RequireAuth is a helper to check if user is authenticated in resolvers
func RequireAuth(ctx context.Context) (*Claims, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	return user, nil
}

// RequireRole checks if user has the required role
func RequireRole(ctx context.Context, requiredRole string) (*Claims, error) {
	user, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if user.Role != requiredRole && user.Role != "admin" {
		return nil, ErrForbidden
	}

	return user, nil
}

var (
	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = &AuthError{Message: "authentication required", Code: "UNAUTHENTICATED"}
	// ErrForbidden is returned when user lacks required permissions
	ErrForbidden = &AuthError{Message: "insufficient permissions", Code: "FORBIDDEN"}
)

// AuthError represents an authentication/authorization error
type AuthError struct {
	Message string
	Code    string
}

func (e *AuthError) Error() string {
	return e.Message
}
