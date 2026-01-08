// Package main is the API Gateway entry point.
// Exposes a public GraphQL API and communicates with microservices via gRPC.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

// main configures and starts the HTTP GraphQL server.
//
// Endpoints:
//   - /health  : Health check endpoint
//   - /        : GraphQL Playground (dev only)
//   - /query   : GraphQL API endpoint
//
// Configuration:
//   - PORT: Server port (default: 8080)
//
// TODO Production:
//   - Implement GraphQL resolvers
//   - Connect to Device Manager via gRPC
//   - Disable Playground in production
//   - Add JWT authentication
//   - Implement rate limiting
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// TODO: Enable once resolvers are implemented
	// srv := handler.NewDefaultServer(
	// 	generated.NewExecutableSchema(
	// 		generated.Config{Resolvers: &graph.Resolver{}},
	// 	),
	// )

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})

	// GraphQL Playground - disable in production
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// TODO: GraphQL endpoint
	// http.Handle("/query", srv)

	log.Printf("🚀 API Gateway started on port %s", port)
	log.Printf("📊 GraphQL Playground: http://localhost:%s/", port)
	log.Printf("🔗 GraphQL API: http://localhost:%s/query", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
