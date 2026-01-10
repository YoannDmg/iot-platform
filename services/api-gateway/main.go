// Package main is the API Gateway entry point.
// Exposes a public GraphQL API and communicates with microservices via gRPC.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/yourusername/iot-platform/services/api-gateway/auth"
	"github.com/yourusername/iot-platform/services/api-gateway/graph"
	"github.com/yourusername/iot-platform/services/api-gateway/graph/generated"
	grpcClient "github.com/yourusername/iot-platform/services/api-gateway/grpc"
)

const (
	defaultPort              = "8080"
	defaultDeviceManagerAddr = "localhost:8081"
	defaultUserServiceAddr   = "localhost:8083"
	defaultJWTSecret         = "dev-jwt-secret-NOT-FOR-PRODUCTION"
)

// main configures and starts the HTTP GraphQL server.
//
// Endpoints:
//   - /health  : Health check endpoint
//   - /        : GraphQL Playground (dev only)
//   - /query   : GraphQL API endpoint
//
// Configuration:
//   - PORT: Server port (default: 8080)
//   - DEVICE_MANAGER_ADDR: Device Manager address (default: localhost:8081)
//   - USER_SERVICE_ADDR: User Service address (default: localhost:8083)
//   - JWT_SECRET: Secret key for JWT tokens (default: dev-jwt-secret-NOT-FOR-PRODUCTION)
//
// TODO Production:
//   - Disable Playground in production
//   - Implement rate limiting
//   - Add TLS support
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	deviceManagerAddr := os.Getenv("DEVICE_MANAGER_ADDR")
	if deviceManagerAddr == "" {
		deviceManagerAddr = defaultDeviceManagerAddr
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = defaultUserServiceAddr
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = defaultJWTSecret
		log.Printf("⚠️  Using default JWT secret (dev only)")
	}

	// Connect to Device Manager via gRPC
	deviceClient, err := grpcClient.NewDeviceClient(deviceManagerAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Device Manager: %v", err)
	}
	defer func() {
		if err := deviceClient.Close(); err != nil {
			log.Printf("⚠️  Failed to close device client: %v", err)
		}
	}()

	// Connect to User Service via gRPC
	userClient, err := grpcClient.NewUserClient(userServiceAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to User Service: %v", err)
	}
	defer func() {
		if err := userClient.Close(); err != nil {
			log.Printf("⚠️  Failed to close user client: %v", err)
		}
	}()

	// Initialize JWT manager (24 hours token duration)
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// Create GraphQL server with auth middleware
	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: &graph.Resolver{
					DeviceClient: deviceClient.GetClient(),
					UserClient:   userClient.GetClient(),
					JWTManager:   jwtManager,
				},
			},
		),
	)

	// Wrap GraphQL handler with JWT middleware
	authMiddleware := auth.Middleware(jwtManager)
	graphqlHandler := authMiddleware(srv)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"api-gateway"}`)); err != nil {
			log.Printf("⚠️  Failed to write health check response: %v", err)
		}
	})

	// GraphQL Playground - disable in production
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// GraphQL API endpoint with auth middleware
	http.Handle("/query", graphqlHandler)

	log.Println("=====================================")
	log.Printf("API Gateway Service")
	log.Println("=====================================")
	log.Printf("Protocol: GraphQL (HTTP)")
	log.Printf("Port: %s", port)
	log.Printf("Device Manager: %s", deviceManagerAddr)
	log.Printf("User Service: %s", userServiceAddr)
	log.Println("-------------------------------------")
	log.Printf("📊 GraphQL Playground: http://localhost:%s/", port)
	log.Printf("🔗 GraphQL API: http://localhost:%s/query", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("=====================================")
	log.Printf("✅ Server started")
	log.Println("=====================================")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

